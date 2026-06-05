package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/storage"
)

// Executor executes unified queue jobs.
type Executor struct {
	db            *gorm.DB
	storage       storage.MultipartStorage
	physicalFiles repository.PhysicalFileRepository
	userFiles     repository.UserFileRepository
	multiparts    repository.MultipartUploadRepository
	recycles      repository.RecycleRepository
	offline       OfflineDownloadExecutor
	archive       ArchiveExecutor
}

// OfflineDownloadExecutor is implemented by the offline download service.
type OfflineDownloadExecutor interface {
	ExecuteOfflineDownload(ctx context.Context, taskID uint) (string, error)
}

// ArchiveExecutor is implemented by the file system runtime service.
type ArchiveExecutor interface {
	ExecuteArchiveCreate(ctx context.Context, userID uint, sessionID string, fileIDs []uint) (string, error)
	ExecuteArchiveExtract(ctx context.Context, userID uint, taskID string) (string, error)
}

// NewExecutor creates a queue job executor.
func NewExecutor(
	db *gorm.DB,
	stor storage.MultipartStorage,
	physicalFiles repository.PhysicalFileRepository,
	userFiles repository.UserFileRepository,
	multiparts repository.MultipartUploadRepository,
	recycles repository.RecycleRepository,
	offline ...OfflineDownloadExecutor,
) *Executor {
	var offlineExecutor OfflineDownloadExecutor
	if len(offline) > 0 {
		offlineExecutor = offline[0]
	}

	return &Executor{
		db:            db,
		storage:       stor,
		physicalFiles: physicalFiles,
		userFiles:     userFiles,
		multiparts:    multiparts,
		recycles:      recycles,
		offline:       offlineExecutor,
	}
}

func (e *Executor) SetArchiveExecutor(archive ArchiveExecutor) {
	e.archive = archive
}

// Execute runs one queue job and returns a structured result string.
func (e *Executor) Execute(ctx context.Context, job model.QueueJob) (string, error) {
	switch job.JobType {
	case JobTypeFileMetadata:
		return e.handleMetadata(ctx, job)
	case JobTypeFileThumbnail:
		return e.handleThumbnail(ctx, job)
	case JobTypePhysicalBlobDelete:
		return e.handleBlobDelete(ctx, job)
	case JobTypeMultipartCleanup:
		return e.handleMultipartCleanup(ctx, job)
	case JobTypeRecycleCleanup:
		return e.handleRecycleCleanup(ctx, job)
	case JobTypeFullTextRebuild:
		return e.handleFullTextRebuild(ctx, job)
	case JobTypeOfflineDownload:
		return e.handleOfflineDownload(ctx, job)
	case JobTypeArchiveCreate:
		return e.handleArchiveCreate(ctx, job)
	case JobTypeArchiveExtract:
		return e.handleArchiveExtract(ctx, job)
	default:
		return "", fmt.Errorf("unsupported queue job type: %s", job.JobType)
	}
}

func (e *Executor) handleArchiveCreate(ctx context.Context, job model.QueueJob) (string, error) {
	if e.archive == nil {
		return "", fmt.Errorf("archive executor is not initialized")
	}
	var payload ArchiveCreatePayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}
	if payload.SessionID == "" {
		payload.SessionID = strings.TrimSpace(job.ResourceID)
	}
	if payload.UserID == 0 || payload.SessionID == "" || len(payload.FileIDs) == 0 {
		return "", fmt.Errorf("archive create payload is incomplete")
	}
	return e.archive.ExecuteArchiveCreate(ctx, payload.UserID, payload.SessionID, payload.FileIDs)
}

func (e *Executor) handleArchiveExtract(ctx context.Context, job model.QueueJob) (string, error) {
	if e.archive == nil {
		return "", fmt.Errorf("archive executor is not initialized")
	}
	var payload ArchiveExtractPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}
	if payload.TaskID == "" {
		payload.TaskID = strings.TrimSpace(job.ResourceID)
	}
	if payload.UserID == 0 || payload.TaskID == "" {
		return "", fmt.Errorf("archive extract payload is incomplete")
	}
	return e.archive.ExecuteArchiveExtract(ctx, payload.UserID, payload.TaskID)
}

func (e *Executor) handleOfflineDownload(ctx context.Context, job model.QueueJob) (string, error) {
	var payload OfflineDownloadPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}
	if payload.TaskID == 0 {
		parsed, err := strconv.ParseUint(strings.TrimSpace(job.ResourceID), 10, 32)
		if err != nil || parsed == 0 {
			return "", fmt.Errorf("offline download task id is empty")
		}
		payload.TaskID = uint(parsed)
	}
	if e.offline == nil {
		return "", fmt.Errorf("offline download executor is not initialized")
	}
	return e.offline.ExecuteOfflineDownload(ctx, payload.TaskID)
}

func (e *Executor) handleMetadata(ctx context.Context, job model.QueueJob) (string, error) {
	var payload FileTaskPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	metadata, err := extractMediaMetadata(e.storage, payload.StoragePath, payload.FileName)
	if err != nil {
		return "", err
	}

	if e.db != nil && payload.PhysicalFileID > 0 && metadata.ContentType != "" {
		if err := e.db.Model(&model.PhysicalFile{}).
			Where("id = ?", payload.PhysicalFileID).
			Update("content_type", metadata.ContentType).Error; err != nil {
			return "", fmt.Errorf("update physical file content type failed: %w", err)
		}
	}

	result := map[string]interface{}{
		"user_file_id":      payload.UserFileID,
		"physical_file_id":  payload.PhysicalFileID,
		"extension":         metadata.Extension,
		"storage_type":      payload.StorageType,
		"thumbnail_support": SupportsThumbnail(payload.FileName),
		"content_type":      metadata.ContentType,
		"width":             metadata.Width,
		"height":            metadata.Height,
		"is_image":          metadata.IsImage,
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal metadata result failed: %w", err)
	}

	return string(bytes), nil
}

func (e *Executor) handleThumbnail(ctx context.Context, job model.QueueJob) (string, error) {
	var payload FileTaskPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if !SupportsThumbnail(payload.FileName) {
		return marshalThumbnailSkippedResult(payload.FileName)
	}

	if !e.storage.Exists(payload.StoragePath) {
		return "", fmt.Errorf("thumbnail source file missing: %s", payload.StoragePath)
	}

	fileHash := fmt.Sprintf("%d", payload.PhysicalFileID)
	if fileHash == "0" {
		fileHash = strings.TrimSuffix(filepath.Base(payload.StoragePath), filepath.Ext(payload.StoragePath))
	}

	thumbnailPath := fmt.Sprintf("thumbnails/%s.jpg", fileHash)
	result, err := generateThumbnail(e.storage, payload.StoragePath, thumbnailPath)
	if err != nil {
		return "", err
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal thumbnail result failed: %w", err)
	}

	return string(bytes), nil
}

func marshalThumbnailSkippedResult(fileName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = "(none)"
	}
	result := map[string]interface{}{
		"status":               "skipped",
		"reason":               "unsupported_thumbnail_format",
		"file_name":            fileName,
		"extension":            ext,
		"supported_extensions": []string{".jpg", ".jpeg", ".png", ".gif"},
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal thumbnail skipped result failed: %w", err)
	}
	return string(bytes), nil
}

func (e *Executor) handleBlobDelete(ctx context.Context, job model.QueueJob) (string, error) {
	var payload BlobDeletePayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	file, err := e.physicalFiles.GetByID(payload.PhysicalFileID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return `{"status":"skipped","reason":"physical_file_missing"}`, nil
		}
		return "", err
	}

	if file.RefCount > 0 {
		return fmt.Sprintf(`{"status":"skipped","reason":"ref_count_positive","ref_count":%d}`, file.RefCount), nil
	}

	if payload.StoragePath != "" && e.storage.Exists(payload.StoragePath) {
		if err := e.storage.Delete(payload.StoragePath); err != nil {
			return "", fmt.Errorf("delete blob file failed: %w", err)
		}
	}

	if err := e.physicalFiles.Delete(payload.PhysicalFileID); err != nil && err != gorm.ErrRecordNotFound {
		return "", fmt.Errorf("delete physical file record failed: %w", err)
	}

	return `{"status":"completed"}`, nil
}

func (e *Executor) handleMultipartCleanup(ctx context.Context, job model.QueueJob) (string, error) {
	var payload MultipartCleanupPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	upload, err := e.multiparts.GetByUploadID(payload.UploadID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return `{"status":"skipped","reason":"upload_missing"}`, nil
		}
		return "", err
	}

	if upload.Status != model.MultipartStatusUploading {
		return fmt.Sprintf(`{"status":"skipped","reason":"upload_status_%s"}`, upload.Status), nil
	}

	chunkPaths := make([]string, upload.TotalChunks)
	for i := 1; i <= upload.TotalChunks; i++ {
		chunkPaths[i-1] = fmt.Sprintf("multipart/%s/chunk_%d", upload.UploadID, i)
	}
	if err := e.storage.DeleteChunks(chunkPaths); err != nil {
		return "", fmt.Errorf("delete multipart chunks failed: %w", err)
	}

	if err := e.multiparts.Delete(upload.UploadID); err != nil {
		return "", fmt.Errorf("delete multipart upload record failed: %w", err)
	}

	return `{"status":"completed"}`, nil
}

func (e *Executor) handleRecycleCleanup(ctx context.Context, job model.QueueJob) (string, error) {
	var payload RecycleCleanupPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	item, err := e.recycles.GetByID(ctx, payload.RecycleID)
	if err != nil {
		return `{"status":"skipped","reason":"recycle_item_missing"}`, nil
	}

	if e.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	err = e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		files, fileErr := e.collectRecycleSubtreeFiles(ctx, tx, item.UserID, item.FileID)
		if fileErr != nil {
			return fileErr
		}

		fileIDs := make([]uint, 0, len(files))
		physicalRefs := make(map[uint]int)
		for _, file := range files {
			fileIDs = append(fileIDs, file.ID)
			if !file.IsFolder && file.PhysicalFileID != nil {
				physicalRefs[*file.PhysicalFileID]++
			}
		}

		for physicalID, removedRefs := range physicalRefs {
			var physical model.PhysicalFile
			if err := tx.Unscoped().First(&physical, physicalID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					continue
				}
				return fmt.Errorf("query recycled physical file failed: %w", err)
			}

			if physical.RefCount <= removedRefs {
				if physical.StoragePath != "" && e.storage.Exists(physical.StoragePath) {
					if err := e.storage.Delete(physical.StoragePath); err != nil {
						return fmt.Errorf("delete recycled physical file failed: %w", err)
					}
				}
				if err := tx.Unscoped().Delete(&model.PhysicalFile{}, physicalID).Error; err != nil {
					return fmt.Errorf("delete recycled physical file record failed: %w", err)
				}
				continue
			}

			if err := tx.Model(&model.PhysicalFile{}).
				Where("id = ?", physicalID).
				UpdateColumn("ref_count", gorm.Expr("ref_count - ?", removedRefs)).Error; err != nil {
				return fmt.Errorf("decrement recycled physical file ref count failed: %w", err)
			}
		}

		if len(fileIDs) > 0 {
			if err := tx.Unscoped().Delete(&model.UserFile{}, fileIDs).Error; err != nil {
				return fmt.Errorf("delete recycled user files failed: %w", err)
			}
		}

		if err := tx.Unscoped().Delete(&model.RecycleBin{}, item.ID).Error; err != nil {
			return fmt.Errorf("delete recycle item failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return `{"status":"completed"}`, nil
}

func (e *Executor) collectRecycleSubtreeFiles(ctx context.Context, tx *gorm.DB, userID uint, rootID uint) ([]model.UserFile, error) {
	ids := []uint{rootID}
	queue := []uint{rootID}
	visited := map[uint]struct{}{}

	for len(queue) > 0 {
		parentID := queue[0]
		queue = queue[1:]
		if _, ok := visited[parentID]; ok {
			continue
		}
		visited[parentID] = struct{}{}

		var children []model.UserFile
		if err := tx.WithContext(ctx).
			Unscoped().
			Where("user_id = ? AND parent_id = ?", userID, parentID).
			Find(&children).Error; err != nil {
			return nil, fmt.Errorf("query recycled child files failed: %w", err)
		}

		for _, child := range children {
			ids = append(ids, child.ID)
			if child.IsFolder {
				queue = append(queue, child.ID)
			}
		}
	}

	ids = uniqueQueueUintIDs(ids)
	var files []model.UserFile
	if err := tx.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND id IN ?", userID, ids).
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("query recycled files failed: %w", err)
	}

	return files, nil
}

func uniqueQueueUintIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func (e *Executor) handleFullTextRebuild(ctx context.Context, job model.QueueJob) (string, error) {
	var payload FullTextRebuildPayload
	if err := DecodePayload(job.Payload, &payload); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if e.db == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	var setting model.FullTextSearchSetting
	if err := e.db.Order("id asc").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("full text search settings not configured")
		}
		return "", fmt.Errorf("load full text search settings failed: %w", err)
	}

	if !setting.Enabled {
		return `{"status":"skipped","reason":"full_text_search_disabled"}`, nil
	}

	extensions := normalizeExtensions(setting.Extensions)
	if len(extensions) == 0 {
		return `{"status":"skipped","reason":"extensions_empty"}`, nil
	}

	return e.rebuildFullTextIndex(ctx, payload, setting)
}

func normalizeExtensions(raw string) []string {
	items := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(strings.ToLower(item))
		item = strings.TrimPrefix(item, ".")
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
