package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/mimetype"
	"xingyunpan-v2/pkg/redis"
	"xingyunpan-v2/pkg/storage"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MultipartService interface {
	InitMultipartUpload(ctx context.Context, userID uint, fileName, fileHash string, fileSize int64, chunkSize int, parentID *uint) (*InitMultipartUploadResult, error)
	GetPresignedURLs(ctx context.Context, uploadID string, userID uint) (*PresignedURLsResult, error)
	RecordChunkUpload(ctx context.Context, uploadID string, userID uint, chunkNumber int, etag string, options ...ChunkRecordOptions) error
	GetCompletedChunks(ctx context.Context, uploadID string, userID uint) ([]int, error)
	CompleteMultipartUpload(ctx context.Context, uploadID string, userID uint, parentID *uint) (*model.UserFile, error)
	CancelMultipartUpload(ctx context.Context, uploadID string, userID uint) error
	GetUploadProgress(ctx context.Context, uploadID string, userID uint) (*UploadProgress, error)
	ListUploadTasks(ctx context.Context, userID uint, status string, page, pageSize int) ([]UploadTaskSummary, int64, error)
}

type multipartService struct {
	db                   *gorm.DB
	multipartRepo        repository.MultipartUploadRepository
	physicalFileRepo     repository.PhysicalFileRepository
	userFileRepo         repository.UserFileRepository
	userRepo             repository.UserRepository
	settingsRepo         repository.FileSystemSettingRepository
	storage              storage.MultipartStorage
	redisMultipart       *redis.MultipartRedis
	storageType          string
	defaultChunkSize     int64
	multipartExpireHours int
	queueDispatch        QueueDispatchService
	masterKeys           MasterKeyResolver
}

type InitMultipartUploadResult struct {
	UploadID              string `json:"upload_id"`
	TotalChunks           int    `json:"total_chunks"`
	ChunkSize             int    `json:"chunk_size"`
	ParallelChunkCount    int    `json:"parallel_chunk_count,omitempty"`
	MaxChunkRetry         int    `json:"max_chunk_retry"`
	TransferParallelism   int    `json:"transfer_parallelism"`
	CacheChunksForRetry   bool   `json:"cache_chunks_for_retry"`
	UploadSessionTTL      int    `json:"upload_session_ttl"`
	BlobSignedURLTTL      int    `json:"blob_signed_url_ttl"`
	BlobSignedURLReuseTTL int    `json:"blob_signed_url_reuse_ttl"`
	ExpiresAt             int64  `json:"expires_at"`
}

type PresignedURLsResult struct {
	UploadID              string         `json:"upload_id"`
	URLs                  map[int]string `json:"urls"`
	ExpiresAt             int64          `json:"expires_at"`
	MaxChunkRetry         int            `json:"max_chunk_retry"`
	TransferParallelism   int            `json:"transfer_parallelism"`
	CacheChunksForRetry   bool           `json:"cache_chunks_for_retry"`
	BlobSignedURLTTL      int            `json:"blob_signed_url_ttl"`
	BlobSignedURLReuseTTL int            `json:"blob_signed_url_reuse_ttl"`
}

type UploadProgress struct {
	UploadID            string  `json:"upload_id"`
	TotalChunks         int     `json:"total_chunks"`
	CompletedChunks     int     `json:"completed_chunks"`
	Progress            float64 `json:"progress"`
	Status              string  `json:"status"`
	MaxChunkRetry       int     `json:"max_chunk_retry"`
	TransferParallelism int     `json:"transfer_parallelism"`
	CacheChunksForRetry bool    `json:"cache_chunks_for_retry"`
}

type ChunkRecordOptions struct {
	Attempt         int
	ActiveTransfers int
}

type multipartRuntimeSnapshot struct {
	MaxChunkRetry       int
	TransferParallelism int
	CacheChunksForRetry bool
	UploadSessionTTL    int
}

type UploadTaskSummary struct {
	UploadID        string     `json:"upload_id"`
	FileName        string     `json:"file_name"`
	FileSize        int64      `json:"file_size"`
	TotalChunks     int        `json:"total_chunks"`
	CompletedChunks int        `json:"completed_chunks"`
	Progress        float64    `json:"progress"`
	Status          string     `json:"status"`
	StorageType     string     `json:"storage_type"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

func NewMultipartService(
	db *gorm.DB,
	multipartRepo repository.MultipartUploadRepository,
	physicalFileRepo repository.PhysicalFileRepository,
	userFileRepo repository.UserFileRepository,
	userRepo repository.UserRepository,
	settingsRepo repository.FileSystemSettingRepository,
	storage storage.MultipartStorage,
	redisMultipart *redis.MultipartRedis,
	storageType string,
	defaultChunkSize int64,
	multipartExpireHours int,
	queueDispatch QueueDispatchService,
) MultipartService {
	return &multipartService{
		db:                   db,
		multipartRepo:        multipartRepo,
		physicalFileRepo:     physicalFileRepo,
		userFileRepo:         userFileRepo,
		userRepo:             userRepo,
		settingsRepo:         settingsRepo,
		storage:              storage,
		redisMultipart:       redisMultipart,
		storageType:          storageType,
		defaultChunkSize:     defaultChunkSize,
		multipartExpireHours: multipartExpireHours,
		queueDispatch:        queueDispatch,
		masterKeys:           NewFileSystemMasterKeyResolver(db, settingsRepo),
	}
}

func (s *multipartService) InitMultipartUpload(ctx context.Context, userID uint, fileName, fileHash string, fileSize int64, chunkSize int, parentID *uint) (*InitMultipartUploadResult, error) {
	policyRuntime := newStoragePolicyRuntime(s.db, userID)
	if err := policyRuntime.ValidateUpload(fileName, fileSize); err != nil {
		return nil, err
	}

	uploadID := uuid.New().String()
	runtimeSnapshot := s.uploadRuntimeSnapshot()

	parallelChunkCount := 0
	if tuning, err := policyRuntime.UploadTuning(); err != nil {
		return nil, err
	} else if tuning != nil && tuning.HasPolicy {
		if tuning.ChunkSizeBytes > 0 {
			chunkSize = tuning.ChunkSizeBytes
		}
		if tuning.ParallelChunkCount > 0 {
			parallelChunkCount = tuning.ParallelChunkCount
		}
	}
	if parallelChunkCount <= 0 || parallelChunkCount > runtimeSnapshot.TransferParallelism {
		parallelChunkCount = runtimeSnapshot.TransferParallelism
	}

	if chunkSize <= 0 {
		chunkSize = int(s.defaultChunkSize)
	}
	totalChunks := int((fileSize + int64(chunkSize) - 1) / int64(chunkSize))

	upload := &model.MultipartUpload{
		UploadID:    uploadID,
		UserID:      userID,
		FileName:    fileName,
		FileHash:    fileHash,
		FileSize:    fileSize,
		TotalChunks: totalChunks,
		ChunkSize:   chunkSize,
		StorageType: s.storageType,
		Status:      model.MultipartStatusUploading,
	}

	if err := s.multipartRepo.Create(upload); err != nil {
		return nil, fmt.Errorf("create upload task failed: %w", err)
	}

	metadata := map[string]interface{}{
		"upload_id":              uploadID,
		"user_id":                userID,
		"file_name":              fileName,
		"file_hash":              fileHash,
		"file_size":              fileSize,
		"total_chunks":           totalChunks,
		"chunk_size":             chunkSize,
		"storage_type":           s.storageType,
		"max_chunk_retry":        runtimeSnapshot.MaxChunkRetry,
		"transfer_parallelism":   runtimeSnapshot.TransferParallelism,
		"cache_chunks_for_retry": runtimeSnapshot.CacheChunksForRetry,
		"upload_session_ttl":     runtimeSnapshot.UploadSessionTTL,
	}
	if parentID != nil {
		metadata["parent_id"] = *parentID
	}

	expiration := time.Duration(runtimeSnapshot.UploadSessionTTL) * time.Second
	if err := s.redisMultipart.SetUploadMetadata(ctx, uploadID, metadata, expiration); err != nil {
		return nil, fmt.Errorf("set upload metadata failed: %w", err)
	}

	if err := s.redisMultipart.SetExpiration(ctx, uploadID, expiration); err != nil {
		return nil, fmt.Errorf("set expiration failed: %w", err)
	}

	if s.queueDispatch != nil {
		if err := s.queueDispatch.EnqueueMultipartCleanup(uploadID); err != nil {
			logger.Warn("enqueue multipart cleanup job failed", zap.String("upload_id", uploadID), zap.Error(err))
		}
	}

	recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
		Action:       "multipart_init",
		UserID:       userID,
		FileName:     fileName,
		FileSize:     fileSize,
		ResourceType: "multipart_upload",
		ResourceID:   uploadID,
		Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
			"chunk_size":           chunkSize,
			"parallel_chunk_count": parallelChunkCount,
			"total_chunks":         totalChunks,
		}),
	})

	return &InitMultipartUploadResult{
		UploadID:              uploadID,
		TotalChunks:           totalChunks,
		ChunkSize:             chunkSize,
		ParallelChunkCount:    parallelChunkCount,
		MaxChunkRetry:         runtimeSnapshot.MaxChunkRetry,
		TransferParallelism:   runtimeSnapshot.TransferParallelism,
		CacheChunksForRetry:   runtimeSnapshot.CacheChunksForRetry,
		UploadSessionTTL:      runtimeSnapshot.UploadSessionTTL,
		BlobSignedURLTTL:      int(s.getBlobSignedURLTTL().Seconds()),
		BlobSignedURLReuseTTL: int(s.getBlobSignedURLReuseTTL().Seconds()),
		ExpiresAt:             time.Now().Add(expiration).Unix(),
	}, nil
}

func (s *multipartService) GetPresignedURLs(ctx context.Context, uploadID string, userID uint) (*PresignedURLsResult, error) {
	upload, err := s.multipartRepo.GetByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload task not found: %w", err)
	}
	if upload.UserID != userID {
		return nil, fmt.Errorf("no permission to access this upload task")
	}
	if upload.Status != model.MultipartStatusUploading {
		return nil, fmt.Errorf("invalid upload task status: %s", upload.Status)
	}
	runtimeSnapshot := s.uploadRuntimeSnapshotForSession(ctx, uploadID)

	storagePath := fmt.Sprintf("multipart/%s", uploadID)
	urls := make(map[int]string)
	expiration := s.getBlobSignedURLTTL()
	reuseTTL := s.getBlobSignedURLReuseTTL()

	if s.redisMultipart != nil {
		cached, cacheErr := s.redisMultipart.GetPresignedURLCache(ctx, uploadID)
		if cacheErr == nil && cached != nil {
			if time.Until(time.Unix(cached.ExpiresAt, 0)) > reuseTTL {
				return &PresignedURLsResult{
					UploadID:              cached.UploadID,
					URLs:                  cached.URLs,
					ExpiresAt:             cached.ExpiresAt,
					MaxChunkRetry:         runtimeSnapshot.MaxChunkRetry,
					TransferParallelism:   runtimeSnapshot.TransferParallelism,
					CacheChunksForRetry:   runtimeSnapshot.CacheChunksForRetry,
					BlobSignedURLTTL:      int(expiration.Seconds()),
					BlobSignedURLReuseTTL: int(reuseTTL.Seconds()),
				}, nil
			}
			_ = s.redisMultipart.DeletePresignedURLCache(ctx, uploadID)
		}
	}

	for i := 1; i <= upload.TotalChunks; i++ {
		chunkPath := fmt.Sprintf("%s/chunk_%d", storagePath, i)
		url, urlErr := s.storage.GeneratePresignedURL(chunkPath, i, int(expiration.Seconds()))
		if urlErr != nil {
			return nil, fmt.Errorf("generate presigned url failed: %w", urlErr)
		}
		urls[i] = url
	}

	result := &PresignedURLsResult{
		UploadID:              uploadID,
		URLs:                  urls,
		ExpiresAt:             time.Now().Add(expiration).Unix(),
		MaxChunkRetry:         runtimeSnapshot.MaxChunkRetry,
		TransferParallelism:   runtimeSnapshot.TransferParallelism,
		CacheChunksForRetry:   runtimeSnapshot.CacheChunksForRetry,
		BlobSignedURLTTL:      int(expiration.Seconds()),
		BlobSignedURLReuseTTL: int(reuseTTL.Seconds()),
	}

	if s.redisMultipart != nil {
		_ = s.redisMultipart.SetPresignedURLCache(ctx, uploadID, &redis.PresignedURLCacheEntry{
			UploadID:  result.UploadID,
			URLs:      result.URLs,
			ExpiresAt: result.ExpiresAt,
		}, expiration)
	}

	return result, nil
}

func (s *multipartService) RecordChunkUpload(ctx context.Context, uploadID string, userID uint, chunkNumber int, etag string, options ...ChunkRecordOptions) error {
	upload, err := s.multipartRepo.GetByUploadID(uploadID)
	if err != nil {
		return fmt.Errorf("upload task not found: %w", err)
	}
	if upload.UserID != userID {
		return fmt.Errorf("no permission to access this upload task")
	}
	if upload.Status != model.MultipartStatusUploading {
		return fmt.Errorf("invalid upload task status: %s", upload.Status)
	}
	if chunkNumber < 1 || chunkNumber > upload.TotalChunks {
		return fmt.Errorf("invalid chunk number: %d", chunkNumber)
	}
	runtimeSnapshot := s.uploadRuntimeSnapshotForSession(ctx, uploadID)
	recordOptions := ChunkRecordOptions{}
	if len(options) > 0 {
		recordOptions = options[0]
	}
	if recordOptions.ActiveTransfers > runtimeSnapshot.TransferParallelism {
		return fmt.Errorf("active transfer count exceeds server limit: max %d", runtimeSnapshot.TransferParallelism)
	}
	if recordOptions.Attempt > runtimeSnapshot.MaxChunkRetry+1 {
		return fmt.Errorf("chunk retry count exceeds server limit: max %d retries", runtimeSnapshot.MaxChunkRetry)
	}
	if !runtimeSnapshot.CacheChunksForRetry && recordOptions.Attempt > 1 {
		completed, completedErr := s.redisMultipart.IsChunkCompleted(ctx, uploadID, chunkNumber)
		if completedErr != nil {
			return fmt.Errorf("check completed chunk failed: %w", completedErr)
		}
		if completed {
			return fmt.Errorf("chunk retry cache is disabled; chunk %d is already completed", chunkNumber)
		}
	}

	if err := s.redisMultipart.AddCompletedChunk(ctx, uploadID, chunkNumber); err != nil {
		return fmt.Errorf("record chunk failed: %w", err)
	}

	return nil
}

func (s *multipartService) GetCompletedChunks(ctx context.Context, uploadID string, userID uint) ([]int, error) {
	upload, err := s.multipartRepo.GetByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload task not found: %w", err)
	}
	if upload.UserID != userID {
		return nil, fmt.Errorf("no permission to access this upload task")
	}

	chunks, err := s.redisMultipart.GetCompletedChunks(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("get completed chunks failed: %w", err)
	}

	result := make([]int, 0, len(chunks))
	for _, chunk := range chunks {
		var chunkNum int
		fmt.Sscanf(chunk, "%d", &chunkNum)
		result = append(result, chunkNum)
	}

	return result, nil
}

func (s *multipartService) CompleteMultipartUpload(ctx context.Context, uploadID string, userID uint, parentID *uint) (*model.UserFile, error) {
	upload, err := s.multipartRepo.GetByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload task not found: %w", err)
	}
	if upload.UserID != userID {
		return nil, fmt.Errorf("no permission to access this upload task")
	}
	if upload.Status != model.MultipartStatusUploading {
		return nil, fmt.Errorf("invalid upload task status: %s", upload.Status)
	}
	if parentID == nil && s.redisMultipart != nil {
		if metadata, metaErr := s.redisMultipart.GetUploadMetadata(ctx, uploadID); metaErr == nil {
			if rawParentID := metadata["parent_id"]; rawParentID != "" {
				if parsedParentID, parseErr := strconv.ParseUint(rawParentID, 10, 32); parseErr == nil {
					value := uint(parsedParentID)
					parentID = &value
				}
			}
		}
	}

	completedCount, err := s.redisMultipart.GetCompletedChunkCount(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("get completed chunk count failed: %w", err)
	}
	if int(completedCount) != upload.TotalChunks {
		return nil, fmt.Errorf("chunks not fully uploaded: %d/%d", completedCount, upload.TotalChunks)
	}

	var userFile *model.UserFile
	var queuedPhysicalFile *model.PhysicalFile
	policyRuntime := newStoragePolicyRuntime(s.db, userID)
	encryptionConfig, err := policyRuntime.EncryptionConfig()
	if err != nil {
		return nil, err
	}
	encryptionConfig, err = prepareBlobEncryptionConfig(encryptionConfig, s.masterKeys)
	if err != nil {
		return nil, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := reserveUserCapacity(tx, userID, upload.FileSize); err != nil {
			return err
		}

		chunkPaths := make([]string, upload.TotalChunks)
		for i := 1; i <= upload.TotalChunks; i++ {
			chunkPaths[i-1] = fmt.Sprintf("multipart/%s/chunk_%d", uploadID, i)
		}

		var physicalFile *model.PhysicalFile
		var existingPhysical model.PhysicalFile
		lookupErr := tx.Where("file_hash = ?", upload.FileHash).First(&existingPhysical).Error
		if lookupErr == nil {
			physicalFile = &existingPhysical
		} else if lookupErr != gorm.ErrRecordNotFound {
			return fmt.Errorf("query physical file failed: %w", lookupErr)
		}

		finalStoragePath := ""
		if physicalFile == nil {
			fallbackStoragePath := fmt.Sprintf("files/%s/%s", upload.FileHash[:2], upload.FileHash)
			storagePath, pathErr := policyRuntime.RenderBlobStoragePath(StoragePolicyBlobPathContext{
				FileName:     upload.FileName,
				FileHash:     upload.FileHash,
				ParentID:     parentID,
				UserFileRepo: s.userFileRepo,
				FallbackPath: fallbackStoragePath,
				StorageExists: func(path string) bool {
					return s.storage != nil && s.storage.Exists(path)
				},
			})
			if pathErr != nil {
				return pathErr
			}
			if encryptionConfig != nil && encryptionConfig.Enabled {
				plainComposePath := fmt.Sprintf("multipart/%s-composed-plain", uploadID)
				if err := s.storage.ComposeChunks(plainComposePath, chunkPaths); err != nil {
					return fmt.Errorf("compose chunks failed: %w", err)
				}
				plainReader, err := s.storage.Read(plainComposePath)
				if err != nil {
					_ = s.storage.Delete(plainComposePath)
					return fmt.Errorf("read composed chunks failed: %w", err)
				}
				if err := maybePreAllocateMultipartBlob(policyRuntime, s.storage, storagePath, physicalBlobStoredSize(upload.FileSize, encryptionConfig)); err != nil {
					_ = plainReader.Close()
					_ = s.storage.Delete(plainComposePath)
					return err
				}
				saveErr := savePhysicalBlob(s.storage, plainReader, storagePath, encryptionConfig, s.masterKeys)
				closeErr := plainReader.Close()
				_ = s.storage.Delete(plainComposePath)
				if saveErr != nil {
					_ = s.storage.Delete(storagePath)
					return fmt.Errorf("save encrypted multipart blob failed: %w", saveErr)
				}
				if closeErr != nil {
					return fmt.Errorf("close composed chunks failed: %w", closeErr)
				}
			} else {
				if err := maybePreAllocateMultipartBlob(policyRuntime, s.storage, storagePath, upload.FileSize); err != nil {
					return err
				}
				if err := s.storage.ComposeChunks(storagePath, chunkPaths); err != nil {
					_ = s.storage.Delete(storagePath)
					return fmt.Errorf("compose chunks failed: %w", err)
				}
			}
			finalStoragePath = storagePath
			physicalFile = &model.PhysicalFile{
				FileHash:        upload.FileHash,
				FileSize:        upload.FileSize,
				StoragePath:     storagePath,
				RefCount:        0,
				StorageType:     upload.StorageType,
				ContentType:     s.resolveContentTypeFromSettings(upload.FileName),
				ChunkCount:      upload.TotalChunks,
				IsMultipart:     true,
				Encrypted:       encryptionConfig != nil && encryptionConfig.Enabled,
				EncryptionKeyID: encryptionKeyID(encryptionConfig),
			}
			if err := tx.Create(physicalFile).Error; err != nil {
				return fmt.Errorf("create physical file failed: %w", err)
			}
		} else {
			finalStoragePath = physicalFile.StoragePath
		}
		if physicalFile != nil && (strings.TrimSpace(physicalFile.ContentType) == "" || physicalFile.ContentType == "application/octet-stream") {
			contentType := s.resolveContentTypeFromSettings(upload.FileName)
			if contentType != "application/octet-stream" {
				if err := tx.Model(&model.PhysicalFile{}).
					Where("id = ?", physicalFile.ID).
					Update("content_type", contentType).Error; err != nil {
					return fmt.Errorf("update physical file content type failed: %w", err)
				}
				physicalFile.ContentType = contentType
			}
		}

		queuedPhysicalFile = physicalFile

		if err := tx.Model(&model.PhysicalFile{}).
			Where("id = ?", physicalFile.ID).
			UpdateColumn("ref_count", gorm.Expr("ref_count + ?", 1)).Error; err != nil {
			return fmt.Errorf("increment ref count failed: %w", err)
		}

		if parentID != nil {
			var parentFolder model.UserFile
			if folderErr := tx.First(&parentFolder, *parentID).Error; folderErr != nil {
				return fmt.Errorf("parent folder not found: %w", folderErr)
			}
			if parentFolder.UserID != userID {
				return fmt.Errorf("no permission to access parent folder")
			}
			if !parentFolder.IsFolder {
				return fmt.Errorf("parent node is not a folder")
			}
		}

		userFile = &model.UserFile{
			UserID:         userID,
			ParentID:       parentID,
			FileName:       upload.FileName,
			PhysicalFileID: &physicalFile.ID,
			IsFolder:       false,
			FileSize:       upload.FileSize,
		}

		if err := tx.Create(userFile).Error; err != nil {
			return fmt.Errorf("create user file failed: %w", err)
		}

		now := time.Now()
		upload.Status = model.MultipartStatusCompleted
		upload.CompletedAt = &now
		upload.StoragePath = finalStoragePath
		if err := tx.Model(&model.MultipartUpload{}).
			Where("upload_id = ?", uploadID).
			Updates(map[string]interface{}{
				"status":       model.MultipartStatusCompleted,
				"completed_at": now,
				"storage_path": finalStoragePath,
			}).Error; err != nil {
			return fmt.Errorf("update upload task status failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.redisMultipart.DeleteUploadData(ctx, uploadID); err != nil {
		fmt.Printf("cleanup upload redis data failed: %v\n", err)
	}

	go func() {
		chunkPaths := make([]string, upload.TotalChunks)
		for i := 1; i <= upload.TotalChunks; i++ {
			chunkPaths[i-1] = fmt.Sprintf("multipart/%s/chunk_%d", uploadID, i)
		}
		if err := s.storage.DeleteChunks(chunkPaths); err != nil {
			fmt.Printf("delete chunks failed: %v\n", err)
		}
	}()

	if s.queueDispatch != nil && userFile != nil && queuedPhysicalFile != nil {
		if err := s.queueDispatch.EnqueueFilePostProcess(
			userFile.ID,
			queuedPhysicalFile.ID,
			userFile.FileName,
			queuedPhysicalFile.StoragePath,
			queuedPhysicalFile.StorageType,
		); err != nil {
			logger.Warn("enqueue multipart file post process jobs failed", zap.Uint("user_file_id", userFile.ID), zap.Error(err))
		}
	}

	recordTrafficEvent(s.db, userID, "upload", upload.FileSize, "multipart", "user_file", fmt.Sprintf("%d", userFile.ID))
	recordStoragePolicyHit(s.db, StoragePolicyHitLogInput{
		Action:       "multipart_complete",
		UserID:       userID,
		FileID:       userFile.ID,
		FileName:     userFile.FileName,
		FileSize:     upload.FileSize,
		ResourceType: "multipart_upload",
		ResourceID:   uploadID,
		Fallback: defaultStoragePolicyHitFallbackConfig(map[string]interface{}{
			"chunk_size":   upload.ChunkSize,
			"total_chunks": upload.TotalChunks,
		}),
	})

	return userFile, nil
}

func (s *multipartService) CancelMultipartUpload(ctx context.Context, uploadID string, userID uint) error {
	upload, err := s.multipartRepo.GetByUploadID(uploadID)
	if err != nil {
		return fmt.Errorf("upload task not found: %w", err)
	}
	if upload.UserID != userID {
		return fmt.Errorf("no permission to access this upload task")
	}

	if err := s.multipartRepo.UpdateStatus(uploadID, model.MultipartStatusCancelled); err != nil {
		return fmt.Errorf("update upload task status failed: %w", err)
	}

	if err := s.redisMultipart.DeleteUploadData(ctx, uploadID); err != nil {
		fmt.Printf("cleanup upload redis data failed: %v\n", err)
	}

	go func() {
		chunkPaths := make([]string, upload.TotalChunks)
		for i := 1; i <= upload.TotalChunks; i++ {
			chunkPaths[i-1] = fmt.Sprintf("multipart/%s/chunk_%d", uploadID, i)
		}
		if err := s.storage.DeleteChunks(chunkPaths); err != nil {
			fmt.Printf("delete chunks failed: %v\n", err)
		}
	}()

	return nil
}

func (s *multipartService) GetUploadProgress(ctx context.Context, uploadID string, userID uint) (*UploadProgress, error) {
	upload, err := s.multipartRepo.GetByUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("upload task not found: %w", err)
	}
	if upload.UserID != userID {
		return nil, fmt.Errorf("no permission to access this upload task")
	}

	completedCount, err := s.redisMultipart.GetCompletedChunkCount(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("get completed chunk count failed: %w", err)
	}

	progress := 0.0
	if upload.TotalChunks > 0 {
		progress = float64(completedCount) / float64(upload.TotalChunks) * 100
	}
	runtimeSnapshot := s.uploadRuntimeSnapshotForSession(ctx, uploadID)

	return &UploadProgress{
		UploadID:            uploadID,
		TotalChunks:         upload.TotalChunks,
		CompletedChunks:     int(completedCount),
		Progress:            progress,
		Status:              upload.Status,
		MaxChunkRetry:       runtimeSnapshot.MaxChunkRetry,
		TransferParallelism: runtimeSnapshot.TransferParallelism,
		CacheChunksForRetry: runtimeSnapshot.CacheChunksForRetry,
	}, nil
}

func (s *multipartService) ListUploadTasks(ctx context.Context, userID uint, status string, page, pageSize int) ([]UploadTaskSummary, int64, error) {
	uploads, total, err := s.multipartRepo.GetByUserID(userID, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list upload tasks failed: %w", err)
	}

	summaries := make([]UploadTaskSummary, 0, len(uploads))
	for _, upload := range uploads {
		completedChunks := 0
		progress := 0.0

		switch upload.Status {
		case model.MultipartStatusCompleted:
			completedChunks = upload.TotalChunks
			progress = 100
		case model.MultipartStatusCancelled:
			progress = 0
		default:
			count, countErr := s.redisMultipart.GetCompletedChunkCount(ctx, upload.UploadID)
			if countErr == nil {
				completedChunks = int(count)
				if upload.TotalChunks > 0 {
					progress = float64(completedChunks) / float64(upload.TotalChunks) * 100
				}
			}
		}

		summaries = append(summaries, UploadTaskSummary{
			UploadID:        upload.UploadID,
			FileName:        upload.FileName,
			FileSize:        upload.FileSize,
			TotalChunks:     upload.TotalChunks,
			CompletedChunks: completedChunks,
			Progress:        progress,
			Status:          upload.Status,
			StorageType:     upload.StorageType,
			CreatedAt:       upload.CreatedAt,
			CompletedAt:     upload.CompletedAt,
		})
	}

	return summaries, total, nil
}

func (s *multipartService) resolveContentTypeFromSettings(fileName string) string {
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

func (s *multipartService) getUploadSessionTTL() time.Duration {
	if s.settingsRepo != nil {
		if setting, err := s.settingsRepo.Get(); err == nil && setting != nil && setting.UploadSessionTTL > 0 {
			return time.Duration(setting.UploadSessionTTL) * time.Second
		}
	}

	return time.Duration(s.multipartExpireHours) * time.Hour
}

func (s *multipartService) uploadRuntimeSnapshot() multipartRuntimeSnapshot {
	defaults := defaultFileSystemSettingPayload()
	snapshot := multipartRuntimeSnapshot{
		MaxChunkRetry:       defaults.MaxChunkRetry,
		TransferParallelism: defaults.TransferParallelism,
		CacheChunksForRetry: defaults.CacheChunksForRetry,
		UploadSessionTTL:    defaults.UploadSessionTTL,
	}
	if snapshot.UploadSessionTTL <= 0 {
		snapshot.UploadSessionTTL = int((time.Duration(s.multipartExpireHours) * time.Hour).Seconds())
	}
	if s.settingsRepo != nil {
		if setting, err := s.settingsRepo.Get(); err == nil && setting != nil {
			if setting.MaxChunkRetry >= 0 {
				snapshot.MaxChunkRetry = clampInt(setting.MaxChunkRetry, 0, 50)
			}
			if setting.TransferParallelism > 0 {
				snapshot.TransferParallelism = clampInt(setting.TransferParallelism, 1, 64)
			}
			snapshot.CacheChunksForRetry = setting.CacheChunksForRetry
			if setting.UploadSessionTTL > 0 {
				snapshot.UploadSessionTTL = clampInt(setting.UploadSessionTTL, 1, 2592000)
			}
		}
	}
	if snapshot.TransferParallelism <= 0 {
		snapshot.TransferParallelism = 1
	}
	return snapshot
}

func (s *multipartService) uploadRuntimeSnapshotForSession(ctx context.Context, uploadID string) multipartRuntimeSnapshot {
	snapshot := s.uploadRuntimeSnapshot()
	if s.redisMultipart == nil || strings.TrimSpace(uploadID) == "" {
		return snapshot
	}
	metadata, err := s.redisMultipart.GetUploadMetadata(ctx, uploadID)
	if err != nil || len(metadata) == 0 {
		return snapshot
	}
	if value, ok := parseMetadataInt(metadata, "max_chunk_retry"); ok {
		snapshot.MaxChunkRetry = clampInt(value, 0, 50)
	}
	if value, ok := parseMetadataInt(metadata, "transfer_parallelism"); ok && value > 0 {
		snapshot.TransferParallelism = clampInt(value, 1, 64)
	}
	if value, ok := parseMetadataBool(metadata, "cache_chunks_for_retry"); ok {
		snapshot.CacheChunksForRetry = value
	}
	if value, ok := parseMetadataInt(metadata, "upload_session_ttl"); ok && value > 0 {
		snapshot.UploadSessionTTL = clampInt(value, 1, 2592000)
	}
	return snapshot
}

func parseMetadataInt(metadata map[string]string, key string) (int, bool) {
	value := strings.TrimSpace(metadata[key])
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func parseMetadataBool(metadata map[string]string, key string) (bool, bool) {
	value := strings.TrimSpace(metadata[key])
	if value == "" {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return parsed, true
}

func (s *multipartService) getBlobSignedURLTTL() time.Duration {
	if s.settingsRepo != nil {
		if setting, err := s.settingsRepo.Get(); err == nil && setting != nil && setting.BlobSignedURLTTL > 0 {
			return time.Duration(setting.BlobSignedURLTTL) * time.Second
		}
	}

	return time.Hour
}

func (s *multipartService) getBlobSignedURLReuseTTL() time.Duration {
	if s.settingsRepo != nil {
		if setting, err := s.settingsRepo.Get(); err == nil && setting != nil && setting.BlobSignedURLReuseTTL > 0 {
			return time.Duration(setting.BlobSignedURLReuseTTL) * time.Second
		}
	}

	return 10 * time.Minute
}
