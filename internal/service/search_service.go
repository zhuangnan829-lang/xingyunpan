package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// SearchService provides file search and suggestion capabilities.
type SearchService interface {
	SearchFiles(ctx context.Context, userID uint, params *SearchParams) (*SearchResult, error)
	GetSuggestions(ctx context.Context, userID uint, prefix string) ([]*SearchSuggestion, error)
}

// SearchParams holds filters for searching files.
type SearchParams struct {
	Keyword        string     `json:"keyword"`
	FileType       string     `json:"file_type"`
	SizeMin        *int64     `json:"size_min"`
	SizeMax        *int64     `json:"size_max"`
	DateFrom       *time.Time `json:"date_from"`
	DateTo         *time.Time `json:"date_to"`
	FolderID       *string    `json:"folder_id"`
	Page           int        `json:"page"`
	PageSize       int        `json:"page_size"`
	Cursor         string     `json:"cursor"`
	MaxResults     int        `json:"-"`
	MaxPageSize    int        `json:"-"`
	PaginationMode string     `json:"-"`
	UseCursor      bool       `json:"-"`
	CursorID       uint       `json:"-"`
}

// SearchResult is the API response payload for search.
type SearchResult struct {
	Files          []*FileItem `json:"files"`
	Total          int64       `json:"total"`
	Page           int         `json:"page"`
	PageSize       int         `json:"page_size"`
	PaginationMode string      `json:"pagination_mode"`
	NextCursor     string      `json:"next_cursor"`
	MaxPageSize    int         `json:"max_page_size"`
}

// FileItem is a simplified file representation for search responses.
type FileItem struct {
	FileID     string    `json:"file_id"`
	FileName   string    `json:"file_name"`
	FileType   string    `json:"file_type"`
	FileSize   int64     `json:"file_size"`
	FilePath   string    `json:"file_path"`
	ModifiedAt time.Time `json:"modified_at"`
	IsFolder   bool      `json:"is_folder"`
}

// SearchSuggestion describes a keyword suggestion.
type SearchSuggestion struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
}

// CacheInterface abstracts the cache dependency for testing.
type CacheInterface interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

type searchService struct {
	db               *gorm.DB
	cache            CacheInterface
	settingsRepo     repository.FileSystemSettingRepository
	fullTextSettings repository.FullTextSearchSettingRepository
	httpClient       *http.Client
}

// NewSearchService creates a search service instance.
func NewSearchService(
	db *gorm.DB,
	cache CacheInterface,
	settingsRepo repository.FileSystemSettingRepository,
	fullTextSettings repository.FullTextSearchSettingRepository,
) SearchService {
	return &searchService{
		db:               db,
		cache:            cache,
		settingsRepo:     settingsRepo,
		fullTextSettings: fullTextSettings,
		httpClient:       &http.Client{Timeout: 15 * time.Second},
	}
}

// SearchFiles searches files using dynamic filters and applies admin limits.
func (s *searchService) SearchFiles(ctx context.Context, userID uint, params *SearchParams) (*SearchResult, error) {
	params = s.normalizeParams(params)

	cacheKey := s.generateCacheKey(userID, params)
	useCache := s.shouldCacheSearch(params)

	var cachedResult SearchResult
	if useCache {
		if err := s.getFromCache(ctx, cacheKey, &cachedResult); err == nil {
			return &cachedResult, nil
		}
	}

	if s.shouldUseFullTextSearch(params) {
		if result, err := s.searchFilesWithFullText(ctx, userID, params); err == nil && result != nil {
			if useCache {
				if err := s.saveToCache(ctx, cacheKey, result, 5*time.Minute); err != nil {
					// Cache failures should not break search.
				}
			}
			return result, nil
		}
	}

	result, err := s.searchFilesWithDatabase(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	if useCache {
		if err := s.saveToCache(ctx, cacheKey, result, 5*time.Minute); err != nil {
			// Cache failures should not break search.
		}
	}

	return result, nil
}

func (s *searchService) shouldCacheSearch(params *SearchParams) bool {
	if params == nil || strings.TrimSpace(params.Keyword) == "" {
		return false
	}
	if strings.TrimSpace(params.FileType) != "" ||
		params.SizeMin != nil ||
		params.SizeMax != nil ||
		params.DateFrom != nil ||
		params.DateTo != nil ||
		params.FolderID != nil {
		return false
	}
	return true
}

func (s *searchService) searchFilesWithDatabase(ctx context.Context, userID uint, params *SearchParams) (*SearchResult, error) {
	query := s.db.Model(&model.UserFile{}).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL")

	if params.Keyword != "" {
		query = query.Where("(LOWER(file_name) LIKE LOWER(?) OR LOWER(file_path) LIKE LOWER(?))", "%"+params.Keyword+"%", "%"+params.Keyword+"%")
	}
	query = s.applyStructuredFilters(query, params)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("query total files failed: %w", err)
	}

	if total > int64(params.MaxResults) {
		total = int64(params.MaxResults)
	}

	offset := (params.Page - 1) * params.PageSize
	if params.UseCursor {
		offset = 0
		if params.CursorID > 0 {
			query = query.Where("id < ?", params.CursorID)
		}
	}
	if offset >= int(total) {
		return &SearchResult{
			Files:          []*FileItem{},
			Total:          total,
			Page:           params.Page,
			PageSize:       params.PageSize,
			PaginationMode: params.PaginationMode,
			MaxPageSize:    params.MaxPageSize,
		}, nil
	}

	limit := params.PageSize
	if remaining := int(total) - offset; remaining < limit {
		limit = remaining
	}

	var files []*model.UserFile
	if err := query.
		Preload("PhysicalFile").
		Order("updated_at DESC, id DESC").
		Offset(offset).
		Limit(limit).
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("query files failed: %w", err)
	}

	fileItems := make([]*FileItem, len(files))
	for i, file := range files {
		fileItems[i] = &FileItem{
			FileID:     fmt.Sprintf("%d", file.ID),
			FileName:   file.FileName,
			FileType:   s.getFileType(file),
			FileSize:   s.getFileSize(file),
			FilePath:   s.buildFilePath(file),
			ModifiedAt: file.UpdatedAt,
			IsFolder:   file.IsFolder,
		}
	}
	nextCursor := ""
	if params.UseCursor && len(files) > 0 {
		nextCursor = NextCursorFromUint(files[len(files)-1].ID)
	}

	return &SearchResult{
		Files:          fileItems,
		Total:          total,
		Page:           params.Page,
		PageSize:       params.PageSize,
		PaginationMode: params.PaginationMode,
		NextCursor:     nextCursor,
		MaxPageSize:    params.MaxPageSize,
	}, nil
}

// GetSuggestions returns file-name based suggestions.
func (s *searchService) GetSuggestions(ctx context.Context, userID uint, prefix string) ([]*SearchSuggestion, error) {
	if len(prefix) < 2 {
		return []*SearchSuggestion{}, nil
	}

	if suggestions, err := s.getSuggestionsWithFullText(ctx, userID, prefix); err == nil && len(suggestions) > 0 {
		return suggestions, nil
	}

	return s.getSuggestionsWithDatabase(ctx, userID, prefix)
}

func (s *searchService) getSuggestionsWithDatabase(ctx context.Context, userID uint, prefix string) ([]*SearchSuggestion, error) {
	var suggestions []*SearchSuggestion
	err := s.db.WithContext(ctx).Model(&model.UserFile{}).
		Select("file_name as keyword, COUNT(*) as count").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Where("file_name LIKE ?", prefix+"%").
		Group("file_name").
		Order("count DESC").
		Limit(10).
		Scan(&suggestions).Error
	if err != nil {
		return nil, fmt.Errorf("query search suggestions failed: %w", err)
	}

	return suggestions, nil
}

func (s *searchService) getSuggestionsWithFullText(ctx context.Context, userID uint, prefix string) ([]*SearchSuggestion, error) {
	if s.fullTextSettings == nil {
		return nil, fmt.Errorf("full text settings repository is not initialized")
	}

	setting, err := s.fullTextSettings.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil || !setting.Enabled {
		return nil, fmt.Errorf("full text search is not enabled")
	}

	endpoint := strings.TrimRight(strings.TrimSpace(setting.MeiliEndpoint), "/")
	if endpoint == "" {
		return nil, fmt.Errorf("meilisearch endpoint is empty")
	}

	responsePayload, err := s.performMeiliSearch(ctx, endpoint, setting.APIKey, meiliSearchRequest{
		Query:                strings.TrimSpace(prefix),
		Limit:                20,
		Filter:               []string{fmt.Sprintf("user_id = %d", userID)},
		AttributesToRetrieve: []string{"file_name"},
	})
	if err != nil {
		return nil, err
	}

	suggestions := make([]*SearchSuggestion, 0, 10)
	seen := make(map[string]struct{}, 10)
	prefixLower := strings.ToLower(strings.TrimSpace(prefix))
	for _, hit := range responsePayload.Hits {
		name := strings.TrimSpace(hit.FileName)
		if name == "" {
			continue
		}
		if !strings.Contains(strings.ToLower(name), prefixLower) {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}

		seen[name] = struct{}{}
		suggestions = append(suggestions, &SearchSuggestion{
			Keyword: name,
			Count:   0,
		})
		if len(suggestions) >= 10 {
			break
		}
	}

	return suggestions, nil
}

type meiliSearchRequest struct {
	Query                string   `json:"q"`
	Offset               int      `json:"offset,omitempty"`
	Limit                int      `json:"limit,omitempty"`
	Filter               []string `json:"filter,omitempty"`
	AttributesToRetrieve []string `json:"attributesToRetrieve,omitempty"`
}

type meiliSearchHit struct {
	FileID   interface{} `json:"file_id"`
	FileName string      `json:"file_name"`
}

type meiliSearchResponse struct {
	Hits               []meiliSearchHit `json:"hits"`
	EstimatedTotalHits int64            `json:"estimatedTotalHits"`
	TotalHits          int64            `json:"totalHits"`
}

func (s *searchService) shouldUseFullTextSearch(params *SearchParams) bool {
	if params == nil {
		return false
	}

	return strings.TrimSpace(params.Keyword) != ""
}

func (s *searchService) searchFilesWithFullText(ctx context.Context, userID uint, params *SearchParams) (*SearchResult, error) {
	if s.fullTextSettings == nil {
		return nil, fmt.Errorf("full text settings repository is not initialized")
	}

	setting, err := s.fullTextSettings.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil || !setting.Enabled {
		return nil, fmt.Errorf("full text search is not enabled")
	}

	endpoint := strings.TrimRight(strings.TrimSpace(setting.MeiliEndpoint), "/")
	if endpoint == "" {
		return nil, fmt.Errorf("meilisearch endpoint is empty")
	}

	limit := params.PageSize * 4
	if limit < 50 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	requestPayload := meiliSearchRequest{
		Query:                strings.TrimSpace(params.Keyword),
		Limit:                limit,
		Filter:               []string{fmt.Sprintf("user_id = %d", userID)},
		AttributesToRetrieve: []string{"file_id"},
	}

	responsePayload, err := s.performMeiliSearch(ctx, endpoint, setting.APIKey, requestPayload)
	if err != nil {
		return nil, err
	}

	orderedIDs := collectOrderedMeiliFileIDs(responsePayload.Hits)
	if len(orderedIDs) == 0 {
		return &SearchResult{
			Files:          []*FileItem{},
			Total:          0,
			Page:           params.Page,
			PageSize:       params.PageSize,
			PaginationMode: params.PaginationMode,
			MaxPageSize:    params.MaxPageSize,
		}, nil
	}

	query := s.db.WithContext(ctx).
		Model(&model.UserFile{}).
		Preload("PhysicalFile").
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Where("id IN ?", orderedIDs)
	query = s.applyStructuredFilters(query, params)

	var files []*model.UserFile
	if err := query.Find(&files).Error; err != nil {
		return nil, fmt.Errorf("query full text matched files failed: %w", err)
	}

	fileMap := make(map[uint]*model.UserFile, len(files))
	for _, file := range files {
		fileMap[file.ID] = file
	}

	orderedFiles := make([]*model.UserFile, 0, len(files))
	for _, id := range orderedIDs {
		if file, ok := fileMap[id]; ok {
			if params.UseCursor && params.CursorID > 0 && file.ID >= params.CursorID {
				continue
			}
			orderedFiles = append(orderedFiles, file)
		}
	}

	total := int64(len(orderedFiles))
	offset := mediaListOffset(params.UseCursor, params.Page, params.PageSize)
	if offset >= len(orderedFiles) {
		return &SearchResult{
			Files:          []*FileItem{},
			Total:          total,
			Page:           params.Page,
			PageSize:       params.PageSize,
			PaginationMode: params.PaginationMode,
			MaxPageSize:    params.MaxPageSize,
		}, nil
	}

	end := offset + params.PageSize
	if end > len(orderedFiles) {
		end = len(orderedFiles)
	}

	pageFiles := orderedFiles[offset:end]
	items := make([]*FileItem, 0, len(pageFiles))
	for _, file := range pageFiles {
		items = append(items, &FileItem{
			FileID:     fmt.Sprintf("%d", file.ID),
			FileName:   file.FileName,
			FileType:   s.getFileType(file),
			FileSize:   s.getFileSize(file),
			FilePath:   s.buildFilePath(file),
			ModifiedAt: file.UpdatedAt,
			IsFolder:   file.IsFolder,
		})
	}
	nextCursor := ""
	if params.UseCursor && len(pageFiles) > 0 {
		nextCursor = NextCursorFromUint(pageFiles[len(pageFiles)-1].ID)
	}

	return &SearchResult{
		Files:          items,
		Total:          total,
		Page:           params.Page,
		PageSize:       params.PageSize,
		PaginationMode: params.PaginationMode,
		NextCursor:     nextCursor,
		MaxPageSize:    params.MaxPageSize,
	}, nil
}

func (s *searchService) performMeiliSearch(
	ctx context.Context,
	endpoint string,
	apiKey string,
	payload meiliSearchRequest,
) (*meiliSearchResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal meilisearch request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint+"/indexes/xingyunpan_fulltext/search",
		strings.NewReader(string(body)),
	)
	if err != nil {
		return nil, fmt.Errorf("create meilisearch request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}

	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request meilisearch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("meilisearch returned status %d", resp.StatusCode)
	}

	var result meiliSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode meilisearch response failed: %w", err)
	}

	return &result, nil
}

func collectOrderedMeiliFileIDs(hits []meiliSearchHit) []uint {
	if len(hits) == 0 {
		return nil
	}

	seen := make(map[uint]struct{}, len(hits))
	result := make([]uint, 0, len(hits))
	for _, hit := range hits {
		id, ok := normalizeMeiliFileID(hit.FileID)
		if !ok || id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}

	return result
}

func normalizeMeiliFileID(value interface{}) (uint, bool) {
	switch v := value.(type) {
	case float64:
		return uint(v), true
	case string:
		id, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(id), true
	case json.Number:
		id, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return uint(id), true
	default:
		return 0, false
	}
}

func (s *searchService) applyStructuredFilters(query *gorm.DB, params *SearchParams) *gorm.DB {
	if params == nil {
		return query
	}

	if params.FileType != "" {
		query = s.applyFileTypeFilter(query, params.FileType)
	}
	if params.SizeMin != nil {
		query = query.Where("file_size >= ?", *params.SizeMin)
	}
	if params.SizeMax != nil {
		query = query.Where("file_size <= ?", *params.SizeMax)
	}
	if params.DateFrom != nil {
		query = query.Where("updated_at >= ?", *params.DateFrom)
	}
	if params.DateTo != nil {
		query = query.Where("updated_at <= ?", *params.DateTo)
	}
	if params.FolderID != nil {
		if parentID, err := strconv.ParseUint(strings.TrimSpace(*params.FolderID), 10, 64); err == nil {
			query = query.Where("parent_id = ?", uint(parentID))
		}
	}

	return query
}

func (s *searchService) applyFileTypeFilter(query *gorm.DB, fileType string) *gorm.DB {
	switch strings.ToLower(strings.TrimSpace(fileType)) {
	case "folder":
		return query.Where("is_folder = ?", true)
	case "image":
		clause, args := extensionLikeClause(categoryExtensionsFromQuery(
			configuredCategoryQuery(s.settingsRepo, "image"),
			[]string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg", "avif"},
		))
		return query.Where("is_folder = ?", false).Where(clause, args...)
	case "video":
		clause, args := extensionLikeClause(categoryExtensionsFromQuery(
			configuredCategoryQuery(s.settingsRepo, "video"),
			[]string{"mp4", "mov", "mkv", "avi", "webm", "m4v", "wmv", "flv", "mpeg", "mpg", "3gp", "3g2", "ts", "mts", "m2ts", "rm", "rmvb", "vob", "ogv", "asf", "divx"},
		))
		return query.Where("is_folder = ?", false).Where(clause, args...)
	case "audio", "music":
		clause, args := extensionLikeClause(categoryExtensionsFromQuery(
			configuredCategoryQuery(s.settingsRepo, "audio"),
			[]string{"mp3", "wav", "flac", "aac", "ogg", "oga", "m4a", "wma", "ape", "opus", "amr", "mid", "midi", "aiff", "aif", "alac", "mka", "mp2", "mpa", "weba"},
		))
		return query.Where("is_folder = ?", false).Where(clause, args...)
	case "document":
		clause, args := extensionLikeClause(categoryExtensionsFromQuery(
			configuredCategoryQuery(s.settingsRepo, "document"),
			[]string{"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "md", "csv", "rtf", "odt", "ods", "odp", "epub", "html", "htm"},
		))
		return query.Where("is_folder = ?", false).Where(clause, args...)
	case "archive":
		clause, args := extensionLikeClause([]string{"zip", "rar", "7z", "tar", "gz", "bz2", "xz"})
		return query.Where("is_folder = ?", false).Where(clause, args...)
	case "other":
		return query.Where("is_folder = ?", false)
	default:
		return query
	}
}

func extensionLikeClause(exts []string) (string, []interface{}) {
	parts := make([]string, 0, len(exts))
	args := make([]interface{}, 0, len(exts))
	for _, ext := range exts {
		parts = append(parts, "LOWER(file_name) LIKE ?")
		args = append(args, "%."+ext)
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

func (s *searchService) generateCacheKey(userID uint, params *SearchParams) string {
	data, _ := json.Marshal(params)
	hash := md5.Sum(data)
	hashStr := hex.EncodeToString(hash[:])
	return fmt.Sprintf("search:%d:%s", userID, hashStr)
}

func (s *searchService) getFromCache(ctx context.Context, key string, dest interface{}) error {
	if s.cache == nil {
		return fmt.Errorf("cache is not initialized")
	}
	return s.cache.Get(ctx, key, dest)
}

func (s *searchService) saveToCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if s.cache == nil {
		return fmt.Errorf("cache is not initialized")
	}
	return s.cache.Set(ctx, key, value, ttl)
}

func (s *searchService) normalizeParams(params *SearchParams) *SearchParams {
	if params == nil {
		params = &SearchParams{}
	}

	normalized := *params
	maxResults := 65535
	pagination := ResolvePagination(s.settingsRepo, normalized.Page, normalized.PageSize, normalized.Cursor, 20)
	if s.settingsRepo != nil {
		if setting, err := s.settingsRepo.Get(); err == nil && setting != nil {
			if setting.MaxRecursiveSearch > 0 {
				maxResults = setting.MaxRecursiveSearch
			}
		}
	}

	normalized.Page = pagination.Page
	normalized.PageSize = pagination.PageSize
	normalized.MaxResults = maxResults
	normalized.MaxPageSize = pagination.MaxPageSize
	normalized.PaginationMode = pagination.Mode
	normalized.UseCursor = pagination.UseCursor
	normalized.CursorID = pagination.Cursor

	return &normalized
}

func (s *searchService) getFileType(file *model.UserFile) string {
	if file.IsFolder {
		return "folder"
	}
	if file.PhysicalFile != nil {
		contentType := strings.ToLower(strings.TrimSpace(file.PhysicalFile.ContentType))
		switch {
		case strings.HasPrefix(contentType, "image/"):
			return "image"
		case strings.HasPrefix(contentType, "video/"):
			return "video"
		case strings.HasPrefix(contentType, "audio/"):
			return "audio"
		}
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(file.FileName)), ".")
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp", "bmp", "svg", "avif":
		return "image"
	case "mp4", "mov", "mkv", "avi", "webm", "m4v", "wmv", "flv", "mpeg", "mpg", "3gp", "3g2", "ts", "mts", "m2ts", "rm", "rmvb", "vob", "ogv", "asf", "divx":
		return "video"
	case "pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "md", "csv", "rtf", "odt", "ods", "odp", "epub", "html", "htm":
		return "document"
	case "zip", "rar", "7z", "tar", "gz", "bz2", "xz":
		return "archive"
	default:
		return "other"
	}
}

func (s *searchService) getFileSize(file *model.UserFile) int64 {
	if file == nil {
		return 0
	}
	if file.FileSize > 0 {
		return file.FileSize
	}
	if file.PhysicalFile != nil && file.PhysicalFile.FileSize > 0 {
		return file.PhysicalFile.FileSize
	}
	return 0
}

func (s *searchService) buildFilePath(file *model.UserFile) string {
	if strings.TrimSpace(file.FilePath) != "" {
		return file.FilePath
	}
	return file.FileName
}
