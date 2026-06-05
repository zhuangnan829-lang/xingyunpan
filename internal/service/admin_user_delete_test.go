package service

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

type adminUserDeleteEnv struct {
	db      *gorm.DB
	service AdminUserService
	group   model.UserGroup
}

func newAdminUserDeleteEnv(t *testing.T) *adminUserDeleteEnv {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admin-user-delete.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserGroup{},
		&model.UserFile{},
		&model.Share{},
		&model.RecycleBin{},
		&model.OfflineDownloadTask{},
		&model.MultipartUpload{},
		&model.DavAccount{},
		&model.OAuthApp{},
		&model.OAuthCredential{},
		&model.Collaboration{},
		&model.TrafficEvent{},
		&model.UserPreference{},
		&model.FileCustomPropertyValue{},
		&model.StoragePolicyHitLog{},
	); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	group := model.UserGroup{Name: "Delete Test", MaxCapacity: 1024}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	return &adminUserDeleteEnv{
		db:      db,
		group:   group,
		service: NewAdminUserService(repository.NewUserRepository(db), repository.NewUserGroupRepository(db)),
	}
}

func TestAdminUserDeleteEmptyAccountSucceeds(t *testing.T) {
	env := newAdminUserDeleteEnv(t)
	user := seedAdminDeleteUser(t, env, "empty")

	if err := env.service.Delete(user.ID, 0); err != nil {
		t.Fatalf("delete empty user: %v", err)
	}
	if _, err := repository.NewUserRepository(env.db).GetByID(user.ID); err == nil {
		t.Fatalf("user still exists after delete")
	}
}

func TestAdminUserDeleteWithFilesIsRejected(t *testing.T) {
	env := newAdminUserDeleteEnv(t)
	user := seedAdminDeleteUser(t, env, "with-files")
	if err := env.db.Create(&model.UserFile{
		UserID:   user.ID,
		FileName: "asset.txt",
		FileSize: 10,
	}).Error; err != nil {
		t.Fatalf("seed file: %v", err)
	}

	err := env.service.Delete(user.ID, 0)
	var blocked *AdminUserDeleteBlockedError
	if !errors.As(err, &blocked) {
		t.Fatalf("delete error = %v, want blocked error", err)
	}
	if blocked.Preview == nil || blocked.Preview.FileCount != 1 {
		t.Fatalf("blocked preview = %#v, want file_count=1", blocked.Preview)
	}
}

func TestAdminUserBatchDeleteRejectsCurrentAdministrator(t *testing.T) {
	env := newAdminUserDeleteEnv(t)
	admin := seedAdminDeleteUser(t, env, "admin")
	other := seedAdminDeleteUser(t, env, "other")

	err := env.service.BatchDelete(&AdminUserBatchDeletePayload{IDs: []uint{other.ID, admin.ID}}, admin.ID)
	if err == nil || !strings.Contains(err.Error(), "currently logged-in administrator") {
		t.Fatalf("batch delete error = %v, want current administrator rejection", err)
	}
}

func TestAdminUserBatchDeleteRejectsAssetUser(t *testing.T) {
	env := newAdminUserDeleteEnv(t)
	empty := seedAdminDeleteUser(t, env, "empty-batch")
	asset := seedAdminDeleteUser(t, env, "asset-batch")
	if err := env.db.Create(&model.Share{
		UserID:     asset.ID,
		ShareToken: "asset-token",
	}).Error; err != nil {
		t.Fatalf("seed share: %v", err)
	}

	err := env.service.BatchDelete(&AdminUserBatchDeletePayload{IDs: []uint{empty.ID, asset.ID}}, 0)
	var blocked *AdminUserDeleteBlockedError
	if !errors.As(err, &blocked) {
		t.Fatalf("batch delete error = %v, want blocked error", err)
	}
	if blocked.UserID != asset.ID || blocked.Preview == nil || blocked.Preview.ShareCount != 1 {
		t.Fatalf("blocked = %#v, preview=%#v", blocked, blocked.Preview)
	}
}

func seedAdminDeleteUser(t *testing.T, env *adminUserDeleteEnv, username string) model.User {
	t.Helper()
	user := model.User{
		Username:    username,
		Email:       username + "@example.com",
		Password:    "password",
		Role:        "user",
		Enabled:     true,
		UserGroupID: env.group.ID,
		Capacity:    1024,
	}
	if err := env.db.Select("*").Create(&user).Error; err != nil {
		t.Fatalf("seed user %s: %v", username, err)
	}
	return user
}
