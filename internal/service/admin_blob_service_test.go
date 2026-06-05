package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/storage"
)

type adminBlobTestEnv struct {
	db         *gorm.DB
	storageDir string
	stor       *storage.LocalStorage
	service    AdminBlobService
	user       model.User
}

func newAdminBlobTestEnv(t *testing.T) *adminBlobTestEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admin-blobs.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.StoragePolicy{},
		&model.UserGroup{},
		&model.User{},
		&model.UserFile{},
		&model.FileVersion{},
		&model.PhysicalFile{},
		&model.FileSystemSetting{},
		&model.QueueJob{},
		&model.BlobScanTask{},
	); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	policy := model.StoragePolicy{Name: "local policy", Type: "local"}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("seed policy: %v", err)
	}
	group := model.UserGroup{Name: "users", StoragePolicyID: policy.ID}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	user := model.User{Username: "blob-user", Email: "blob-user@example.com", Password: "x", Role: "admin", Enabled: true, UserGroupID: group.ID}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&model.FileSystemSetting{MaxBatchActionSize: 2, MimeMap: `{".txt":"text/plain"}`}).Error; err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	storageDir := t.TempDir()
	stor := storage.NewLocalStorage(storageDir)
	return &adminBlobTestEnv{db: db, storageDir: storageDir, stor: stor, service: NewAdminBlobService(db, stor), user: user}
}

func (e *adminBlobTestEnv) createPhysical(t *testing.T, hash, path string, data string, refCount int) model.PhysicalFile {
	t.Helper()
	if data != "" {
		if err := e.stor.Save(strings.NewReader(data), path); err != nil {
			t.Fatalf("save storage blob: %v", err)
		}
	}
	physical := model.PhysicalFile{
		FileHash:    hash,
		FileSize:    int64(len(data)),
		StoragePath: path,
		RefCount:    refCount,
		StorageType: "local",
		ContentType: "text/plain",
	}
	if err := e.db.Create(&physical).Error; err != nil {
		t.Fatalf("create physical file: %v", err)
	}
	return physical
}

func TestAdminBlobReferencedBlobCannotBeDeleted(t *testing.T) {
	env := newAdminBlobTestEnv(t)
	physical := env.createPhysical(t, "hash-ref", "referenced.txt", "referenced", 1)
	file := model.UserFile{UserID: env.user.ID, FileName: "referenced.txt", PhysicalFileID: &physical.ID, FileSize: physical.FileSize}
	if err := env.db.Create(&file).Error; err != nil {
		t.Fatalf("create user file: %v", err)
	}

	err := env.service.Delete(physical.ID)
	if err == nil || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("Delete referenced blob error = %v, want referenced blocker", err)
	}
	if !env.stor.Exists(physical.StoragePath) {
		t.Fatalf("storage file was deleted despite live user_file reference")
	}
}

func TestAdminBlobOrphanBlobDeletesDatabaseAndStorageFile(t *testing.T) {
	env := newAdminBlobTestEnv(t)
	physical := env.createPhysical(t, "hash-orphan", "orphan.txt", "orphan", 0)

	if err := env.service.Delete(physical.ID); err != nil {
		t.Fatalf("delete orphan blob: %v", err)
	}
	if env.stor.Exists(physical.StoragePath) {
		t.Fatalf("storage file still exists after orphan delete")
	}
	var count int64
	if err := env.db.Unscoped().Model(&model.PhysicalFile{}).Where("id = ?", physical.ID).Count(&count).Error; err != nil {
		t.Fatalf("count physical file: %v", err)
	}
	if count != 0 {
		t.Fatalf("physical file row count = %d, want 0", count)
	}
}

func TestAdminBlobBatchDeleteHonorsMaxBatchActionSize(t *testing.T) {
	env := newAdminBlobTestEnv(t)
	_, err := env.service.BatchDelete([]uint{1, 2, 3})
	if err == nil || !strings.Contains(err.Error(), "max_batch_action_size") {
		t.Fatalf("BatchDelete error = %v, want max_batch_action_size limit", err)
	}
}

func TestAdminBlobScanDetectsRefCountMismatch(t *testing.T) {
	env := newAdminBlobTestEnv(t)
	physical := env.createPhysical(t, "hash-mismatch", "mismatch.txt", "mismatch", 7)
	file := model.UserFile{UserID: env.user.ID, FileName: "mismatch.txt", PhysicalFileID: &physical.ID, FileSize: physical.FileSize}
	if err := env.db.Create(&file).Error; err != nil {
		t.Fatalf("create user file: %v", err)
	}

	result, err := env.service.Scan()
	if err != nil {
		t.Fatalf("scan blobs: %v", err)
	}
	if result.RefCountMismatch == 0 {
		t.Fatalf("scan RefCountMismatch = 0, want mismatch detected")
	}
	if !scanHasIssue(result, "ref_count_mismatch", physical.ID) {
		t.Fatalf("scan issues %#v do not include ref_count_mismatch for blob %d", result.Issues, physical.ID)
	}
}

func TestAdminBlobScanDetectsMissingOnStorage(t *testing.T) {
	env := newAdminBlobTestEnv(t)
	physical := env.createPhysical(t, "hash-missing", "missing.txt", "", 0)
	physical.FileSize = 10
	if err := env.db.Save(&physical).Error; err != nil {
		t.Fatalf("update missing physical size: %v", err)
	}
	_ = os.Remove(filepath.Join(env.storageDir, physical.StoragePath))

	result, err := env.service.Scan()
	if err != nil {
		t.Fatalf("scan blobs: %v", err)
	}
	if result.MissingOnStorage == 0 {
		t.Fatalf("scan MissingOnStorage = 0, want missing detected")
	}
	if !scanHasIssue(result, "missing_on_storage", physical.ID) {
		t.Fatalf("scan issues %#v do not include missing_on_storage for blob %d", result.Issues, physical.ID)
	}
}

func scanHasIssue(result *AdminBlobScanTaskPayload, issueType string, blobID uint) bool {
	for _, issue := range result.Issues {
		if issue.Type == issueType && issue.BlobID == blobID {
			return true
		}
	}
	return false
}
