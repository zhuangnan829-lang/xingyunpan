package queue

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

const (
	fullTextIndexUID      = "xingyunpan_fulltext"
	fullTextBatchSize     = 32
	fullTextTaskTimeout   = 2 * time.Minute
	fullTextTaskPollDelay = 1200 * time.Millisecond
)

type meiliTaskEnvelope struct {
	TaskUID int64 `json:"taskUid"`
	UID     int64 `json:"uid"`
}

type meiliTaskStatus struct {
	TaskUID int64  `json:"taskUid"`
	UID     int64  `json:"uid"`
	Status  string `json:"status"`
	Error   *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
		Type    string `json:"type"`
		Link    string `json:"link"`
	} `json:"error"`
}

type fullTextDocument struct {
	ID             string `json:"id"`
	FileID         uint   `json:"file_id"`
	PhysicalFileID uint   `json:"physical_file_id"`
	UserID         uint   `json:"user_id"`
	FileName       string `json:"file_name"`
	FilePath       string `json:"file_path"`
	Extension      string `json:"extension"`
	Content        string `json:"content"`
	ChunkIndex     int    `json:"chunk_index"`
	ChunkTotal     int    `json:"chunk_total"`
	FileSize       int64  `json:"file_size"`
	UpdatedAt      string `json:"updated_at"`
}

type fullTextRebuildMetrics struct {
	totalFiles          int64
	candidateFiles      int64
	indexedFiles        int64
	indexedChunks       int64
	skippedFiles        int64
	failedFiles         int64
	lastProcessedFileID uint
}

func (e *Executor) rebuildFullTextIndex(ctx context.Context, payload FullTextRebuildPayload, setting model.FullTextSearchSetting) (string, error) {
	maxFileSizeBytes, err := sizeToBytes(setting.MaxFileSize, setting.MaxFileSizeUnit)
	if err != nil {
		return "", fmt.Errorf("invalid max file size configuration: %w", err)
	}

	chunkSizeBytes, err := sizeToBytes(setting.ChunkSize, setting.ChunkUnit)
	if err != nil {
		return "", fmt.Errorf("invalid chunk size configuration: %w", err)
	}
	if chunkSizeBytes <= 0 {
		chunkSizeBytes = 2000
	}

	extensions := normalizeExtensions(setting.Extensions)
	metrics, err := e.prepareFullTextMetrics(ctx, extensions)
	if err != nil {
		return "", err
	}

	meiliEndpoint := normalizeMeiliEndpoint(setting.MeiliEndpoint)
	tikaEndpoint := normalizeTikaEndpoint(setting.TikaEndpoint)
	client := &http.Client{Timeout: 90 * time.Second}

	if err := resetMeiliIndex(ctx, client, meiliEndpoint, setting.APIKey); err != nil {
		return "", err
	}

	var lastID uint
	pendingDocs := make([]fullTextDocument, 0, fullTextBatchSize)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		files, err := e.loadFullTextCandidates(ctx, extensions, lastID, fullTextBatchSize)
		if err != nil {
			return "", err
		}
		if len(files) == 0 {
			break
		}

		for _, file := range files {
			lastID = file.ID
			metrics.lastProcessedFileID = file.ID

			chunks, fileErr := e.extractSearchChunks(ctx, client, tikaEndpoint, maxFileSizeBytes, chunkSizeBytes, &file)
			if fileErr != nil {
				metrics.failedFiles++
				return "", fmt.Errorf("rebuild index failed on file #%d (%s): %w", file.ID, file.FileName, fileErr)
			}
			if len(chunks) == 0 {
				metrics.skippedFiles++
				continue
			}

			metrics.indexedFiles++
			metrics.indexedChunks += int64(len(chunks))
			for _, doc := range chunks {
				pendingDocs = append(pendingDocs, doc)
				if len(pendingDocs) >= fullTextBatchSize {
					if err := uploadDocumentsToMeili(ctx, client, meiliEndpoint, setting.APIKey, pendingDocs); err != nil {
						return "", err
					}
					pendingDocs = pendingDocs[:0]
				}
			}
		}
	}

	if len(pendingDocs) > 0 {
		if err := uploadDocumentsToMeili(ctx, client, meiliEndpoint, setting.APIKey, pendingDocs); err != nil {
			return "", err
		}
	}

	documentCount, err := fetchMeiliDocumentCount(ctx, client, meiliEndpoint, setting.APIKey)
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{
		"status":              "completed",
		"triggered_by":        payload.TriggeredBy,
		"engine":              "meilisearch",
		"extractor":           "tika",
		"indexed_candidates":  metrics.candidateFiles,
		"total_files":         metrics.totalFiles,
		"indexed_files":       metrics.indexedFiles,
		"indexed_chunks":      metrics.indexedChunks,
		"indexed_documents":   documentCount,
		"skipped_files":       metrics.skippedFiles,
		"failed_files":        metrics.failedFiles,
		"extensions":          extensions,
		"meili_endpoint":      meiliEndpoint,
		"tika_endpoint":       tikaEndpoint,
		"index_uid":           fullTextIndexUID,
		"last_processed_file": metrics.lastProcessedFileID,
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal full text rebuild result failed: %w", err)
	}

	return string(bytes), nil
}

func (e *Executor) prepareFullTextMetrics(ctx context.Context, extensions []string) (*fullTextRebuildMetrics, error) {
	metrics := &fullTextRebuildMetrics{}

	if err := e.db.WithContext(ctx).
		Model(&model.UserFile{}).
		Where("deleted_at IS NULL").
		Where("is_folder = ?", false).
		Count(&metrics.totalFiles).Error; err != nil {
		return nil, fmt.Errorf("count total files for full text rebuild failed: %w", err)
	}

	query := e.db.WithContext(ctx).
		Model(&model.UserFile{}).
		Where("deleted_at IS NULL").
		Where("is_folder = ?", false).
		Where("physical_file_id IS NOT NULL")
	query = withFullTextExtensions(query, extensions)
	if err := query.Count(&metrics.candidateFiles).Error; err != nil {
		return nil, fmt.Errorf("count full text candidates failed: %w", err)
	}

	return metrics, nil
}

func (e *Executor) loadFullTextCandidates(ctx context.Context, extensions []string, lastID uint, limit int) ([]model.UserFile, error) {
	files := make([]model.UserFile, 0, limit)

	query := e.db.WithContext(ctx).
		Preload("PhysicalFile").
		Where("deleted_at IS NULL").
		Where("is_folder = ?", false).
		Where("physical_file_id IS NOT NULL").
		Where("id > ?", lastID)
	query = withFullTextExtensions(query, extensions)

	if err := query.Order("id asc").Limit(limit).Find(&files).Error; err != nil {
		return nil, fmt.Errorf("list full text candidates failed: %w", err)
	}

	return files, nil
}

func withFullTextExtensions(query *gorm.DB, extensions []string) *gorm.DB {
	if len(extensions) == 0 {
		return query
	}

	clauses := make([]string, 0, len(extensions))
	args := make([]interface{}, 0, len(extensions))
	for _, ext := range extensions {
		clauses = append(clauses, "LOWER(file_name) LIKE ?")
		args = append(args, "%."+ext)
	}

	return query.Where("("+strings.Join(clauses, " OR ")+")", args...)
}

func (e *Executor) extractSearchChunks(
	ctx context.Context,
	client *http.Client,
	tikaEndpoint string,
	maxFileSizeBytes int64,
	chunkSizeBytes int64,
	file *model.UserFile,
) ([]fullTextDocument, error) {
	if file == nil || file.PhysicalFile == nil {
		return nil, nil
	}
	if file.PhysicalFile.StoragePath == "" {
		return nil, nil
	}
	if maxFileSizeBytes > 0 && file.FileSize > maxFileSizeBytes {
		return nil, nil
	}
	if !e.storage.Exists(file.PhysicalFile.StoragePath) {
		return nil, nil
	}

	reader, err := e.storage.Read(file.PhysicalFile.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("read source file failed: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read source file bytes failed: %w", err)
	}

	text, err := extractSearchText(ctx, client, tikaEndpoint, file.FileName, file.PhysicalFile.ContentType, data)
	if err != nil {
		return nil, err
	}

	text = normalizeSearchText(text)
	if text == "" {
		return nil, nil
	}

	chunks := splitTextByBytes(text, chunkSizeBytes)
	if len(chunks) == 0 {
		return nil, nil
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.FileName)), ".")
	docs := make([]fullTextDocument, 0, len(chunks))
	for idx, chunk := range chunks {
		docs = append(docs, fullTextDocument{
			ID:             fmt.Sprintf("%d-%d", file.ID, idx),
			FileID:         file.ID,
			PhysicalFileID: file.PhysicalFile.ID,
			UserID:         file.UserID,
			FileName:       file.FileName,
			FilePath:       file.FilePath,
			Extension:      ext,
			Content:        chunk,
			ChunkIndex:     idx,
			ChunkTotal:     len(chunks),
			FileSize:       file.FileSize,
			UpdatedAt:      file.UpdatedAt.Format(time.RFC3339),
		})
	}

	return docs, nil
}

func extractSearchText(
	ctx context.Context,
	client *http.Client,
	tikaEndpoint string,
	fileName string,
	contentType string,
	data []byte,
) (string, error) {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")

	switch ext {
	case "txt", "md", "csv", "html", "htm":
		return decodeTextBytes(data), nil
	case "docx":
		if text := extractDocxText(data); text != "" {
			return text, nil
		}
	}

	return extractPlainTextWithTika(ctx, client, tikaEndpoint, fileName, contentType, bytes.NewReader(data))
}

func extractPlainTextWithTika(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	fileName string,
	contentType string,
	reader io.Reader,
) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, reader)
	if err != nil {
		return "", fmt.Errorf("create tika request failed: %w", err)
	}

	req.Header.Set("Accept", "text/plain")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	if fileName != "" {
		req.Header.Set("X-Tika-Filename", fileName)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request tika failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read tika response failed: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("tika returned %d: %s", resp.StatusCode, strings.TrimSpace(truncateText(string(body), 320)))
	}

	return string(body), nil
}

func normalizeSearchText(text string) string {
	text = repairUTF8Mojibake(text)
	text = strings.ReplaceAll(text, "\u0000", "")
	text = strings.TrimPrefix(text, "\uFEFF")
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	return strings.Join(strings.Fields(text), " ")
}

func decodeTextBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	switch {
	case bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}):
		return string(data[3:])
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}):
		return decodeUTF16Bytes(data[2:], binary.LittleEndian)
	case bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		return decodeUTF16Bytes(data[2:], binary.BigEndian)
	default:
		return string(data)
	}
}

func decodeUTF16Bytes(data []byte, order binary.ByteOrder) string {
	if len(data) < 2 {
		return ""
	}

	units := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		units = append(units, order.Uint16(data[i:i+2]))
	}

	runes := utf16.Decode(units)
	return string(runes)
}

func repairUTF8Mojibake(text string) string {
	if text == "" {
		return ""
	}

	raw := make([]byte, 0, len(text))
	for _, r := range text {
		if r <= 255 {
			raw = append(raw, byte(r))
			continue
		}
		raw = utf8.AppendRune(raw, r)
	}
	if !utf8.Valid(raw) {
		return text
	}

	repaired := string(raw)
	if countHanRunes(repaired) > countHanRunes(text) {
		return repaired
	}

	return text
}

func extractDocxText(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ""
	}

	var builder strings.Builder
	for _, file := range reader.File {
		if file.Name != "word/document.xml" {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			return ""
		}

		decoder := xml.NewDecoder(rc)
		for {
			token, tokenErr := decoder.Token()
			if tokenErr == io.EOF {
				break
			}
			if tokenErr != nil {
				rc.Close()
				return ""
			}

			switch typed := token.(type) {
			case xml.CharData:
				builder.WriteString(string(typed))
				builder.WriteByte(' ')
			}
		}

		rc.Close()
		break
	}

	return builder.String()
}

func countHanRunes(text string) int {
	total := 0
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			total++
		}
	}
	return total
}

func splitTextByBytes(text string, maxBytes int64) []string {
	if text == "" {
		return nil
	}
	if maxBytes <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	maxChunkBytes := int(maxBytes)
	chunks := make([]string, 0, len(words)/8+1)
	var builder strings.Builder
	currentBytes := 0

	flush := func() {
		if builder.Len() == 0 {
			return
		}
		chunks = append(chunks, strings.TrimSpace(builder.String()))
		builder.Reset()
		currentBytes = 0
	}

	for _, word := range words {
		wordBytes := len(word)
		if wordBytes > maxChunkBytes {
			flush()
			chunks = append(chunks, splitLongWord(word, maxChunkBytes)...)
			continue
		}

		extra := wordBytes
		if currentBytes > 0 {
			extra++
		}
		if currentBytes+extra > maxChunkBytes {
			flush()
		}
		if builder.Len() > 0 {
			builder.WriteByte(' ')
			currentBytes++
		}
		builder.WriteString(word)
		currentBytes += wordBytes
	}
	flush()

	return chunks
}

func splitLongWord(word string, maxBytes int) []string {
	if maxBytes <= 0 {
		return []string{word}
	}

	chunks := make([]string, 0, len(word)/maxBytes+1)
	var builder strings.Builder
	currentBytes := 0
	for _, r := range word {
		runeBytes := utf8.RuneLen(r)
		if runeBytes < 0 {
			runeBytes = 1
		}
		if currentBytes+runeBytes > maxBytes && builder.Len() > 0 {
			chunks = append(chunks, builder.String())
			builder.Reset()
			currentBytes = 0
		}
		builder.WriteRune(r)
		currentBytes += runeBytes
	}
	if builder.Len() > 0 {
		chunks = append(chunks, builder.String())
	}
	return chunks
}

func resetMeiliIndex(ctx context.Context, client *http.Client, endpoint string, apiKey string) error {
	taskUID, err := deleteMeiliIndex(ctx, client, endpoint, apiKey, fullTextIndexUID)
	if err != nil {
		return err
	}
	if taskUID > 0 {
		if err := waitForMeiliTask(ctx, client, endpoint, apiKey, taskUID); err != nil && !isMeiliIndexNotFoundError(err) {
			return err
		}
	}

	taskUID, err = createMeiliIndex(ctx, client, endpoint, apiKey, fullTextIndexUID)
	if err != nil {
		return err
	}
	if taskUID > 0 {
		if err := waitForMeiliTask(ctx, client, endpoint, apiKey, taskUID); err != nil {
			return err
		}
	}

	return nil
}

func isMeiliIndexNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "index `"+fullTextIndexUID+"` not found") ||
		strings.Contains(message, "index not found")
}

func deleteMeiliIndex(ctx context.Context, client *http.Client, endpoint string, apiKey string, indexUID string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint+"/indexes/"+url.PathEscape(indexUID), nil)
	if err != nil {
		return 0, fmt.Errorf("create meilisearch delete request failed: %w", err)
	}
	applyMeiliHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("delete meilisearch index failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read meilisearch delete response failed: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return 0, fmt.Errorf("delete meilisearch index returned %d: %s", resp.StatusCode, strings.TrimSpace(truncateText(string(body), 320)))
	}

	return parseMeiliTaskUID(body)
}

func createMeiliIndex(ctx context.Context, client *http.Client, endpoint string, apiKey string, indexUID string) (int64, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"uid":        indexUID,
		"primaryKey": "id",
	})
	if err != nil {
		return 0, fmt.Errorf("marshal meilisearch create index payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/indexes", bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("create meilisearch create request failed: %w", err)
	}
	applyMeiliHeaders(req, apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("create meilisearch index failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read meilisearch create response failed: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return 0, fmt.Errorf("create meilisearch index returned %d: %s", resp.StatusCode, strings.TrimSpace(truncateText(string(body), 320)))
	}

	return parseMeiliTaskUID(body)
}

func uploadDocumentsToMeili(ctx context.Context, client *http.Client, endpoint string, apiKey string, docs []fullTextDocument) error {
	payload, err := json.Marshal(docs)
	if err != nil {
		return fmt.Errorf("marshal meilisearch documents failed: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint+"/indexes/"+url.PathEscape(fullTextIndexUID)+"/documents?primaryKey=id",
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("create meilisearch documents request failed: %w", err)
	}
	applyMeiliHeaders(req, apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload meilisearch documents failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read meilisearch documents response failed: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("upload meilisearch documents returned %d: %s", resp.StatusCode, strings.TrimSpace(truncateText(string(body), 320)))
	}

	taskUID, err := parseMeiliTaskUID(body)
	if err != nil {
		return err
	}

	return waitForMeiliTask(ctx, client, endpoint, apiKey, taskUID)
}

func fetchMeiliDocumentCount(ctx context.Context, client *http.Client, endpoint string, apiKey string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/indexes/"+url.PathEscape(fullTextIndexUID)+"/stats", nil)
	if err != nil {
		return 0, fmt.Errorf("create meilisearch stats request failed: %w", err)
	}
	applyMeiliHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request meilisearch stats failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read meilisearch stats response failed: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return 0, fmt.Errorf("meilisearch stats returned %d: %s", resp.StatusCode, strings.TrimSpace(truncateText(string(body), 320)))
	}

	var payload struct {
		NumberOfDocuments int64 `json:"numberOfDocuments"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, fmt.Errorf("unmarshal meilisearch stats failed: %w", err)
	}

	return payload.NumberOfDocuments, nil
}

func waitForMeiliTask(ctx context.Context, client *http.Client, endpoint string, apiKey string, taskUID int64) error {
	if taskUID <= 0 {
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, fullTextTaskTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("meilisearch task %d timed out", taskUID)
		default:
		}

		req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, fmt.Sprintf("%s/tasks/%d", endpoint, taskUID), nil)
		if err != nil {
			return fmt.Errorf("create meilisearch task request failed: %w", err)
		}
		applyMeiliHeaders(req, apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("query meilisearch task failed: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read meilisearch task response failed: %w", readErr)
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("query meilisearch task returned %d: %s", resp.StatusCode, strings.TrimSpace(truncateText(string(body), 320)))
		}

		var task meiliTaskStatus
		if err := json.Unmarshal(body, &task); err != nil {
			return fmt.Errorf("decode meilisearch task response failed: %w", err)
		}

		switch strings.ToLower(task.Status) {
		case "succeeded":
			return nil
		case "failed":
			if task.Error != nil && task.Error.Message != "" {
				return fmt.Errorf("meilisearch task %d failed: %s", taskUID, task.Error.Message)
			}
			return fmt.Errorf("meilisearch task %d failed", taskUID)
		}

		time.Sleep(fullTextTaskPollDelay)
	}
}

func parseMeiliTaskUID(body []byte) (int64, error) {
	var task meiliTaskEnvelope
	if err := json.Unmarshal(body, &task); err != nil {
		return 0, fmt.Errorf("decode meilisearch task response failed: %w", err)
	}
	if task.TaskUID > 0 {
		return task.TaskUID, nil
	}
	if task.UID > 0 {
		return task.UID, nil
	}
	return 0, fmt.Errorf("meilisearch task uid missing")
}

func applyMeiliHeaders(req *http.Request, apiKey string) {
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Accept", "application/json")
}

func normalizeTikaEndpoint(endpoint string) string {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return ""
	}
	if strings.HasSuffix(strings.ToLower(endpoint), "/tika") {
		return endpoint
	}
	return endpoint + "/tika"
}

func normalizeMeiliEndpoint(endpoint string) string {
	return strings.TrimRight(strings.TrimSpace(endpoint), "/")
}

func sizeToBytes(value int, unit string) (int64, error) {
	if value <= 0 {
		return 0, nil
	}

	multiplier := int64(1)
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "B", "":
		multiplier = 1
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unsupported unit %q", unit)
	}

	return int64(value) * multiplier, nil
}

func truncateText(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}
