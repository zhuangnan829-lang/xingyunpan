package service

import (
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
)

// QueueDispatchService enqueues unified queue jobs for business events.
type QueueDispatchService interface {
	EnqueueFilePostProcess(userFileID, physicalFileID uint, fileName, storagePath, storageType string) error
	EnqueueBlobDelete(physicalFileID uint, storagePath, storageType string) error
	EnqueueMultipartCleanup(uploadID string) error
	EnqueueRecycleCleanup(recycleID uint, scheduledAt time.Time) error
	EnqueueFullTextRebuild(triggeredBy uint) (*model.QueueJob, error)
}

type queueDispatchService struct {
	settings     repository.QueueSettingRepository
	fileSettings repository.FileSystemSettingRepository
	jobs         repository.QueueJobRepository
}

// NewQueueDispatchService creates a queue dispatch service.
func NewQueueDispatchService(
	settings repository.QueueSettingRepository,
	jobs repository.QueueJobRepository,
	fileSettings ...repository.FileSystemSettingRepository,
) QueueDispatchService {
	var fileSettingsRepo repository.FileSystemSettingRepository
	if len(fileSettings) > 0 {
		fileSettingsRepo = fileSettings[0]
	}
	return &queueDispatchService{
		settings:     settings,
		fileSettings: fileSettingsRepo,
		jobs:         jobs,
	}
}

func (s *queueDispatchService) EnqueueFilePostProcess(userFileID, physicalFileID uint, fileName, storagePath, storageType string) error {
	payload := queue.FileTaskPayload{
		UserFileID:     userFileID,
		PhysicalFileID: physicalFileID,
		FileName:       fileName,
		StoragePath:    storagePath,
		StorageType:    storageType,
	}

	metadataSetting, err := queue.ResolveSetting(s.settings, string(queue.KeyMetadata))
	if err != nil {
		return err
	}
	metadataJob, err := queue.BuildMetadataJob(payload, metadataSetting)
	if err != nil {
		return err
	}
	if _, err := s.jobs.EnqueueIfAbsent(metadataJob); err != nil {
		return err
	}

	if queue.SupportsThumbnail(fileName) {
		thumbnailSetting, err := queue.ResolveSetting(s.settings, string(queue.KeyThumbnail))
		if err != nil {
			return err
		}
		thumbnailJob, err := queue.BuildThumbnailJob(payload, thumbnailSetting)
		if err != nil {
			return err
		}
		if _, err := s.jobs.EnqueueIfAbsent(thumbnailJob); err != nil {
			return err
		}
	}

	return nil
}

func (s *queueDispatchService) EnqueueBlobDelete(physicalFileID uint, storagePath, storageType string) error {
	setting, err := queue.ResolveSetting(s.settings, string(queue.KeyBlob))
	if err != nil {
		return err
	}

	job, err := queue.BuildBlobDeleteJob(queue.BlobDeletePayload{
		PhysicalFileID: physicalFileID,
		StoragePath:    storagePath,
		StorageType:    storageType,
	}, setting)
	if err != nil {
		return err
	}

	if delay := s.configuredDuration("blob"); delay > 0 {
		job.ScheduledAt = time.Now().Add(delay)
	}

	_, err = s.jobs.EnqueueIfAbsent(job)
	return err
}

func (s *queueDispatchService) EnqueueMultipartCleanup(uploadID string) error {
	setting, err := queue.ResolveSetting(s.settings, string(queue.KeyIO))
	if err != nil {
		return err
	}

	job, err := queue.BuildMultipartCleanupJob(queue.MultipartCleanupPayload{UploadID: uploadID}, setting)
	if err != nil {
		return err
	}

	_, err = s.jobs.EnqueueIfAbsent(job)
	return err
}

func (s *queueDispatchService) EnqueueRecycleCleanup(recycleID uint, scheduledAt time.Time) error {
	setting, err := queue.ResolveSetting(s.settings, string(queue.KeyOffline))
	if err != nil {
		return err
	}

	job, err := queue.BuildRecycleCleanupJob(queue.RecycleCleanupPayload{RecycleID: recycleID}, setting, scheduledAt)
	if err != nil {
		return err
	}

	_, err = s.jobs.EnqueueIfAbsent(job)
	return err
}

func (s *queueDispatchService) configuredDuration(kind string) time.Duration {
	if s.fileSettings == nil {
		return 0
	}
	setting, err := s.fileSettings.Get()
	if err != nil || setting == nil {
		return 0
	}

	raw := ""
	switch kind {
	case "blob":
		raw = setting.BlobRecycleInterval
	case "recycle":
		raw = setting.RecycleScanInterval
	}
	return parseScheduleDuration(raw)
}

func parseScheduleDuration(raw string) time.Duration {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "@every")
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return duration
}

func (s *queueDispatchService) EnqueueFullTextRebuild(triggeredBy uint) (*model.QueueJob, error) {
	setting, err := queue.ResolveSetting(s.settings, string(queue.KeyIO))
	if err != nil {
		return nil, err
	}

	job, err := queue.BuildFullTextRebuildJob(queue.FullTextRebuildPayload{TriggeredBy: triggeredBy}, setting)
	if err != nil {
		return nil, err
	}

	return s.jobs.EnqueueIfAbsent(job)
}
