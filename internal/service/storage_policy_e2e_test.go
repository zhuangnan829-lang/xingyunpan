package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	sqlite "github.com/glebarez/sqlite"
	goredis "github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	xredis "xingyunpan-v2/pkg/redis"
	"xingyunpan-v2/pkg/storage"
)

type storagePolicyE2EEnv struct {
	db               *gorm.DB
	storageDir       string
	stor             *storage.LocalStorage
	redisServer      *miniredis.Miniredis
	redisClient      *goredis.Client
	policy           model.StoragePolicy
	group            model.UserGroup
	user             model.User
	fileService      FileService
	multipartService MultipartService
	shareService     ShareService
	policyService    StoragePolicyService
}

func newStoragePolicyE2EEnv(t *testing.T) *storagePolicyE2EEnv {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "policy-e2e.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.StoragePolicy{},
		&model.StoragePolicyAudit{},
		&model.StoragePolicyHitLog{},
		&model.UserGroup{},
		&model.User{},
		&model.UserFile{},
		&model.PhysicalFile{},
		&model.MultipartUpload{},
		&model.Share{},
		&model.ShareFile{},
		&model.TrafficEvent{},
		&model.FileSystemSetting{},
	); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	policy := model.StoragePolicy{
		Name:               "E2E storage policy",
		Type:               "local",
		BlobPath:           "tenant/{uid}/{path}",
		BlobNamePattern:    "blob-{uid}-{hash}-{originname}",
		MaxFileSize:        1,
		MaxFileSizeUnit:    "KB",
		ExtensionMode:      "allow",
		Extensions:         "txt",
		NameRuleMode:       "allow",
		NameRegex:          `^ok-[a-z0-9-]+\.txt$`,
		ChunkSize:          1,
		ChunkSizeUnit:      "KB",
		PreAllocate:        true,
		ParallelChunkCount: 3,
		EnableCDN:          false,
		EnableEncryption:   true,
		EncryptionKeyID:    "e2e-key",
	}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("seed policy: %v", err)
	}
	group := model.UserGroup{Name: "E2E Users", StoragePolicyID: policy.ID}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	user := model.User{
		Username:    "policy-e2e",
		Email:       "policy-e2e@example.com",
		Password:    "x",
		Role:        "user",
		Enabled:     true,
		UserGroupID: group.ID,
		Capacity:    1024 * 1024,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&model.FileSystemSetting{
		UploadSessionTTL:      3600,
		BlobSignedURLTTL:      3600,
		BlobSignedURLReuseTTL: 600,
		StaticCacheTTL:        0,
		MimeMap:               `{"txt":"text/plain"}`,
	}).Error; err != nil {
		t.Fatalf("seed file system settings: %v", err)
	}

	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(redisServer.Close)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	storageDir := t.TempDir()
	stor := storage.NewLocalStorage(storageDir)
	physicalRepo := repository.NewPhysicalFileRepository(db)
	userFileRepo := repository.NewUserFileRepository(db)
	userRepo := repository.NewUserRepository(db)
	settingsRepo := repository.NewFileSystemSettingRepository(db)
	multipartRepo := repository.NewMultipartUploadRepository(db)
	shareRepo := repository.NewShareRepository(db)
	policyRepo := repository.NewStoragePolicyRepository(db)

	return &storagePolicyE2EEnv{
		db:          db,
		storageDir:  storageDir,
		stor:        stor,
		redisServer: redisServer,
		redisClient: redisClient,
		policy:      policy,
		group:       group,
		user:        user,
		fileService: NewFileService(
			db,
			physicalRepo,
			userFileRepo,
			userRepo,
			settingsRepo,
			nil,
			stor,
			redisClient,
			nil,
			nil,
		),
		multipartService: NewMultipartService(
			db,
			multipartRepo,
			physicalRepo,
			userFileRepo,
			userRepo,
			settingsRepo,
			stor,
			xredis.NewMultipartRedis(redisClient),
			"local",
			5*1024*1024,
			24,
			nil,
		),
		shareService:  NewShareService(shareRepo, userFileRepo, redisClient, "test-secret", "http://localhost:3000", stor, db),
		policyService: NewStoragePolicyService(policyRepo, db),
	}
}

func TestStoragePolicyRealFlowE2E(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	ctx := context.Background()
	userID := env.user.ID

	folder, err := NewFolderService(env.db, repository.NewUserFileRepository(env.db), nil).Create(userID, "docs", nil)
	if err != nil {
		t.Fatalf("create folder for path variable: %v", err)
	}

	if _, err := env.fileService.Upload(userID, "ok-big.txt", 2*1024, strings.NewReader(strings.Repeat("x", 2*1024)), nil); !IsStoragePolicyValidationError(err) {
		t.Fatalf("file size limit did not block upload, err=%v", err)
	}
	if _, err := env.fileService.Upload(userID, "ok-image.jpg", 16, strings.NewReader("jpg"), nil); !IsStoragePolicyValidationError(err) {
		t.Fatalf("extension limit did not block upload, err=%v", err)
	}
	if _, err := env.fileService.CreateFile(userID, "bad-name.txt", []byte("bad"), nil); !IsStoragePolicyValidationError(err) {
		t.Fatalf("name regex did not block create file, err=%v", err)
	}

	encryptedFile, err := env.fileService.CreateFile(userID, "ok-secret.txt", []byte("secret payload"), &folder.ID)
	if err != nil {
		t.Fatalf("create encrypted file through upload flow: %v", err)
	}
	encryptedPhysical := mustLoadPhysicalFile(t, env.db, encryptedFile.PhysicalFileID)
	expectedHash := sha256Hex([]byte("secret payload"))
	if encryptedPhysical.FileHash != expectedHash {
		t.Fatalf("physical file hash = %s, want %s", encryptedPhysical.FileHash, expectedHash)
	}
	if !strings.HasPrefix(encryptedPhysical.StoragePath, "tenant/1/docs/") {
		t.Fatalf("blob directory policy was not applied: %s", encryptedPhysical.StoragePath)
	}
	if !strings.Contains(filepath.Base(encryptedPhysical.StoragePath), "ok-secret.txt") || !strings.Contains(encryptedPhysical.StoragePath, expectedHash) {
		t.Fatalf("blob name policy was not applied: %s", encryptedPhysical.StoragePath)
	}
	if !encryptedPhysical.Encrypted || encryptedPhysical.EncryptionKeyID != "e2e-key" {
		t.Fatalf("encryption metadata not stored: %#v", encryptedPhysical)
	}
	rawEncrypted, err := os.ReadFile(filepath.Join(env.storageDir, filepath.FromSlash(encryptedPhysical.StoragePath)))
	if err != nil {
		t.Fatalf("read encrypted blob: %v", err)
	}
	if bytes.Contains(rawEncrypted, []byte("secret payload")) {
		t.Fatal("encrypted blob was stored as plaintext")
	}
	reader, _, _, err := env.fileService.Download(userID, encryptedFile.ID)
	if err != nil {
		t.Fatalf("download encrypted file: %v", err)
	}
	downloaded, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil || string(downloaded) != "secret payload" {
		t.Fatalf("encrypted download did not decrypt: data=%q err=%v", string(downloaded), err)
	}
	if err := env.fileService.Rename(userID, encryptedFile.ID, "bad-rename.txt"); !IsStoragePolicyValidationError(err) {
		t.Fatalf("name regex did not block rename, err=%v", err)
	}
	if err := env.fileService.Rename(userID, encryptedFile.ID, "ok-renamed.txt"); err != nil {
		t.Fatalf("valid rename failed: %v", err)
	}

	policy := env.policy
	policy.MaxFileSize = 2
	policy.EnableEncryption = false
	policy.EnableCDN = true
	policy.DownloadCDN = "https://cdn.example.com/assets"
	if err := env.db.Save(&policy).Error; err != nil {
		t.Fatalf("enable cdn policy: %v", err)
	}

	cdnFile, err := env.fileService.CreateFile(userID, "ok-cdn.txt", []byte("cdn payload"), nil)
	if err != nil {
		t.Fatalf("create cdn file: %v", err)
	}
	cdnPhysical := mustLoadPhysicalFile(t, env.db, cdnFile.PhysicalFileID)
	result, err := env.fileService.DownloadWithDelivery(userID, cdnFile.ID, false)
	if err != nil {
		t.Fatalf("download with cdn: %v", err)
	}
	if result.RedirectURL == "" || !strings.HasPrefix(result.RedirectURL, "https://cdn.example.com/assets/") || !strings.Contains(result.RedirectURL, escapeStoragePathForURL(cdnPhysical.StoragePath)) {
		t.Fatalf("cdn download url was not generated from policy: %#v storage=%s", result, cdnPhysical.StoragePath)
	}
	preview, err := env.fileService.DownloadWithDelivery(userID, cdnFile.ID, true)
	if err != nil {
		t.Fatalf("preview with policy: %v", err)
	}
	if preview.RedirectURL != "" || preview.Reader == nil {
		t.Fatalf("preview should keep local stream, got %#v", preview)
	}
	_ = preview.Reader.Close()

	share, err := env.shareService.CreateShare(ctx, userID, []string{formatStoragePolicyFileResourceID(cdnFile.ID)}, nil, nil, nil)
	if err != nil {
		t.Fatalf("create share: %v", err)
	}
	shareDownload, err := env.shareService.DownloadShareWithDelivery(ctx, share.ShareToken)
	if err != nil {
		t.Fatalf("share download with cdn: %v", err)
	}
	if shareDownload.RedirectURL == "" || !strings.HasPrefix(shareDownload.RedirectURL, "https://cdn.example.com/assets/") {
		t.Fatalf("share download did not use policy cdn: %#v", shareDownload)
	}

	multipartData := []byte(strings.Repeat("m", 1500))
	multipartHash := sha256Hex(multipartData)
	initResult, err := env.multipartService.InitMultipartUpload(ctx, userID, "ok-multipart.txt", multipartHash, int64(len(multipartData)), 64*1024, nil)
	if err != nil {
		t.Fatalf("init multipart upload: %v", err)
	}
	if initResult.ChunkSize != 1024 || initResult.ParallelChunkCount != 3 || initResult.TotalChunks != 2 {
		t.Fatalf("multipart tuning did not use policy: %#v", initResult)
	}
	if err := env.stor.Save(bytes.NewReader(multipartData[:1024]), "multipart/"+initResult.UploadID+"/chunk_1"); err != nil {
		t.Fatalf("save chunk 1: %v", err)
	}
	if err := env.stor.Save(bytes.NewReader(multipartData[1024:]), "multipart/"+initResult.UploadID+"/chunk_2"); err != nil {
		t.Fatalf("save chunk 2: %v", err)
	}
	if err := env.multipartService.RecordChunkUpload(ctx, initResult.UploadID, userID, 1, "etag-1"); err != nil {
		t.Fatalf("record chunk 1: %v", err)
	}
	if err := env.multipartService.RecordChunkUpload(ctx, initResult.UploadID, userID, 2, "etag-2"); err != nil {
		t.Fatalf("record chunk 2: %v", err)
	}
	completed, err := env.multipartService.CompleteMultipartUpload(ctx, initResult.UploadID, userID, nil)
	if err != nil {
		t.Fatalf("complete multipart upload: %v", err)
	}
	completedPhysical := mustLoadPhysicalFile(t, env.db, completed.PhysicalFileID)
	if !completedPhysical.IsMultipart || completedPhysical.ChunkCount != 2 {
		t.Fatalf("multipart metadata missing: %#v", completedPhysical)
	}
	if !strings.HasPrefix(completedPhysical.StoragePath, "tenant/1/") || !strings.Contains(completedPhysical.StoragePath, "ok-multipart.txt") {
		t.Fatalf("multipart blob path/name policy was not applied: %s", completedPhysical.StoragePath)
	}

	assertEventuallyHitActions(t, env.db, env.policy.ID, []string{
		"upload", "download", "preview", "share_download", "multipart_init", "multipart_complete",
	})
}

func TestEncryptedUploadDownloadUsesConfiguredMasterKeyStorage(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	userID := env.user.ID
	settingsRepo := repository.NewFileSystemSettingRepository(env.db)
	resolver := NewFileSystemMasterKeyResolver(env.db, settingsRepo)

	fileKeyPath := filepath.Join(t.TempDir(), "master.key")
	if err := os.WriteFile(fileKeyPath, []byte("file-master-key"), 0600); err != nil {
		t.Fatalf("write master key file: %v", err)
	}
	t.Setenv("XINGYUNPAN_MASTER_KEY", "env-master-key")
	t.Setenv("XINGYUNPAN_MASTER_KEY_FILE", fileKeyPath)

	cases := []struct {
		mode           string
		expectedSource string
		content        string
	}{
		{mode: "database", expectedSource: "database:file_system_secrets.master_key", content: "database encrypted payload"},
		{mode: "env", expectedSource: "env:XINGYUNPAN_MASTER_KEY", content: "env encrypted payload"},
		{mode: "file", expectedSource: "file:" + fileKeyPath, content: "file encrypted payload"},
	}

	for _, tc := range cases {
		t.Run(tc.mode, func(t *testing.T) {
			setMasterKeyStorage(t, env.db, tc.mode)
			status := resolver.Status()
			if status == nil || !status.Available || status.Source != tc.expectedSource || status.Fingerprint == "" {
				t.Fatalf("master key status = %#v, want available source %s with fingerprint", status, tc.expectedSource)
			}

			file, err := env.fileService.CreateFile(userID, "ok-"+tc.mode+"-key.txt", []byte(tc.content), nil)
			if err != nil {
				t.Fatalf("create encrypted file with %s key: %v", tc.mode, err)
			}
			physical := mustLoadPhysicalFile(t, env.db, file.PhysicalFileID)
			if !physical.Encrypted {
				t.Fatalf("physical file is not encrypted under %s mode", tc.mode)
			}
			raw, err := os.ReadFile(filepath.Join(env.storageDir, filepath.FromSlash(physical.StoragePath)))
			if err != nil {
				t.Fatalf("read raw encrypted blob: %v", err)
			}
			if bytes.Contains(raw, []byte(tc.content)) {
				t.Fatalf("raw blob contains plaintext under %s mode", tc.mode)
			}

			reader, _, _, err := env.fileService.Download(userID, file.ID)
			if err != nil {
				t.Fatalf("download encrypted file with %s key: %v", tc.mode, err)
			}
			downloaded, err := io.ReadAll(reader)
			_ = reader.Close()
			if err != nil {
				t.Fatalf("read downloaded file: %v", err)
			}
			if string(downloaded) != tc.content {
				t.Fatalf("downloaded content = %q, want %q", string(downloaded), tc.content)
			}
		})
	}
}

func TestMissingEnvOrFileMasterKeyFailsClearly(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	t.Setenv("XINGYUNPAN_MASTER_KEY", "")
	t.Setenv("XINGYUNPAN_MASTER_KEY_FILE", filepath.Join(t.TempDir(), "missing.key"))

	setMasterKeyStorage(t, env.db, "env")
	if _, err := env.fileService.CreateFile(env.user.ID, "ok-missing-env.txt", []byte("payload"), nil); err == nil || !strings.Contains(err.Error(), "XINGYUNPAN_MASTER_KEY is empty") {
		t.Fatalf("env missing upload err = %v, want explicit env error", err)
	}

	setMasterKeyStorage(t, env.db, "file")
	if _, err := env.fileService.CreateFile(env.user.ID, "ok-missing-file.txt", []byte("payload"), nil); err == nil || !strings.Contains(err.Error(), "master key file") {
		t.Fatalf("file missing upload err = %v, want explicit file error", err)
	}
}

func TestMultipartUploadRuntimeSnapshotAndServerLimits(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	ctx := context.Background()
	updateFileSystemSettingFields(t, env.db, map[string]interface{}{
		"max_chunk_retry":        2,
		"cache_chunks_for_retry": false,
		"transfer_parallelism":   3,
		"upload_session_ttl":     90,
	})

	payload := []byte("runtime snapshot payload")
	initResult, err := env.multipartService.InitMultipartUpload(ctx, env.user.ID, "ok-runtime.txt", sha256Hex(payload), int64(len(payload)), 8, nil)
	if err != nil {
		t.Fatalf("init multipart upload: %v", err)
	}
	if initResult.MaxChunkRetry != 2 || initResult.TransferParallelism != 3 || initResult.CacheChunksForRetry {
		t.Fatalf("init runtime snapshot = retry:%d parallel:%d cache:%v", initResult.MaxChunkRetry, initResult.TransferParallelism, initResult.CacheChunksForRetry)
	}
	if initResult.ParallelChunkCount != 3 {
		t.Fatalf("parallel chunk count = %d, want 3", initResult.ParallelChunkCount)
	}

	updateFileSystemSettingFields(t, env.db, map[string]interface{}{
		"max_chunk_retry":        9,
		"cache_chunks_for_retry": true,
		"transfer_parallelism":   8,
	})
	urls, err := env.multipartService.GetPresignedURLs(ctx, initResult.UploadID, env.user.ID)
	if err != nil {
		t.Fatalf("get presigned urls: %v", err)
	}
	if urls.MaxChunkRetry != 2 || urls.TransferParallelism != 3 || urls.CacheChunksForRetry {
		t.Fatalf("presigned runtime snapshot changed after settings update: %#v", urls)
	}

	if err := env.multipartService.RecordChunkUpload(ctx, initResult.UploadID, env.user.ID, 1, "etag", ChunkRecordOptions{ActiveTransfers: 4}); err == nil || !strings.Contains(err.Error(), "max 3") {
		t.Fatalf("active transfer limit err = %v, want max 3", err)
	}
	if err := env.multipartService.RecordChunkUpload(ctx, initResult.UploadID, env.user.ID, 1, "etag", ChunkRecordOptions{Attempt: 4}); err == nil || !strings.Contains(err.Error(), "max 2 retries") {
		t.Fatalf("retry limit err = %v, want max 2 retries", err)
	}
	if err := env.multipartService.RecordChunkUpload(ctx, initResult.UploadID, env.user.ID, 1, "etag", ChunkRecordOptions{Attempt: 1, ActiveTransfers: 3}); err != nil {
		t.Fatalf("record chunk within limits: %v", err)
	}
	if err := env.multipartService.RecordChunkUpload(ctx, initResult.UploadID, env.user.ID, 1, "etag", ChunkRecordOptions{Attempt: 2, ActiveTransfers: 1}); err == nil || !strings.Contains(err.Error(), "retry cache is disabled") {
		t.Fatalf("duplicate chunk with retry cache disabled err = %v", err)
	}
	progress, err := env.multipartService.GetUploadProgress(ctx, initResult.UploadID, env.user.ID)
	if err != nil {
		t.Fatalf("get upload progress: %v", err)
	}
	if progress.MaxChunkRetry != 2 || progress.TransferParallelism != 3 || progress.CacheChunksForRetry {
		t.Fatalf("progress runtime snapshot changed after settings update: %#v", progress)
	}
}

func TestMultipartUploadTTLAndPresignedURLCacheFollowSettings(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	ctx := context.Background()

	updateFileSystemSettingFields(t, env.db, map[string]interface{}{"upload_session_ttl": 30})
	firstInit, err := env.multipartService.InitMultipartUpload(ctx, env.user.ID, "ok-ttl-a.txt", sha256Hex([]byte("ttl-a")), 8, 4, nil)
	if err != nil {
		t.Fatalf("first init: %v", err)
	}
	updateFileSystemSettingFields(t, env.db, map[string]interface{}{"upload_session_ttl": 120})
	secondInit, err := env.multipartService.InitMultipartUpload(ctx, env.user.ID, "ok-ttl-b.txt", sha256Hex([]byte("ttl-b")), 8, 4, nil)
	if err != nil {
		t.Fatalf("second init: %v", err)
	}
	if firstInit.UploadSessionTTL != 30 || secondInit.UploadSessionTTL != 120 {
		t.Fatalf("upload session ttl response = %d/%d, want 30/120", firstInit.UploadSessionTTL, secondInit.UploadSessionTTL)
	}
	if secondInit.ExpiresAt-firstInit.ExpiresAt < 60 {
		t.Fatalf("expires_at did not reflect updated upload_session_ttl: first=%d second=%d", firstInit.ExpiresAt, secondInit.ExpiresAt)
	}

	updateFileSystemSettingFields(t, env.db, map[string]interface{}{
		"blob_signed_url_ttl":       20,
		"blob_signed_url_reuse_ttl": 5,
	})
	firstURLs, err := env.multipartService.GetPresignedURLs(ctx, secondInit.UploadID, env.user.ID)
	if err != nil {
		t.Fatalf("first presigned urls: %v", err)
	}
	if firstURLs.BlobSignedURLTTL != 20 || firstURLs.BlobSignedURLReuseTTL != 5 {
		t.Fatalf("first url ttl response = %d/%d, want 20/5", firstURLs.BlobSignedURLTTL, firstURLs.BlobSignedURLReuseTTL)
	}

	updateFileSystemSettingFields(t, env.db, map[string]interface{}{
		"blob_signed_url_ttl":       80,
		"blob_signed_url_reuse_ttl": 1,
	})
	reusedURLs, err := env.multipartService.GetPresignedURLs(ctx, secondInit.UploadID, env.user.ID)
	if err != nil {
		t.Fatalf("reused presigned urls: %v", err)
	}
	if reusedURLs.ExpiresAt != firstURLs.ExpiresAt {
		t.Fatalf("cached url should be reused when remaining ttl exceeds reuse threshold: first=%d reused=%d", firstURLs.ExpiresAt, reusedURLs.ExpiresAt)
	}
	if reusedURLs.BlobSignedURLTTL != 80 || reusedURLs.BlobSignedURLReuseTTL != 1 {
		t.Fatalf("reused url ttl response should expose current settings: %d/%d", reusedURLs.BlobSignedURLTTL, reusedURLs.BlobSignedURLReuseTTL)
	}

	time.Sleep(time.Second)
	updateFileSystemSettingFields(t, env.db, map[string]interface{}{
		"blob_signed_url_ttl":       80,
		"blob_signed_url_reuse_ttl": 1000,
	})
	refreshedURLs, err := env.multipartService.GetPresignedURLs(ctx, secondInit.UploadID, env.user.ID)
	if err != nil {
		t.Fatalf("refreshed presigned urls: %v", err)
	}
	if refreshedURLs.ExpiresAt <= firstURLs.ExpiresAt {
		t.Fatalf("cached url should refresh when remaining ttl is below reuse threshold: first=%d refreshed=%d", firstURLs.ExpiresAt, refreshedURLs.ExpiresAt)
	}
}

func TestUserGroupStoragePolicySwitchControlsFileRules(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	userID := env.user.ID

	policyA := model.StoragePolicy{
		Name:               "Policy A txt only",
		Type:               "local",
		BlobPath:           "policy-a/{uid}",
		BlobNamePattern:    "{hash}-{originname}",
		MaxFileSize:        1,
		MaxFileSizeUnit:    "MB",
		ExtensionMode:      "allow",
		Extensions:         ".txt",
		ChunkSize:          1,
		ChunkSizeUnit:      "MB",
		ParallelChunkCount: 1,
	}
	policyB := model.StoragePolicy{
		Name:               "Policy B jpg only",
		Type:               "local",
		BlobPath:           "policy-b/{uid}",
		BlobNamePattern:    "{hash}-{originname}",
		MaxFileSize:        10,
		MaxFileSizeUnit:    "MB",
		ExtensionMode:      "allow",
		Extensions:         ".jpg",
		ChunkSize:          1,
		ChunkSizeUnit:      "MB",
		ParallelChunkCount: 1,
	}
	if err := env.db.Create(&policyA).Error; err != nil {
		t.Fatalf("seed policy A: %v", err)
	}
	if err := env.db.Create(&policyB).Error; err != nil {
		t.Fatalf("seed policy B: %v", err)
	}
	if err := env.db.Model(&model.UserGroup{}).
		Where("id = ?", env.group.ID).
		Update("storage_policy_id", policyA.ID).Error; err != nil {
		t.Fatalf("bind policy A to group: %v", err)
	}

	txtFile, err := env.fileService.Upload(userID, "note.txt", 16, strings.NewReader("hello policy a"), nil)
	if err != nil {
		t.Fatalf("policy A should allow small txt upload: %v", err)
	}
	txtPhysical := mustLoadPhysicalFile(t, env.db, txtFile.PhysicalFileID)
	if !strings.HasPrefix(txtPhysical.StoragePath, "policy-a/") {
		t.Fatalf("policy A blob path was not applied: %s", txtPhysical.StoragePath)
	}
	if _, err := env.fileService.Upload(userID, "photo.jpg", 16, strings.NewReader("jpg"), nil); !IsStoragePolicyValidationError(err) {
		t.Fatalf("policy A should reject jpg upload, err=%v", err)
	}
	oversize := strings.NewReader(strings.Repeat("x", 2*1024*1024))
	if _, err := env.fileService.Upload(userID, "large.txt", 2*1024*1024, oversize, nil); !IsStoragePolicyValidationError(err) {
		t.Fatalf("policy A should reject 2MB txt upload, err=%v", err)
	}
	assertEventuallyPolicyHit(t, env.db, policyA.ID, "upload", userID)

	if err := env.db.Model(&model.UserGroup{}).
		Where("id = ?", env.group.ID).
		Update("storage_policy_id", policyB.ID).Error; err != nil {
		t.Fatalf("bind policy B to group: %v", err)
	}
	jpgFile, err := env.fileService.Upload(userID, "photo.jpg", 16, strings.NewReader("jpg"), nil)
	if err != nil {
		t.Fatalf("policy B should allow jpg upload after group switch: %v", err)
	}
	jpgPhysical := mustLoadPhysicalFile(t, env.db, jpgFile.PhysicalFileID)
	if !strings.HasPrefix(jpgPhysical.StoragePath, "policy-b/") {
		t.Fatalf("policy B blob path was not applied: %s", jpgPhysical.StoragePath)
	}
	if err := env.fileService.Rename(userID, jpgFile.ID, "photo.txt"); !IsStoragePolicyValidationError(err) {
		t.Fatalf("policy B should reject rename to txt, err=%v", err)
	}
	assertEventuallyPolicyHit(t, env.db, policyB.ID, "upload", userID)

	if err := env.db.Model(&model.UserGroup{}).
		Where("id = ?", env.group.ID).
		Update("storage_policy_id", uint(999999)).Error; err != nil {
		t.Fatalf("bind missing policy to group: %v", err)
	}
	if _, err := env.fileService.Upload(userID, "broken.jpg", 16, strings.NewReader("jpg"), nil); err == nil || !strings.Contains(err.Error(), "storage policy 999999") {
		t.Fatalf("missing group policy should return clear error, err=%v", err)
	}
}

func TestStoragePolicyAuditRollbackAndDeleteGuardE2E(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	sparePolicy := model.StoragePolicy{
		Name:            "Spare policy",
		Type:            "local",
		BlobPath:        "files/{uid}",
		BlobNamePattern: "{hash}",
	}
	if err := env.db.Create(&sparePolicy).Error; err != nil {
		t.Fatalf("seed spare policy: %v", err)
	}

	payload, err := env.policyService.Get(env.policy.ID)
	if err != nil {
		t.Fatalf("get policy: %v", err)
	}
	payload.Name = "E2E storage policy updated"
	payload.MaxFileSize = 2
	if _, err := env.policyService.Update(env.policy.ID, payload, StoragePolicyActor{ID: 7, Name: "tester"}); err != nil {
		t.Fatalf("update policy: %v", err)
	}
	payload, err = env.policyService.Get(env.policy.ID)
	if err != nil {
		t.Fatalf("get updated policy: %v", err)
	}
	payload.MaxFileSize = 3
	if _, err := env.policyService.Update(env.policy.ID, payload, StoragePolicyActor{ID: 7, Name: "tester"}); err != nil {
		t.Fatalf("second update policy: %v", err)
	}

	audits, err := env.policyService.History(env.policy.ID, 20)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(audits) == 0 {
		t.Fatal("expected audit records after update")
	}
	var updateAudit *StoragePolicyAuditPayload
	for i := range audits {
		if audits[i].Action == "update" && audits[i].After != nil && audits[i].After.MaxFileSize == 2 {
			updateAudit = &audits[i]
			break
		}
	}
	if updateAudit == nil || updateAudit.Before == nil || updateAudit.After == nil {
		t.Fatalf("update audit did not include before/after snapshots: %#v", audits)
	}
	if updateAudit.Before.MaxFileSize != 1 || updateAudit.After.MaxFileSize != 2 {
		t.Fatalf("audit diff source values are wrong: %#v", updateAudit)
	}
	if updateAudit.UserCount != 1 || len(updateAudit.Groups) != 1 || updateAudit.Groups[0].ID != env.group.ID {
		t.Fatalf("audit did not capture affected groups/users: %#v", updateAudit)
	}

	rolledBack, err := env.policyService.Rollback(env.policy.ID, updateAudit.ID, StoragePolicyActor{ID: 8, Name: "rollback-tester"})
	if err != nil {
		t.Fatalf("rollback policy: %v", err)
	}
	if rolledBack.MaxFileSize != 2 {
		t.Fatalf("rollback should restore selected audit snapshot from later value, got %d", rolledBack.MaxFileSize)
	}
	auditsAfterRollback, err := env.policyService.History(env.policy.ID, 20)
	if err != nil {
		t.Fatalf("history after rollback: %v", err)
	}
	foundRollback := false
	for _, audit := range auditsAfterRollback {
		if audit.Action == "rollback" && audit.SourceAuditID == updateAudit.ID {
			foundRollback = true
			break
		}
	}
	if !foundRollback {
		t.Fatalf("rollback audit with source id was not recorded: %#v", auditsAfterRollback)
	}

	err = env.policyService.Delete(env.policy.ID, StoragePolicyActor{ID: 9, Name: "delete-tester"})
	if err == nil || !strings.Contains(err.Error(), "bound to user groups") {
		t.Fatalf("delete should be blocked while user group is bound, err=%v", err)
	}
}

func TestStoragePolicyRepairLegacyDefaultsIsAuditedAndIdempotent(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	legacy := model.StoragePolicy{
		Name:               "榛樿瀛樺偍绛栫暐",
		Type:               "本地服务器集群",
		GroupsJSON:         "",
		BlobPath:           "",
		BlobNamePattern:    "",
		MaxFileSizeUnit:    "兆",
		ExtensionMode:      "允许",
		NameRuleMode:       "允许",
		ChunkSize:          0,
		ChunkSizeUnit:      "兆",
		ParallelChunkCount: 0,
	}
	if err := env.db.Create(&legacy).Error; err != nil {
		t.Fatalf("seed legacy policy: %v", err)
	}

	repaired, err := env.policyService.RepairLegacyDefaults(StoragePolicyActor{ID: 99, Name: "repair-admin"})
	if err != nil {
		t.Fatalf("repair legacy defaults: %v", err)
	}
	if len(repaired) != 1 || repaired[0].Action != "repair_legacy" {
		t.Fatalf("expected one repair audit, got %#v", repaired)
	}

	payload, err := env.policyService.Get(legacy.ID)
	if err != nil {
		t.Fatalf("get repaired policy: %v", err)
	}
	if payload.Name != "默认存储策略" || payload.Type != "local" {
		t.Fatalf("legacy name/type not repaired: %#v", payload)
	}
	if payload.BlobPath != "/cloudreve/data/uploads/{uid}/{path}" || payload.BlobNamePattern != "{uid}_{randomkey8}_{originname}" {
		t.Fatalf("legacy blob config not repaired: %#v", payload)
	}
	if payload.MaxFileSizeUnit != "MB" || payload.ExtensionMode != "allow" || payload.NameRuleMode != "allow" || payload.ChunkSize != 25 || payload.ParallelChunkCount != 1 {
		t.Fatalf("legacy abnormal config not repaired: %#v", payload)
	}
	if repaired[0].Before == nil || repaired[0].After == nil || repaired[0].Before.Name == repaired[0].After.Name {
		t.Fatalf("repair audit did not capture before/after diff: %#v", repaired[0])
	}

	secondRun, err := env.policyService.RepairLegacyDefaults(StoragePolicyActor{ID: 99, Name: "repair-admin"})
	if err != nil {
		t.Fatalf("second repair legacy defaults: %v", err)
	}
	if len(secondRun) != 0 {
		t.Fatalf("repair should be idempotent, got %#v", secondRun)
	}
	var repairAuditCount int64
	if err := env.db.Model(&model.StoragePolicyAudit{}).
		Where("storage_policy_id = ? AND action = ?", legacy.ID, "repair_legacy").
		Count(&repairAuditCount).Error; err != nil {
		t.Fatalf("count repair audits: %v", err)
	}
	if repairAuditCount != 1 {
		t.Fatalf("expected exactly one repair audit, got %d", repairAuditCount)
	}
}

func TestStoragePolicyImportExportCopyAuditE2E(t *testing.T) {
	env := newStoragePolicyE2EEnv(t)
	actor := StoragePolicyActor{ID: 12, Name: "import-admin"}

	copied, err := env.policyService.Copy(env.policy.ID, actor)
	if err != nil {
		t.Fatalf("copy policy: %v", err)
	}
	if copied.ID == 0 || copied.ID == env.policy.ID || !strings.Contains(copied.Name, "副本") {
		t.Fatalf("copy did not create a distinct policy: %#v", copied)
	}
	copyAudits, err := env.policyService.History(copied.ID, 5)
	if err != nil {
		t.Fatalf("copy history: %v", err)
	}
	if len(copyAudits) == 0 || copyAudits[0].Action != "copy" || copyAudits[0].After == nil {
		t.Fatalf("copy audit was not recorded with snapshot: %#v", copyAudits)
	}

	exported, err := env.policyService.Export(env.policy.ID)
	if err != nil {
		t.Fatalf("export policy: %v", err)
	}
	if exported.Name != env.policy.Name || exported.BlobPath != env.policy.BlobPath {
		t.Fatalf("exported payload mismatch: %#v", exported)
	}

	conflict := *exported
	conflict.ID = 0
	if _, err := env.policyService.Import(&conflict, 0, actor); err == nil || !strings.Contains(err.Error(), "user group binding conflict") {
		t.Fatalf("import should reject bound user group conflict, err=%v", err)
	}

	invalidPath := *exported
	invalidPath.Name = "Invalid path import"
	invalidPath.Groups = []string{"Detached Import Group"}
	invalidPath.BlobPath = "../bad/{hash}"
	if _, err := env.policyService.Import(&invalidPath, 0, actor); err == nil || !strings.Contains(err.Error(), "relative path traversal") {
		t.Fatalf("import should reject illegal blob path, err=%v", err)
	}

	invalidCDN := *exported
	invalidCDN.Name = "Invalid CDN import"
	invalidCDN.Groups = []string{"Detached Import Group"}
	invalidCDN.EnableCDN = true
	invalidCDN.DownloadCDN = "ftp://cdn.example.com"
	if _, err := env.policyService.Import(&invalidCDN, 0, actor); err == nil || !strings.Contains(err.Error(), "http or https") {
		t.Fatalf("import should reject invalid CDN config, err=%v", err)
	}

	invalidKey := *exported
	invalidKey.Name = "Invalid key import"
	invalidKey.Groups = []string{"Detached Import Group"}
	invalidKey.EnableEncryption = true
	invalidKey.EncryptionKeyID = "bad/key"
	if _, err := env.policyService.Import(&invalidKey, 0, actor); err == nil || !strings.Contains(err.Error(), "encryption key id") {
		t.Fatalf("import should reject invalid encryption key, err=%v", err)
	}

	valid := *exported
	valid.ID = 0
	valid.Name = "Imported policy"
	valid.Groups = []string{"Detached Import Group"}
	valid.EnableCDN = true
	valid.DownloadCDN = "https://cdn.import.example/assets"
	imported, err := env.policyService.Import(&valid, 0, actor)
	if err != nil {
		t.Fatalf("import valid policy: %v", err)
	}
	if imported.ID == 0 || imported.Name != "Imported policy" || imported.DownloadCDN != valid.DownloadCDN {
		t.Fatalf("valid import returned wrong payload: %#v", imported)
	}
	importAudits, err := env.policyService.History(imported.ID, 5)
	if err != nil {
		t.Fatalf("import history: %v", err)
	}
	if len(importAudits) == 0 || importAudits[0].Action != "import" || importAudits[0].After == nil {
		t.Fatalf("import audit was not recorded: %#v", importAudits)
	}

	overwrite := *imported
	overwrite.MaxFileSize = 8
	overwrite.DownloadCDN = "https://cdn.import.example/overwritten"
	overwritten, err := env.policyService.Import(&overwrite, copied.ID, actor)
	if err != nil {
		t.Fatalf("import overwrite policy: %v", err)
	}
	if overwritten.ID != copied.ID || overwritten.MaxFileSize != 8 || overwritten.DownloadCDN != overwrite.DownloadCDN {
		t.Fatalf("overwrite returned wrong payload: %#v", overwritten)
	}
	overwriteAudits, err := env.policyService.History(copied.ID, 10)
	if err != nil {
		t.Fatalf("overwrite history: %v", err)
	}
	foundOverwrite := false
	for _, audit := range overwriteAudits {
		if audit.Action == "import_overwrite" && audit.Before != nil && audit.After != nil && audit.After.MaxFileSize == 8 {
			foundOverwrite = true
			break
		}
	}
	if !foundOverwrite {
		t.Fatalf("import overwrite audit with diff was not recorded: %#v", overwriteAudits)
	}
}

func mustLoadPhysicalFile(t *testing.T, db *gorm.DB, id *uint) *model.PhysicalFile {
	t.Helper()
	if id == nil {
		t.Fatal("user file has no physical file id")
	}
	var physical model.PhysicalFile
	if err := db.First(&physical, *id).Error; err != nil {
		t.Fatalf("load physical file: %v", err)
	}
	return &physical
}

func setMasterKeyStorage(t *testing.T, db *gorm.DB, mode string) {
	t.Helper()
	if err := db.Model(&model.FileSystemSetting{}).Where("1 = 1").Update("master_key_storage", mode).Error; err != nil {
		t.Fatalf("set master key storage to %s: %v", mode, err)
	}
}

func updateFileSystemSettingFields(t *testing.T, db *gorm.DB, fields map[string]interface{}) {
	t.Helper()
	if err := db.Model(&model.FileSystemSetting{}).Where("1 = 1").Updates(fields).Error; err != nil {
		t.Fatalf("update file system settings %#v: %v", fields, err)
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func assertEventuallyHitActions(t *testing.T, db *gorm.DB, policyID uint, expected []string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	var lastCounts map[string]int
	for time.Now().Before(deadline) {
		lastCounts = map[string]int{}
		var rows []model.StoragePolicyHitLog
		if err := db.Where("storage_policy_id = ?", policyID).Find(&rows).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("query hit logs: %v", err)
		}
		for _, row := range rows {
			lastCounts[row.Action]++
			if row.HitType != "user_group_policy" {
				t.Fatalf("expected user_group_policy hit, got %#v", row)
			}
			if row.UserGroupName != "E2E Users" {
				t.Fatalf("hit log did not capture user group: %#v", row)
			}
			var cfg map[string]interface{}
			if err := json.Unmarshal([]byte(row.ConfigJSON), &cfg); err != nil {
				t.Fatalf("decode hit config: %v", err)
			}
			if cfg["chunk_size"] == nil || cfg["parallel_chunk_count"] == nil || cfg["enable_encryption"] == nil {
				t.Fatalf("hit config missed key policy fields: %#v", cfg)
			}
		}
		ok := true
		for _, action := range expected {
			if lastCounts[action] == 0 {
				ok = false
				break
			}
		}
		if ok {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("missing storage policy hit actions, got=%#v want=%#v", lastCounts, expected)
}

func assertEventuallyPolicyHit(t *testing.T, db *gorm.DB, policyID uint, action string, userID uint) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var count int64
		if err := db.Model(&model.StoragePolicyHitLog{}).
			Where("storage_policy_id = ? AND action = ? AND user_id = ?", policyID, action, userID).
			Count(&count).Error; err != nil {
			t.Fatalf("query policy hit log: %v", err)
		}
		if count > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("missing storage policy hit log: policy=%d action=%s user=%d", policyID, action, userID)
}
