package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
	"testing"

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

const quotaTwoGB = int64(2 * 1024 * 1024 * 1024)

type quotaTestEnv struct {
	db               *gorm.DB
	stor             *storage.LocalStorage
	redisServer      *miniredis.Miniredis
	redisClient      *goredis.Client
	user             model.User
	fileService      FileService
	multipartService MultipartService
}

func newQuotaTestEnv(t *testing.T, capacity int64) *quotaTestEnv {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "quota.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.UserGroup{},
		&model.User{},
		&model.UserFile{},
		&model.PhysicalFile{},
		&model.MultipartUpload{},
		&model.FileSystemSetting{},
		&model.StoragePolicy{},
		&model.StoragePolicyHitLog{},
		&model.TrafficEvent{},
	); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	group := model.UserGroup{Name: "Quota Users", MaxCapacity: capacity}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	user := model.User{
		Username:    "quota-user",
		Email:       "quota-user@example.com",
		Password:    "x",
		Role:        "user",
		Enabled:     true,
		UserGroupID: group.ID,
		Capacity:    capacity,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&model.FileSystemSetting{
		UploadSessionTTL:      3600,
		BlobSignedURLTTL:      3600,
		BlobSignedURLReuseTTL: 600,
		StaticCacheTTL:        0,
	}).Error; err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(redisServer.Close)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	stor := storage.NewLocalStorage(t.TempDir())
	physicalRepo := repository.NewPhysicalFileRepository(db)
	userFileRepo := repository.NewUserFileRepository(db)
	userRepo := repository.NewUserRepository(db)
	settingsRepo := repository.NewFileSystemSettingRepository(db)
	multipartRepo := repository.NewMultipartUploadRepository(db)

	return &quotaTestEnv{
		db:          db,
		stor:        stor,
		redisServer: redisServer,
		redisClient: redisClient,
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
	}
}

func TestQuotaUnlimitedAllowsDirectUpload(t *testing.T) {
	env := newQuotaTestEnv(t, 0)

	uploaded, err := env.fileService.Upload(env.user.ID, "unlimited.txt", 5, strings.NewReader("hello"), nil)
	if err != nil {
		t.Fatalf("upload under unlimited capacity: %v", err)
	}
	if uploaded.FileSize != 5 {
		t.Fatalf("uploaded file size = %d, want 5", uploaded.FileSize)
	}
	assertUserUsage(t, env.db, env.user.ID, 5)
}

func TestQuotaDirectUploadFailsNearTwoGBWithoutSideEffects(t *testing.T) {
	env := newQuotaTestEnv(t, quotaTwoGB)
	nearLimit := quotaTwoGB - 5
	setUserUsage(t, env.db, env.user.ID, nearLimit)

	_, err := env.fileService.Upload(env.user.ID, "too-large.txt", 6, strings.NewReader("123456"), nil)
	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("upload should fail with quota exceeded, err=%v", err)
	}
	assertUserUsage(t, env.db, env.user.ID, nearLimit)
	assertUserFileCount(t, env.db, env.user.ID, 0)
}

func TestQuotaMultipartCompleteFailsNearTwoGBWithoutSideEffects(t *testing.T) {
	env := newQuotaTestEnv(t, quotaTwoGB)
	ctx := context.Background()
	data := []byte("123456")
	uploadID := initRecordedMultipart(t, env, ctx, "multipart-too-large.txt", data)
	nearLimit := quotaTwoGB - 5
	setUserUsage(t, env.db, env.user.ID, nearLimit)

	_, err := env.multipartService.CompleteMultipartUpload(ctx, uploadID, env.user.ID, nil)
	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("multipart complete should fail with quota exceeded, err=%v", err)
	}
	assertUserUsage(t, env.db, env.user.ID, nearLimit)
	assertUserFileCount(t, env.db, env.user.ID, 0)
}

func TestQuotaCopyFailsNearTwoGBWithoutSideEffects(t *testing.T) {
	env := newQuotaTestEnv(t, 0)
	source, err := env.fileService.Upload(env.user.ID, "source.txt", 6, strings.NewReader("123456"), nil)
	if err != nil {
		t.Fatalf("seed source upload: %v", err)
	}
	if err := env.db.Model(&model.User{}).
		Where("id = ?", env.user.ID).
		Updates(map[string]interface{}{"capacity": quotaTwoGB, "used_size": quotaTwoGB - 5}).Error; err != nil {
		t.Fatalf("set near limit: %v", err)
	}

	_, err = env.fileService.Copy(env.user.ID, source.ID, nil)
	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("copy should fail with quota exceeded, err=%v", err)
	}
	assertUserUsage(t, env.db, env.user.ID, quotaTwoGB-5)
	assertUserFileCount(t, env.db, env.user.ID, 1)
}

func TestAdminUserCapacityInheritsGroupQuota(t *testing.T) {
	env := newQuotaTestEnv(t, quotaTwoGB)
	group := model.UserGroup{Name: "Inherited Group", MaxCapacity: quotaTwoGB}
	if err := env.db.Create(&group).Error; err != nil {
		t.Fatalf("seed inherited group: %v", err)
	}

	adminService := NewAdminUserService(repository.NewUserRepository(env.db), repository.NewUserGroupRepository(env.db))
	created, err := adminService.Create(&AdminUserUpsertPayload{
		Username:            "inherited",
		Email:               "inherited@example.com",
		Password:            "secret1",
		Role:                "user",
		Enabled:             true,
		UserGroupID:         group.ID,
		Capacity:            0,
		FollowGroupCapacity: true,
	})
	if err != nil {
		t.Fatalf("admin create user: %v", err)
	}
	if created.Capacity != quotaTwoGB {
		t.Fatalf("created capacity = %d, want %d", created.Capacity, quotaTwoGB)
	}

	moved, err := adminService.BatchUpdateGroup(&AdminUserBatchGroupPayload{
		IDs:         []uint{env.user.ID},
		UserGroupID: group.ID,
	}, 0)
	if err != nil {
		t.Fatalf("batch move group: %v", err)
	}
	if len(moved) != 1 || moved[0].Capacity != quotaTwoGB {
		t.Fatalf("moved user did not inherit group capacity: %#v", moved)
	}
}

func TestAdminUserGroupChangeUpdatesStoragePolicyRuntime(t *testing.T) {
	env := newQuotaTestEnv(t, quotaTwoGB)
	firstPolicy := model.StoragePolicy{Name: "First Policy", Type: "local"}
	secondPolicy := model.StoragePolicy{Name: "Second Policy", Type: "local"}
	if err := env.db.Create(&firstPolicy).Error; err != nil {
		t.Fatalf("seed first policy: %v", err)
	}
	if err := env.db.Create(&secondPolicy).Error; err != nil {
		t.Fatalf("seed second policy: %v", err)
	}

	var currentGroup model.UserGroup
	if err := env.db.First(&currentGroup, env.user.UserGroupID).Error; err != nil {
		t.Fatalf("load current group: %v", err)
	}
	currentGroup.StoragePolicyID = firstPolicy.ID
	if err := env.db.Save(&currentGroup).Error; err != nil {
		t.Fatalf("update current group policy: %v", err)
	}

	targetGroup := model.UserGroup{Name: "Policy Target", StoragePolicyID: secondPolicy.ID, MaxCapacity: quotaTwoGB}
	if err := env.db.Create(&targetGroup).Error; err != nil {
		t.Fatalf("seed target group: %v", err)
	}

	beforePolicy, err := newStoragePolicyRuntime(env.db, env.user.ID).loadPolicy()
	if err != nil {
		t.Fatalf("load before policy: %v", err)
	}
	if beforePolicy == nil || beforePolicy.ID != firstPolicy.ID {
		t.Fatalf("before policy = %#v, want %d", beforePolicy, firstPolicy.ID)
	}

	adminService := NewAdminUserService(repository.NewUserRepository(env.db), repository.NewUserGroupRepository(env.db))
	_, err = adminService.Update(env.user.ID, &AdminUserUpsertPayload{
		Username:            env.user.Username,
		Email:               env.user.Email,
		Role:                env.user.Role,
		Enabled:             true,
		UserGroupID:         targetGroup.ID,
		FollowGroupCapacity: true,
	}, 0)
	if err != nil {
		t.Fatalf("update user group: %v", err)
	}

	afterPolicy, err := newStoragePolicyRuntime(env.db, env.user.ID).loadPolicy()
	if err != nil {
		t.Fatalf("load after policy: %v", err)
	}
	if afterPolicy == nil || afterPolicy.ID != secondPolicy.ID {
		t.Fatalf("after policy = %#v, want %d", afterPolicy, secondPolicy.ID)
	}
}

func TestAdminUserUpdateFollowGroupCapacityUsesTargetGroupMaxCapacity(t *testing.T) {
	env := newQuotaTestEnv(t, quotaTwoGB)
	targetCapacity := int64(5 * 1024 * 1024 * 1024)
	group := model.UserGroup{Name: "Five GB", MaxCapacity: targetCapacity}
	if err := env.db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}

	adminService := NewAdminUserService(repository.NewUserRepository(env.db), repository.NewUserGroupRepository(env.db))
	updated, err := adminService.Update(env.user.ID, &AdminUserUpsertPayload{
		Username:            env.user.Username,
		Email:               env.user.Email,
		Role:                env.user.Role,
		Enabled:             true,
		UserGroupID:         group.ID,
		Capacity:            123,
		FollowGroupCapacity: true,
	}, 0)
	if err != nil {
		t.Fatalf("update follow group capacity: %v", err)
	}
	if updated.Capacity != targetCapacity {
		t.Fatalf("capacity = %d, want group max %d", updated.Capacity, targetCapacity)
	}
}

func TestAdminUserUpdateCustomCapacityIsNotOverwrittenByGroup(t *testing.T) {
	env := newQuotaTestEnv(t, quotaTwoGB)
	group := model.UserGroup{Name: "Large Group", MaxCapacity: int64(8 * 1024 * 1024 * 1024)}
	if err := env.db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	customCapacity := int64(3 * 1024 * 1024 * 1024)

	adminService := NewAdminUserService(repository.NewUserRepository(env.db), repository.NewUserGroupRepository(env.db))
	updated, err := adminService.Update(env.user.ID, &AdminUserUpsertPayload{
		Username:            env.user.Username,
		Email:               env.user.Email,
		Role:                env.user.Role,
		Enabled:             true,
		UserGroupID:         group.ID,
		Capacity:            customCapacity,
		FollowGroupCapacity: false,
	}, 0)
	if err != nil {
		t.Fatalf("update custom capacity: %v", err)
	}
	if updated.Capacity != customCapacity {
		t.Fatalf("capacity = %d, want custom %d", updated.Capacity, customCapacity)
	}
}

func TestAdminUserLowerCapacityBelowUsedBlocksNewUploadWithoutDeletingExistingFiles(t *testing.T) {
	env := newQuotaTestEnv(t, 0)
	uploaded, err := env.fileService.Upload(env.user.ID, "existing.txt", 6, strings.NewReader("123456"), nil)
	if err != nil {
		t.Fatalf("seed existing upload: %v", err)
	}

	adminService := NewAdminUserService(repository.NewUserRepository(env.db), repository.NewUserGroupRepository(env.db))
	updated, err := adminService.Update(env.user.ID, &AdminUserUpsertPayload{
		Username:            env.user.Username,
		Email:               env.user.Email,
		Role:                env.user.Role,
		Enabled:             true,
		UserGroupID:         env.user.UserGroupID,
		Capacity:            5,
		FollowGroupCapacity: false,
	}, 0)
	if err != nil {
		t.Fatalf("lower capacity: %v", err)
	}
	if updated.Capacity != 5 {
		t.Fatalf("capacity = %d, want 5", updated.Capacity)
	}

	_, err = env.fileService.Upload(env.user.ID, "new.txt", 1, strings.NewReader("x"), nil)
	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("new upload should fail with quota exceeded, err=%v", err)
	}

	var existing model.UserFile
	if err := env.db.First(&existing, uploaded.ID).Error; err != nil {
		t.Fatalf("existing file disappeared: %v", err)
	}
	if existing.FileName != "existing.txt" {
		t.Fatalf("existing file name = %q", existing.FileName)
	}
	assertUserUsage(t, env.db, env.user.ID, 6)
	assertUserFileCount(t, env.db, env.user.ID, 1)
}

func initRecordedMultipart(t *testing.T, env *quotaTestEnv, ctx context.Context, fileName string, data []byte) string {
	t.Helper()
	hash := sha256HexBytes(data)
	result, err := env.multipartService.InitMultipartUpload(ctx, env.user.ID, fileName, hash, int64(len(data)), 3, nil)
	if err != nil {
		t.Fatalf("init multipart: %v", err)
	}
	for chunk := 1; chunk <= result.TotalChunks; chunk++ {
		start := (chunk - 1) * result.ChunkSize
		end := start + result.ChunkSize
		if end > len(data) {
			end = len(data)
		}
		path := "multipart/" + result.UploadID + "/chunk_" + string(rune('0'+chunk))
		if err := env.stor.Save(bytes.NewReader(data[start:end]), path); err != nil {
			t.Fatalf("save chunk %d: %v", chunk, err)
		}
		if err := env.multipartService.RecordChunkUpload(ctx, result.UploadID, env.user.ID, chunk, "etag"); err != nil {
			t.Fatalf("record chunk %d: %v", chunk, err)
		}
	}
	return result.UploadID
}

func setUserUsage(t *testing.T, db *gorm.DB, userID uint, usedSize int64) {
	t.Helper()
	if err := db.Model(&model.User{}).Where("id = ?", userID).Update("used_size", usedSize).Error; err != nil {
		t.Fatalf("set user usage: %v", err)
	}
}

func assertUserUsage(t *testing.T, db *gorm.DB, userID uint, want int64) {
	t.Helper()
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.UsedSize != want {
		t.Fatalf("used_size = %d, want %d", user.UsedSize, want)
	}
}

func assertUserFileCount(t *testing.T, db *gorm.DB, userID uint, want int64) {
	t.Helper()
	var count int64
	if err := db.Model(&model.UserFile{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		t.Fatalf("count user files: %v", err)
	}
	if count != want {
		t.Fatalf("user file count = %d, want %d", count, want)
	}
}

func sha256HexBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
