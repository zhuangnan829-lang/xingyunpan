package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/cache"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/mimetype"
	"xingyunpan-v2/pkg/redis"
	"xingyunpan-v2/pkg/storage"

	"gorm.io/gorm"

	goredis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// FileService file operations.
type FileService interface {
	Upload(userID uint, fileName string, fileSize int64, reader io.Reader, parentID *uint) (*model.UserFile, error)
	CreateFile(userID uint, fileName string, content []byte, parentID *uint) (*model.UserFile, error)
	CheckExists(fileHash string) (*model.PhysicalFile, bool, error)
	List(userID uint, parentID *uint, page, pageSize int) ([]*model.UserFile, int64, error)
	ListAfterID(userID uint, parentID *uint, afterID uint, pageSize int) ([]*model.UserFile, int64, error)
	GetDirectoryStats(userID uint, folderIDs []uint) (map[uint]DirectoryStatsPayload, error)
	GetEncryptionStatuses(userID uint, fileIDs []uint) (map[uint]EncryptionStatusPayload, error)
	Rename(userID, fileID uint, newName string) error
	Move(userID, fileID uint, newParentID *uint) error
	Copy(userID, fileID uint, targetParentID *uint) (*model.UserFile, error)
	Delete(userID, fileID uint) error
	Download(userID, fileID uint) (io.ReadCloser, string, string, error)
	DownloadWithAccess(userID, fileID uint, inline bool) (io.ReadCloser, string, string, error)
	DownloadWithDelivery(userID, fileID uint, inline bool) (*FileDownloadResult, error)
	PreviewPDF(userID, fileID uint) (io.ReadCloser, string, error)
	GetThumbnail(userID, fileID uint) ([]byte, string, error)
}

type FileDownloadResult struct {
	Reader      io.ReadCloser
	FileName    string
	ContentType string
	RedirectURL string
}

type EncryptionStatusPayload struct {
	Visible           bool   `json:"visible"`
	Encrypted         bool   `json:"encrypted"`
	StoragePolicyID   uint   `json:"storage_policy_id"`
	StoragePolicyName string `json:"storage_policy_name"`
	KeyID             string `json:"key_id,omitempty"`
}

type DirectoryStatsPayload struct {
	ChildCount      int64 `json:"child_count"`
	FileCount       int64 `json:"file_count"`
	FolderCount     int64 `json:"folder_count"`
	TotalSize       int64 `json:"total_size"`
	Cached          bool  `json:"cached"`
	TTLSeconds      int   `json:"ttl_seconds"`
	CacheEnabled    bool  `json:"cache_enabled"`
	CacheConfigured bool  `json:"cache_configured"`
}

type fileService struct {
	db                *gorm.DB
	physicalFileRepo  repository.PhysicalFileRepository
	userFileRepo      repository.UserFileRepository
	userRepo          repository.UserRepository
	settingsRepo      repository.FileSystemSettingRepository
	collaborationRepo repository.CollaborationRepository
	storage           *storage.LocalStorage
	redisClient       *goredis.Client
	cache             *cache.CacheService
	queueDispatch     QueueDispatchService
	fileEvents        FileEventService
	masterKeys        MasterKeyResolver
}

func NewFileService(
	db *gorm.DB,
	physicalFileRepo repository.PhysicalFileRepository,
	userFileRepo repository.UserFileRepository,
	userRepo repository.UserRepository,
	settingsRepo repository.FileSystemSettingRepository,
	collaborationRepo repository.CollaborationRepository,
	storage *storage.LocalStorage,
	redisClient *goredis.Client,
	cache *cache.CacheService,
	queueDispatch QueueDispatchService,
	fileEvents ...FileEventService,
) FileService {
	var eventService FileEventService
	if len(fileEvents) > 0 {
		eventService = fileEvents[0]
	}
	return &fileService{
		db:                db,
		physicalFileRepo:  physicalFileRepo,
		userFileRepo:      userFileRepo,
		userRepo:          userRepo,
		settingsRepo:      settingsRepo,
		collaborationRepo: collaborationRepo,
		storage:           storage,
		redisClient:       redisClient,
		cache:             cache,
		queueDispatch:     queueDispatch,
		fileEvents:        eventService,
		masterKeys:        NewFileSystemMasterKeyResolver(db, settingsRepo),
	}
}

func (s *fileService) Upload(userID uint, fileName string, fileSize int64, reader io.Reader, parentID *uint) (*model.UserFile, error) {
	policyRuntime := newStoragePolicyRuntime(s.db, userID)
	if err := policyRuntime.ValidateUpload(fileName, fileSize); err != nil {
		return nil, err
	}

	if parentID != nil {
		parentFolder, err := s.userFileRepo.GetByID(*parentID)
		if err != nil {
			return nil, fmt.Errorf("parent folder not found")
		}
		if parentFolder.UserID != userID {
			return nil, fmt.Errorf("no permission to access this folder")
		}
		if !parentFolder.IsFolder {
			return nil, fmt.Errorf("parent node is not a folder")
		}
	}

	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return nil, fmt.Errorf("create temp file failed: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	hasher := sha256.New()
	teeReader := io.TeeReader(reader, hasher)
	actualFileSize, err := io.Copy(tempFile, teeReader)
	if err != nil {
		return nil, fmt.Errorf("persist temp upload failed: %w", err)
	}
	if actualFileSize > 0 {
		fileSize = actualFileSize
	}
	if fileSize <= 0 {
		return nil, fmt.Errorf("uploaded file is empty")
	}
	if err := policyRuntime.ValidateUpload(fileName, fileSize); err != nil {
		return nil, err
	}

	contentType := s.detectContentType(fileName, tempFile)

	fileHash := hex.EncodeToString(hasher.Sum(nil))
	encryptionConfig, err := policyRuntime.EncryptionConfig()
	if err != nil {
		return nil, err
	}
	encryptionConfig, err = prepareBlobEncryptionConfig(encryptionConfig, s.masterKeys)
	if err != nil {
		return nil, err
	}
	fallbackStoragePath := fmt.Sprintf("files/%s", fileHash)
	policyStoragePath, err := policyRuntime.RenderBlobStoragePath(StoragePolicyBlobPathContext{
		FileName:     fileName,
		FileHash:     fileHash,
		ParentID:     parentID,
		UserFileRepo: s.userFileRepo,
		FallbackPath: fallbackStoragePath,
		StorageExists: func(path string) bool {
			return s.storage != nil && s.storage.Exists(path)
		},
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	lock := redis.NewLock(s.redisClient, fmt.Sprintf("upload:%s", fileHash), 10*time.Second)
	acquired, err := lock.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire distributed lock failed: %w", err)
	}
	if !acquired {
		time.Sleep(100 * time.Millisecond)
		acquired, err = lock.Acquire(ctx)
		if err != nil || !acquired {
			return nil, fmt.Errorf("file is being uploaded by another user")
		}
	}
	defer lock.Release(ctx)

	var userFile *model.UserFile
	var queuedPhysicalFile *model.PhysicalFile

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := reserveUserCapacity(tx, userID, fileSize); err != nil {
			return err
		}

		physicalFile, lookupErr := s.physicalFileRepo.GetByHash(fileHash)
		if lookupErr != nil {
			if _, err := tempFile.Seek(0, 0); err != nil {
				return fmt.Errorf("seek temp file failed: %w", err)
			}
			if err := maybePreAllocateBlob(policyRuntime, s.storage, policyStoragePath, physicalBlobStoredSize(fileSize, encryptionConfig)); err != nil {
				return err
			}
			if err := savePhysicalBlob(s.storage, tempFile, policyStoragePath, encryptionConfig, s.masterKeys); err != nil {
				if s.storage != nil {
					_ = s.storage.Delete(policyStoragePath)
				}
				return fmt.Errorf("save file to storage failed: %w", err)
			}

			physicalFile = &model.PhysicalFile{
				FileHash:        fileHash,
				FileSize:        fileSize,
				StoragePath:     policyStoragePath,
				RefCount:        0,
				StorageType:     "local",
				ContentType:     contentType,
				Encrypted:       encryptionConfig != nil && encryptionConfig.Enabled,
				EncryptionKeyID: encryptionKeyID(encryptionConfig),
			}
			if err := tx.Create(physicalFile).Error; err != nil {
				return err
			}
		} else {
			updates := map[string]interface{}{}
			if physicalFile.FileSize <= 0 && fileSize > 0 {
				updates["file_size"] = fileSize
				physicalFile.FileSize = fileSize
			}
			if isGenericContentType(physicalFile.ContentType) && !isGenericContentType(contentType) {
				updates["content_type"] = contentType
				physicalFile.ContentType = contentType
			}
			if len(updates) > 0 {
				if err := tx.Model(&model.PhysicalFile{}).
					Where("id = ?", physicalFile.ID).
					Updates(updates).Error; err != nil {
					return err
				}
			}
		}

		if err := tx.Model(&model.PhysicalFile{}).
			Where("id = ?", physicalFile.ID).
			UpdateColumn("ref_count", gorm.Expr("ref_count + ?", 1)).Error; err != nil {
			return err
		}

		userFile = &model.UserFile{
			UserID:         userID,
			PhysicalFileID: &physicalFile.ID,
			FileName:       fileName,
			FileSize:       fileSize,
			IsFolder:       false,
			ParentID:       parentID,
		}
		if err := tx.Create(userFile).Error; err != nil {
			return err
		}

		queuedPhysicalFile = physicalFile
		return nil
	})
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateFileList(context.Background(), userID, parentID); err != nil {
		}
		if parentID != nil {
			_ = s.cache.InvalidateDirectoryStats(context.Background(), userID, *parentID)
		}
	}

	if s.queueDispatch != nil && userFile != nil && queuedPhysicalFile != nil {
		if err := s.queueDispatch.EnqueueFilePostProcess(
			userFile.ID,
			queuedPhysicalFile.ID,
			userFile.FileName,
			queuedPhysicalFile.StoragePath,
			queuedPhysicalFile.StorageType,
		); err != nil {
			logger.Warn("enqueue file post process jobs failed", zap.Uint("user_file_id", userFile.ID), zap.Error(err))
		}
	}

	recordTrafficEvent(s.db, userID, "upload", fileSize, "file", "user_file", fmt.Sprintf("%d", userFile.ID))
	recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
		Action:       "upload",
		UserID:       userID,
		FileID:       userFile.ID,
		FileName:     userFile.FileName,
		FileSize:     fileSize,
		ResourceType: "user_file",
		ResourceID:   formatStoragePolicyFileResourceID(userFile.ID),
		Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
			"upload_mode": "direct",
		}),
	})
	s.publishFileEvent(userID, parentID, "file.created", "file", userFile.ID)

	return userFile, nil
}

func (s *fileService) CreateFile(userID uint, fileName string, content []byte, parentID *uint) (*model.UserFile, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, fmt.Errorf("file name is required")
	}
	if strings.ContainsAny(fileName, `/\`) || fileName == "." || fileName == ".." {
		return nil, fmt.Errorf("invalid file name")
	}
	if content == nil {
		content = []byte{}
	}

	return s.Upload(userID, fileName, int64(len(content)), strings.NewReader(string(content)), parentID)
}

func (s *fileService) CheckExists(fileHash string) (*model.PhysicalFile, bool, error) {
	file, err := s.physicalFileRepo.GetByHash(fileHash)
	if err != nil {
		return nil, false, nil
	}
	return file, true, nil
}

func (s *fileService) List(userID uint, parentID *uint, page, pageSize int) ([]*model.UserFile, int64, error) {
	ctx := context.Background()

	if page == 1 {
		var cachedFiles []*model.UserFile
		staticCacheTTL := s.getStaticCacheTTL()
		if staticCacheTTL > 0 && s.cache != nil && s.cache.GetFileListWithPageSize(ctx, userID, parentID, pageSize, &cachedFiles) == nil {
			s.repairMissingFileSizes(cachedFiles)
			if !hasMissingFileSizes(cachedFiles) {
				_, total, _ := s.userFileRepo.List(userID, parentID, page, pageSize)
				return cachedFiles, total, nil
			}
		}
	}

	files, total, err := s.userFileRepo.List(userID, parentID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.repairMissingFileSizes(files)

	if page == 1 && s.cache != nil {
		staticCacheTTL := s.getStaticCacheTTL()
		if staticCacheTTL > 0 {
			_ = s.cache.CacheFileListWithTTL(ctx, userID, parentID, pageSize, files, time.Duration(staticCacheTTL)*time.Second)
		}
	}

	return files, total, nil
}

func (s *fileService) ListAfterID(userID uint, parentID *uint, afterID uint, pageSize int) ([]*model.UserFile, int64, error) {
	files, total, err := s.userFileRepo.ListAfterID(userID, parentID, afterID, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.repairMissingFileSizes(files)
	return files, total, nil
}

func (s *fileService) GetDirectoryStats(userID uint, folderIDs []uint) (map[uint]DirectoryStatsPayload, error) {
	result := make(map[uint]DirectoryStatsPayload, len(folderIDs))
	if len(folderIDs) == 0 {
		return result, nil
	}

	uniqueFolderIDs := make([]uint, 0, len(folderIDs))
	seen := make(map[uint]struct{}, len(folderIDs))
	for _, folderID := range folderIDs {
		if folderID == 0 {
			continue
		}
		if _, ok := seen[folderID]; ok {
			continue
		}
		seen[folderID] = struct{}{}
		uniqueFolderIDs = append(uniqueFolderIDs, folderID)
	}

	if len(uniqueFolderIDs) == 0 {
		return result, nil
	}

	ttlSeconds := s.getDirectoryStatTTL()
	cacheEnabled := ttlSeconds > 0
	ctx := context.Background()
	pending := make([]uint, 0, len(uniqueFolderIDs))

	for _, folderID := range uniqueFolderIDs {
		entry := DirectoryStatsPayload{
			TTLSeconds:      ttlSeconds,
			CacheEnabled:    cacheEnabled,
			CacheConfigured: cacheEnabled && s.cache != nil,
		}

		if cacheEnabled && s.cache != nil {
			var cached cache.DirectoryStatsCacheEntry
			if err := s.cache.GetDirectoryStats(ctx, userID, folderID, &cached); err == nil {
				entry.ChildCount = cached.ChildCount
				entry.FileCount = cached.FileCount
				entry.FolderCount = cached.FolderCount
				entry.TotalSize = cached.TotalSize
				entry.Cached = true
				result[folderID] = entry
				continue
			}
		}

		result[folderID] = entry
		pending = append(pending, folderID)
	}

	if len(pending) == 0 {
		return result, nil
	}

	stats, err := s.userFileRepo.GetImmediateChildStats(userID, pending)
	if err != nil {
		return nil, err
	}

	for _, folderID := range pending {
		entry := result[folderID]
		if stat, ok := stats[folderID]; ok {
			entry.ChildCount = stat.ChildCount
			entry.FileCount = stat.FileCount
			entry.FolderCount = stat.FolderCount
			entry.TotalSize = stat.TotalSize
		}
		result[folderID] = entry

		if cacheEnabled && s.cache != nil {
			_ = s.cache.CacheDirectoryStats(ctx, userID, folderID, cache.DirectoryStatsCacheEntry{
				ChildCount:  entry.ChildCount,
				FileCount:   entry.FileCount,
				FolderCount: entry.FolderCount,
				TotalSize:   entry.TotalSize,
			}, time.Duration(ttlSeconds)*time.Second)
		}
	}

	return result, nil
}

func (s *fileService) GetEncryptionStatuses(userID uint, fileIDs []uint) (map[uint]EncryptionStatusPayload, error) {
	result := make(map[uint]EncryptionStatusPayload, len(fileIDs))
	if len(fileIDs) == 0 || s.db == nil {
		return result, nil
	}

	show := true
	if s.settingsRepo != nil {
		if setting, err := s.settingsRepo.Get(); err == nil && setting != nil {
			show = setting.ShowEncryptionStatus
		}
	}
	if !show {
		return result, nil
	}

	var rows []struct {
		FileID            uint
		StoragePolicyID   uint
		StoragePolicyName string
		Encrypted         bool
		KeyID             string
	}
	if err := s.db.Table("user_files").
		Select(`
			user_files.id AS file_id,
			COALESCE(storage_policies.id, 0) AS storage_policy_id,
			COALESCE(storage_policies.name, '') AS storage_policy_name,
			COALESCE(physical_files.encrypted, FALSE) AS encrypted,
			COALESCE(physical_files.encryption_key_id, '') AS key_id
		`).
		Joins("INNER JOIN users ON users.id = user_files.user_id").
		Joins("LEFT JOIN user_groups ON user_groups.id = users.user_group_id").
		Joins("LEFT JOIN storage_policies ON storage_policies.id = user_groups.storage_policy_id").
		Joins("LEFT JOIN physical_files ON physical_files.id = user_files.physical_file_id").
		Where("user_files.user_id = ? AND user_files.id IN ? AND user_files.deleted_at IS NULL", userID, fileIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		result[row.FileID] = EncryptionStatusPayload{
			Visible:           true,
			Encrypted:         row.Encrypted,
			StoragePolicyID:   row.StoragePolicyID,
			StoragePolicyName: strings.TrimSpace(row.StoragePolicyName),
			KeyID:             strings.TrimSpace(row.KeyID),
		}
	}
	for _, fileID := range fileIDs {
		if _, ok := result[fileID]; !ok {
			result[fileID] = EncryptionStatusPayload{Visible: true}
		}
	}

	return result, nil
}

func (s *fileService) Delete(userID, fileID uint) error {
	userFile, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if userFile.UserID != userID {
		return fmt.Errorf("no permission to delete this file")
	}

	parentID := userFile.ParentID
	var blobPhysicalID uint
	var blobStoragePath string
	var blobStorageType string
	var shouldEnqueueBlob bool

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.UserFile{}, fileID).Error; err != nil {
			return err
		}

		if userFile.PhysicalFileID != nil {
			var physicalFile model.PhysicalFile
			if err := tx.First(&physicalFile, *userFile.PhysicalFileID).Error; err != nil {
				return err
			}

			if err := tx.Model(&model.PhysicalFile{}).
				Where("id = ?", userFile.PhysicalFileID).
				UpdateColumn("ref_count", gorm.Expr("ref_count - ?", 1)).Error; err != nil {
				return err
			}

			if physicalFile.RefCount-1 <= 0 {
				shouldEnqueueBlob = true
				blobPhysicalID = physicalFile.ID
				blobStoragePath = physicalFile.StoragePath
				blobStorageType = physicalFile.StorageType
			}
		}

		if userFile.FileSize > 0 {
			if err := tx.Model(&model.User{}).
				Where("id = ?", userID).
				UpdateColumn("used_size", gorm.Expr("CASE WHEN used_size >= ? THEN used_size - ? ELSE 0 END", userFile.FileSize, userFile.FileSize)).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateFileList(context.Background(), userID, parentID); err != nil {
		}
		if parentID != nil {
			_ = s.cache.InvalidateDirectoryStats(context.Background(), userID, *parentID)
		}
	}

	if shouldEnqueueBlob && s.queueDispatch != nil {
		if err := s.queueDispatch.EnqueueBlobDelete(blobPhysicalID, blobStoragePath, blobStorageType); err != nil {
			logger.Warn("enqueue blob delete job failed", zap.Uint("physical_file_id", blobPhysicalID), zap.Error(err))
		}
	}

	s.publishFileEvent(userID, parentID, "file.deleted", "file", fileID)
	return nil
}

func (s *fileService) Move(userID, fileID uint, newParentID *uint) error {
	userFile, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if userFile.UserID != userID {
		return fmt.Errorf("no permission to move this file")
	}
	if userFile.IsFolder {
		return fmt.Errorf("target is not a file")
	}
	if newParentID != nil {
		if err := s.validateParentFolder(userID, *newParentID); err != nil {
			return err
		}
	}

	oldParentID := userFile.ParentID
	userFile.ParentID = newParentID
	if err := s.userFileRepo.Update(userFile); err != nil {
		return err
	}

	s.invalidateFilePlacementCaches(userID, oldParentID, newParentID)
	s.publishFileEvent(userID, oldParentID, "file.moved", "file", fileID)
	s.publishFileEvent(userID, newParentID, "file.moved", "file", fileID)
	return nil
}

func (s *fileService) Copy(userID, fileID uint, targetParentID *uint) (*model.UserFile, error) {
	source, err := s.userFileRepo.GetByIDWithPhysicalFile(fileID)
	if err != nil {
		return nil, err
	}
	if source.UserID != userID {
		return nil, fmt.Errorf("no permission to copy this file")
	}
	if source.IsFolder || source.PhysicalFileID == nil {
		return nil, fmt.Errorf("target is not a file")
	}
	if targetParentID != nil {
		if err := s.validateParentFolder(userID, *targetParentID); err != nil {
			return nil, err
		}
	}

	copied := &model.UserFile{
		UserID:         userID,
		ParentID:       targetParentID,
		FileName:       source.FileName,
		PhysicalFileID: source.PhysicalFileID,
		IsFolder:       false,
		FileSize:       source.FileSize,
		FilePath:       source.FilePath,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := reserveUserCapacity(tx, userID, source.FileSize); err != nil {
			return err
		}
		if err := tx.Model(&model.PhysicalFile{}).
			Where("id = ?", *source.PhysicalFileID).
			UpdateColumn("ref_count", gorm.Expr("ref_count + ?", 1)).Error; err != nil {
			return err
		}
		if err := tx.Create(copied).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.invalidateFilePlacementCaches(userID, nil, targetParentID)
	s.publishFileEvent(userID, targetParentID, "file.copied", "file", copied.ID)
	return copied, nil
}

func (s *fileService) Rename(userID, fileID uint, newName string) error {
	userFile, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return err
	}
	if userFile.UserID != userID {
		return fmt.Errorf("no permission to rename this file")
	}
	if userFile.IsFolder {
		return fmt.Errorf("target is not a file")
	}
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("file name cannot be empty")
	}
	if err := newStoragePolicyRuntime(s.db, userID).ValidateFileName(newName); err != nil {
		return err
	}

	userFile.FileName = newName
	if err := s.userFileRepo.Update(userFile); err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateFileList(context.Background(), userID, userFile.ParentID); err != nil {
		}
	}

	s.publishFileEvent(userID, userFile.ParentID, "file.renamed", "file", fileID)
	return nil
}

func (s *fileService) publishFileEvent(userID uint, folderID *uint, eventType string, resource string, resourceID uint) {
	if s.fileEvents == nil {
		return
	}
	s.fileEvents.Publish(userID, folderID, eventType, resource, resourceID)
}

func (s *fileService) Download(userID, fileID uint) (io.ReadCloser, string, string, error) {
	return s.DownloadWithAccess(userID, fileID, false)
}

func (s *fileService) DownloadWithAccess(userID, fileID uint, inline bool) (io.ReadCloser, string, string, error) {
	result, err := s.downloadWithDelivery(userID, fileID, inline, false)
	if err != nil {
		return nil, "", "", err
	}
	if result == nil || result.Reader == nil {
		return nil, "", "", fmt.Errorf("download is not available")
	}
	return result.Reader, result.FileName, result.ContentType, nil
}

func (s *fileService) DownloadWithDelivery(userID, fileID uint, inline bool) (*FileDownloadResult, error) {
	return s.downloadWithDelivery(userID, fileID, inline, true)
}

func (s *fileService) downloadWithDelivery(userID, fileID uint, inline bool, allowCDN bool) (*FileDownloadResult, error) {
	userFile, err := s.userFileRepo.GetByIDWithPhysicalFile(fileID)
	if err != nil {
		return nil, err
	}
	if userFile.UserID != userID {
		permission, err := s.collaborationPermission(context.Background(), userID, fileID)
		if err != nil {
			return nil, err
		}
		if !canAccessSharedFile(permission, inline) {
			return nil, fmt.Errorf("no permission to access this file")
		}
	}
	if userFile.IsFolder || userFile.PhysicalFile == nil {
		return nil, fmt.Errorf("download is only available for files")
	}
	s.repairMissingFileSize(userFile)

	contentType := strings.TrimSpace(userFile.PhysicalFile.ContentType)
	if isGenericContentType(contentType) {
		contentType = s.resolveContentTypeFromSettings(userFile.FileName)
	}

	if allowCDN && !inline && !userFile.PhysicalFile.Encrypted {
		if redirectURL, err := newStoragePolicyRuntime(s.db, userFile.UserID).CDNDownloadURL(userFile.PhysicalFile.StoragePath); err != nil {
			return nil, err
		} else if redirectURL != "" {
			recordTrafficEvent(s.db, userID, "download", userFile.FileSize, "file", "user_file", fmt.Sprintf("%d", userFile.ID))
			recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
				Action:       "download",
				UserID:       userFile.UserID,
				FileID:       userFile.ID,
				FileName:     userFile.FileName,
				FileSize:     userFile.FileSize,
				ResourceType: "user_file",
				ResourceID:   formatStoragePolicyFileResourceID(userFile.ID),
				Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
					"delivery": "cdn",
				}),
			})
			return &FileDownloadResult{
				FileName:    userFile.FileName,
				ContentType: contentType,
				RedirectURL: redirectURL,
			}, nil
		}
	}

	reader, err := readPhysicalBlob(s.storage, userFile.PhysicalFile, s.masterKeys)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	if !inline {
		recordTrafficEvent(s.db, userID, "download", userFile.FileSize, "file", "user_file", fmt.Sprintf("%d", userFile.ID))
	}
	action := "download"
	if inline {
		action = "preview"
	}
	recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
		Action:       action,
		UserID:       userFile.UserID,
		FileID:       userFile.ID,
		FileName:     userFile.FileName,
		FileSize:     userFile.FileSize,
		ResourceType: "user_file",
		ResourceID:   formatStoragePolicyFileResourceID(userFile.ID),
		Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
			"delivery": "local_stream",
		}),
	})

	return &FileDownloadResult{
		Reader:      reader,
		FileName:    userFile.FileName,
		ContentType: contentType,
	}, nil
}

func (s *fileService) PreviewPDF(userID, fileID uint) (io.ReadCloser, string, error) {
	userFile, err := s.userFileRepo.GetByIDWithPhysicalFile(fileID)
	if err != nil {
		return nil, "", err
	}
	if userFile.UserID != userID {
		permission, err := s.collaborationPermission(context.Background(), userID, fileID)
		if err != nil {
			return nil, "", err
		}
		if !canAccessSharedFile(permission, true) {
			return nil, "", fmt.Errorf("no permission to preview this file")
		}
	}
	if userFile.IsFolder || userFile.PhysicalFile == nil {
		return nil, "", fmt.Errorf("preview is only available for files")
	}

	ext := strings.ToLower(filepath.Ext(userFile.FileName))
	if ext == ".pdf" {
		reader, _, _, err := s.DownloadWithAccess(userID, fileID, true)
		return reader, replaceFileExt(userFile.FileName, ".pdf"), err
	}
	if ext != ".ppt" && ext != ".pptx" {
		return nil, "", fmt.Errorf("PDF preview conversion is only supported for ppt and pptx files")
	}

	source, err := readPhysicalBlob(s.storage, userFile.PhysicalFile, s.masterKeys)
	if err != nil {
		return nil, "", fmt.Errorf("read source file failed: %w", err)
	}
	defer source.Close()

	tmpDir, err := os.MkdirTemp("", "xingyunpan-preview-*")
	if err != nil {
		return nil, "", fmt.Errorf("create preview temp directory failed: %w", err)
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(tmpDir)
		}
	}()

	inputPath := filepath.Join(tmpDir, "source"+ext)
	outputPath := filepath.Join(tmpDir, "preview.pdf")
	input, err := os.Create(inputPath)
	if err != nil {
		return nil, "", fmt.Errorf("create preview source failed: %w", err)
	}
	if _, err := io.Copy(input, source); err != nil {
		_ = input.Close()
		return nil, "", fmt.Errorf("write preview source failed: %w", err)
	}
	if err := input.Close(); err != nil {
		return nil, "", fmt.Errorf("close preview source failed: %w", err)
	}

	if err := convertPresentationToPDF(inputPath, outputPath); err != nil {
		return nil, "", err
	}

	file, err := os.Open(outputPath)
	if err != nil {
		return nil, "", fmt.Errorf("open preview pdf failed: %w", err)
	}

	cleanup = false
	recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
		Action:       "preview",
		UserID:       userFile.UserID,
		FileID:       userFile.ID,
		FileName:     userFile.FileName,
		FileSize:     userFile.FileSize,
		ResourceType: "user_file",
		ResourceID:   formatStoragePolicyFileResourceID(userFile.ID),
		Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
			"delivery": "converted_pdf",
		}),
	})
	return &cleanupReadCloser{ReadCloser: file, cleanup: func() { _ = os.RemoveAll(tmpDir) }}, replaceFileExt(userFile.FileName, ".pdf"), nil
}

type cleanupReadCloser struct {
	io.ReadCloser
	cleanup func()
}

func (r *cleanupReadCloser) Close() error {
	err := r.ReadCloser.Close()
	if r.cleanup != nil {
		r.cleanup()
	}
	return err
}

func convertPresentationToPDF(inputPath, outputPath string) error {
	if runtime.GOOS == "windows" {
		return convertPresentationToPDFWithPowerPoint(inputPath, outputPath)
	}
	return convertPresentationToPDFWithLibreOffice(inputPath, outputPath)
}

func convertPresentationToPDFWithPowerPoint(inputPath, outputPath string) error {
	script := `
$ErrorActionPreference = 'Stop'
$inputPath = $env:XINGYUNPAN_PREVIEW_INPUT
$outputPath = $env:XINGYUNPAN_PREVIEW_OUTPUT
$powerPoint = $null
$presentation = $null
try {
  $powerPoint = New-Object -ComObject PowerPoint.Application
  $presentation = $powerPoint.Presentations.Open($inputPath, $true, $false, $false)
  $presentation.SaveAs($outputPath, 32)
} finally {
  if ($presentation -ne $null) { $presentation.Close() | Out-Null }
  if ($powerPoint -ne $null) { $powerPoint.Quit() | Out-Null }
  [System.GC]::Collect()
  [System.GC]::WaitForPendingFinalizers()
}
`
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.Env = append(os.Environ(),
		"XINGYUNPAN_PREVIEW_INPUT="+inputPath,
		"XINGYUNPAN_PREVIEW_OUTPUT="+outputPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PowerPoint PDF conversion failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("PowerPoint did not generate preview PDF: %w", err)
	}
	return nil
}

func convertPresentationToPDFWithLibreOffice(inputPath, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	cmd := exec.Command("soffice", "--headless", "--convert-to", "pdf", "--outdir", outputDir, inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("LibreOffice PDF conversion failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	generatedPath := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))+".pdf")
	if generatedPath != outputPath {
		if err := os.Rename(generatedPath, outputPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("rename generated preview PDF failed: %w", err)
		}
	}
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("LibreOffice did not generate preview PDF: %w", err)
	}
	return nil
}

func replaceFileExt(fileName, ext string) string {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if base == "" {
		base = "preview"
	}
	return base + ext
}

func (s *fileService) collaborationPermission(ctx context.Context, userID, fileID uint) (string, error) {
	if s.collaborationRepo == nil {
		return "", fmt.Errorf("no permission to access this file")
	}

	permission, err := s.collaborationRepo.GetPermission(ctx, fileID, userID)
	if err != nil {
		return "", fmt.Errorf("no permission to access this file")
	}
	return strings.ToLower(strings.TrimSpace(permission)), nil
}

func canAccessSharedFile(permission string, inline bool) bool {
	switch permission {
	case "edit", "download":
		return true
	case "view":
		return inline
	default:
		return false
	}
}

func (s *fileService) GetThumbnail(userID, fileID uint) ([]byte, string, error) {
	userFile, err := s.userFileRepo.GetByIDWithPhysicalFile(fileID)
	if err != nil {
		return nil, "", err
	}
	if userFile.UserID != userID {
		return nil, "", fmt.Errorf("no permission to access this file")
	}
	if userFile.IsFolder || userFile.PhysicalFile == nil || userFile.PhysicalFileID == nil {
		return nil, "", fmt.Errorf("thumbnail not available")
	}

	thumbnailPath := fmt.Sprintf("thumbnails/%d.jpg", *userFile.PhysicalFileID)
	if !s.storage.Exists(thumbnailPath) {
		return nil, "", fmt.Errorf("thumbnail not available")
	}

	reader, err := s.storage.Read(thumbnailPath)
	if err != nil {
		return nil, "", fmt.Errorf("read thumbnail failed: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("read thumbnail bytes failed: %w", err)
	}

	return data, "image/jpeg", nil
}

func (s *fileService) resolveContentTypeFromSettings(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return "application/octet-stream"
	}

	if s.settingsRepo != nil {
		setting, err := s.settingsRepo.Get()
		if err == nil && setting != nil && strings.TrimSpace(setting.MimeMap) != "" {
			var mimeMap map[string]string
			if err := json.Unmarshal([]byte(setting.MimeMap), &mimeMap); err == nil {
				if value := strings.TrimSpace(mimeMap[ext]); value != "" {
					return value
				}
			}
		}
	}

	if value := strings.TrimSpace(mimetype.FromFileName(fileName)); value != "" {
		return value
	}

	return "application/octet-stream"
}

func (s *fileService) detectContentType(fileName string, file *os.File) string {
	if file != nil {
		var header [512]byte
		if _, err := file.Seek(0, 0); err == nil {
			n, _ := file.Read(header[:])
			_, _ = file.Seek(0, 0)
			if n > 0 {
				detected := strings.TrimSpace(http.DetectContentType(header[:n]))
				if !isGenericContentType(detected) {
					return detected
				}
			}
		}
	}

	return s.resolveContentTypeFromSettings(fileName)
}

func isGenericContentType(contentType string) bool {
	return mimetype.IsGeneric(contentType)
}

func (s *fileService) validateParentFolder(userID, parentID uint) error {
	parentFolder, err := s.userFileRepo.GetByID(parentID)
	if err != nil {
		return fmt.Errorf("parent folder not found")
	}
	if parentFolder.UserID != userID {
		return fmt.Errorf("no permission to access parent folder")
	}
	if !parentFolder.IsFolder {
		return fmt.Errorf("parent node is not a folder")
	}
	return nil
}

func (s *fileService) invalidateFilePlacementCaches(userID uint, oldParentID, newParentID *uint) {
	if s.cache == nil {
		return
	}

	ctx := context.Background()
	if oldParentID != nil {
		_ = s.cache.InvalidateFileList(ctx, userID, oldParentID)
		_ = s.cache.InvalidateDirectoryStats(ctx, userID, *oldParentID)
	} else {
		_ = s.cache.InvalidateFileList(ctx, userID, nil)
	}

	if newParentID != nil {
		_ = s.cache.InvalidateFileList(ctx, userID, newParentID)
		_ = s.cache.InvalidateDirectoryStats(ctx, userID, *newParentID)
	} else {
		_ = s.cache.InvalidateFileList(ctx, userID, nil)
	}
}

func (s *fileService) repairMissingFileSizes(files []*model.UserFile) {
	for _, file := range files {
		s.repairMissingFileSize(file)
	}
}

func (s *fileService) repairMissingFileSize(file *model.UserFile) {
	if file == nil || file.IsFolder || file.PhysicalFile == nil || s.storage == nil {
		return
	}
	resolvedContentType := s.resolveContentTypeFromSettings(file.FileName)
	shouldRepairContentType := isGenericContentType(file.PhysicalFile.ContentType) && !isGenericContentType(resolvedContentType)
	if file.FileSize > 0 && file.PhysicalFile.FileSize > 0 && !shouldRepairContentType {
		return
	}
	var size int64
	if file.FileSize <= 0 || file.PhysicalFile.FileSize <= 0 {
		storagePath := strings.TrimSpace(file.PhysicalFile.StoragePath)
		if storagePath != "" {
			if storageSize, err := s.storage.Size(storagePath); err == nil && storageSize > 0 {
				size = storageSize
				file.FileSize = storageSize
				file.PhysicalFile.FileSize = storageSize
			}
		}
	}

	if shouldRepairContentType {
		file.PhysicalFile.ContentType = resolvedContentType
	}

	if s.db == nil {
		return
	}
	if size > 0 {
		_ = s.db.Model(&model.UserFile{}).
			Where("id = ? AND file_size <= 0", file.ID).
			Update("file_size", size).Error
	}
	if file.PhysicalFileID != nil {
		updates := map[string]interface{}{}
		if size > 0 {
			updates["file_size"] = size
		}
		if shouldRepairContentType {
			updates["content_type"] = resolvedContentType
		}
		if len(updates) > 0 {
			_ = s.db.Model(&model.PhysicalFile{}).
				Where("id = ?", *file.PhysicalFileID).
				Updates(updates).Error
		}
	}
}

func hasMissingFileSizes(files []*model.UserFile) bool {
	for _, file := range files {
		if file != nil && !file.IsFolder && file.PhysicalFileID != nil && file.FileSize <= 0 {
			return true
		}
	}
	return false
}

func (s *fileService) getDirectoryStatTTL() int {
	if s.settingsRepo == nil {
		return 0
	}

	setting, err := s.settingsRepo.Get()
	if err != nil || setting == nil || setting.DirectoryStatTTL < 0 {
		return 0
	}

	return setting.DirectoryStatTTL
}

func (s *fileService) getStaticCacheTTL() int {
	if s.settingsRepo == nil {
		return 60
	}

	setting, err := s.settingsRepo.Get()
	if err != nil || setting == nil || setting.StaticCacheTTL < 0 {
		return 60
	}

	return setting.StaticCacheTTL
}
