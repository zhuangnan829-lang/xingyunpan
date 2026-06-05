package queue

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
)

const (
	JobTypeFileMetadata       = "file.metadata"
	JobTypeFileThumbnail      = "file.thumbnail"
	JobTypePhysicalBlobDelete = "physical.blob.delete"
	JobTypeMultipartCleanup   = "multipart.cleanup"
	JobTypeRecycleCleanup     = "recycle.cleanup"
	JobTypeFullTextRebuild    = "fulltext.rebuild"
	JobTypeOfflineDownload    = "offline.download"
	JobTypeArchiveCreate      = "archive.create"
	JobTypeArchiveExtract     = "archive.extract"

	ResourceTypeUserFile       = "user_file"
	ResourceTypePhysicalFile   = "physical_file"
	ResourceTypeMultipart      = "multipart_upload"
	ResourceTypeRecycle        = "recycle_item"
	ResourceTypeFullText       = "full_text_search"
	ResourceTypeOfflineTask    = "offline_download_task"
	ResourceTypeArchiveSession = "archive_package_session"
	ResourceTypeArchiveExtract = "archive_extract_task"
)

// FileTaskPayload is used by metadata and thumbnail jobs.
type FileTaskPayload struct {
	UserFileID     uint   `json:"user_file_id"`
	PhysicalFileID uint   `json:"physical_file_id"`
	FileName       string `json:"file_name"`
	StoragePath    string `json:"storage_path"`
	StorageType    string `json:"storage_type"`
}

// BlobDeletePayload is used by blob collection jobs.
type BlobDeletePayload struct {
	PhysicalFileID uint   `json:"physical_file_id"`
	StoragePath    string `json:"storage_path"`
	StorageType    string `json:"storage_type"`
}

// MultipartCleanupPayload is used by io cleanup jobs.
type MultipartCleanupPayload struct {
	UploadID string `json:"upload_id"`
}

// RecycleCleanupPayload is used by offline cleanup jobs.
type RecycleCleanupPayload struct {
	RecycleID uint `json:"recycle_id"`
}

// FullTextRebuildPayload is used by full text index rebuild jobs.
type FullTextRebuildPayload struct {
	TriggeredBy uint `json:"triggered_by"`
}

// OfflineDownloadPayload is used by offline download jobs.
type OfflineDownloadPayload struct {
	TaskID uint `json:"task_id"`
}

// ArchiveCreatePayload is used by archive package generation jobs.
type ArchiveCreatePayload struct {
	UserID    uint   `json:"user_id"`
	SessionID string `json:"session_id"`
	FileIDs   []uint `json:"file_ids"`
}

// ArchiveExtractPayload is used by archive extraction jobs.
type ArchiveExtractPayload struct {
	UserID uint   `json:"user_id"`
	TaskID string `json:"task_id"`
}

// EncodePayload marshals a queue payload into a DB-safe string.
func EncodePayload(payload interface{}) (string, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal queue payload failed: %w", err)
	}

	return string(bytes), nil
}

// DecodePayload unmarshals a queue payload string.
func DecodePayload(raw string, target interface{}) error {
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return fmt.Errorf("unmarshal queue payload failed: %w", err)
	}

	return nil
}

// BuildMetadataJob creates a metadata extraction job.
func BuildMetadataJob(payload FileTaskPayload, setting Setting) (*model.QueueJob, error) {
	raw, err := EncodePayload(payload)
	if err != nil {
		return nil, err
	}

	return &model.QueueJob{
		QueueKey:     string(KeyMetadata),
		JobType:      JobTypeFileMetadata,
		ResourceType: ResourceTypeUserFile,
		ResourceID:   fmt.Sprintf("%d", payload.UserFileID),
		DedupeKey:    fmt.Sprintf("%s:%s:%d", KeyMetadata, ResourceTypeUserFile, payload.UserFileID),
		Payload:      raw,
		Status:       model.QueueJobStatusPending,
		MaxAttempts:  sanitizeMaxAttempts(setting.MaxRetry),
		ScheduledAt:  time.Now(),
	}, nil
}

// BuildThumbnailJob creates a thumbnail generation job.
func BuildThumbnailJob(payload FileTaskPayload, setting Setting) (*model.QueueJob, error) {
	raw, err := EncodePayload(payload)
	if err != nil {
		return nil, err
	}

	return &model.QueueJob{
		QueueKey:     string(KeyThumbnail),
		JobType:      JobTypeFileThumbnail,
		ResourceType: ResourceTypeUserFile,
		ResourceID:   fmt.Sprintf("%d", payload.UserFileID),
		DedupeKey:    fmt.Sprintf("%s:%s:%d", KeyThumbnail, ResourceTypeUserFile, payload.UserFileID),
		Payload:      raw,
		Status:       model.QueueJobStatusPending,
		MaxAttempts:  sanitizeMaxAttempts(setting.MaxRetry),
		ScheduledAt:  time.Now(),
	}, nil
}

// BuildBlobDeleteJob creates a blob collection job.
func BuildBlobDeleteJob(payload BlobDeletePayload, setting Setting) (*model.QueueJob, error) {
	raw, err := EncodePayload(payload)
	if err != nil {
		return nil, err
	}

	return &model.QueueJob{
		QueueKey:     string(KeyBlob),
		JobType:      JobTypePhysicalBlobDelete,
		ResourceType: ResourceTypePhysicalFile,
		ResourceID:   fmt.Sprintf("%d", payload.PhysicalFileID),
		DedupeKey:    fmt.Sprintf("%s:%s:%d", KeyBlob, ResourceTypePhysicalFile, payload.PhysicalFileID),
		Payload:      raw,
		Status:       model.QueueJobStatusPending,
		MaxAttempts:  sanitizeMaxAttempts(setting.MaxRetry),
		ScheduledAt:  time.Now(),
	}, nil
}

// BuildMultipartCleanupJob creates an io cleanup job.
func BuildMultipartCleanupJob(payload MultipartCleanupPayload, setting Setting) (*model.QueueJob, error) {
	raw, err := EncodePayload(payload)
	if err != nil {
		return nil, err
	}

	delaySeconds := setting.MaxExecution
	if delaySeconds <= 0 {
		delaySeconds = 24 * 3600
	}

	return &model.QueueJob{
		QueueKey:     string(KeyIO),
		JobType:      JobTypeMultipartCleanup,
		ResourceType: ResourceTypeMultipart,
		ResourceID:   payload.UploadID,
		DedupeKey:    fmt.Sprintf("%s:%s:%s", KeyIO, ResourceTypeMultipart, payload.UploadID),
		Payload:      raw,
		Status:       model.QueueJobStatusPending,
		MaxAttempts:  sanitizeMaxAttempts(setting.MaxRetry),
		ScheduledAt:  time.Now().Add(time.Duration(delaySeconds) * time.Second),
	}, nil
}

// BuildRecycleCleanupJob creates an offline cleanup job.
func BuildRecycleCleanupJob(payload RecycleCleanupPayload, setting Setting, scheduledAt time.Time) (*model.QueueJob, error) {
	raw, err := EncodePayload(payload)
	if err != nil {
		return nil, err
	}

	if scheduledAt.IsZero() {
		scheduledAt = time.Now()
	}

	return &model.QueueJob{
		QueueKey:     string(KeyOffline),
		JobType:      JobTypeRecycleCleanup,
		ResourceType: ResourceTypeRecycle,
		ResourceID:   fmt.Sprintf("%d", payload.RecycleID),
		DedupeKey:    fmt.Sprintf("%s:%s:%d", KeyOffline, ResourceTypeRecycle, payload.RecycleID),
		Payload:      raw,
		Status:       model.QueueJobStatusPending,
		MaxAttempts:  sanitizeMaxAttempts(setting.MaxRetry),
		ScheduledAt:  scheduledAt,
	}, nil
}

// BuildFullTextRebuildJob creates a full text search rebuild job.
func BuildFullTextRebuildJob(payload FullTextRebuildPayload, setting Setting) (*model.QueueJob, error) {
	raw, err := EncodePayload(payload)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	return &model.QueueJob{
		QueueKey:     string(KeyIO),
		JobType:      JobTypeFullTextRebuild,
		ResourceType: ResourceTypeFullText,
		ResourceID:   "global",
		// Full text rebuild is an explicit admin action and should always enqueue a fresh job.
		DedupeKey:   fmt.Sprintf("%s:%s:%s:%d", KeyIO, ResourceTypeFullText, "global", now.UnixNano()),
		Payload:     raw,
		Status:      model.QueueJobStatusPending,
		MaxAttempts: sanitizeMaxAttempts(setting.MaxRetry),
		ScheduledAt: now,
	}, nil
}

// BuildOfflineDownloadJob creates a real offline download job.
func BuildOfflineDownloadJob(payload OfflineDownloadPayload, setting Setting) (*model.QueueJob, error) {
	raw, err := EncodePayload(payload)
	if err != nil {
		return nil, err
	}

	return &model.QueueJob{
		QueueKey:       string(KeyOffline),
		JobType:        JobTypeOfflineDownload,
		ResourceType:   ResourceTypeOfflineTask,
		ResourceID:     fmt.Sprintf("%d", payload.TaskID),
		DedupeKey:      fmt.Sprintf("%s:%s:%d", KeyOffline, ResourceTypeOfflineTask, payload.TaskID),
		Payload:        raw,
		NodeCapability: "offline_download",
		Status:         model.QueueJobStatusPending,
		MaxAttempts:    sanitizeMaxAttempts(setting.MaxRetry),
		ScheduledAt:    time.Now(),
	}, nil
}

// SupportsThumbnail returns whether the file is a thumbnail candidate.
func SupportsThumbnail(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	default:
		return false
	}
}

func sanitizeMaxAttempts(maxRetry int) int {
	if maxRetry < 0 {
		return 0
	}

	return maxRetry
}
