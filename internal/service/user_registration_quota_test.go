package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	sqlite "github.com/glebarez/sqlite"
	goredis "github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

const registrationDefaultCapacity = int64(10 * 1024 * 1024 * 1024)

type registrationQuotaEnv struct {
	db          *gorm.DB
	redisClient *goredis.Client
	service     UserService
}

func newRegistrationQuotaEnv(t *testing.T, globalDefaultCapacity int64) *registrationQuotaEnv {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "registration-quota.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserGroup{}, &model.SiteSetting{}, &model.UserPreference{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(redisServer.Close)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	return &registrationQuotaEnv{
		db:          db,
		redisClient: redisClient,
		service: NewUserService(
			repository.NewUserRepository(db),
			repository.NewUserGroupRepository(db),
			repository.NewSiteSettingRepository(db),
			"test-secret",
			24,
			168,
			globalDefaultCapacity,
			nil,
			redisClient,
			nil,
			300,
			60,
		),
	}
}

func TestRegisterInheritsUserGroupCapacityFromDefaultUserGroup(t *testing.T) {
	env := newRegistrationQuotaEnv(t, registrationDefaultCapacity)
	userGroup := seedRegistrationGroup(t, env.db, "User", quotaTwoGB)
	seedDefaultGroupSetting(t, env.db, "User")

	user := registerWithCode(t, env, "new-user", "new-user@example.com")
	if user.UserGroupID != userGroup.ID {
		t.Fatalf("user_group_id = %d, want %d", user.UserGroupID, userGroup.ID)
	}
	if user.Capacity != quotaTwoGB {
		t.Fatalf("capacity = %d, want %d", user.Capacity, quotaTwoGB)
	}
}

func TestRegisterInheritsUnlimitedCapacityFromDefaultAdminGroup(t *testing.T) {
	env := newRegistrationQuotaEnv(t, registrationDefaultCapacity)
	adminGroup := seedRegistrationGroup(t, env.db, "Admin", 0)
	seedRegistrationGroup(t, env.db, "User", quotaTwoGB)
	seedDefaultGroupSetting(t, env.db, "Admin")

	user := registerWithCode(t, env, "admin-default-user", "admin-default-user@example.com")
	if user.UserGroupID != adminGroup.ID {
		t.Fatalf("user_group_id = %d, want %d", user.UserGroupID, adminGroup.ID)
	}
	if user.Capacity != 0 {
		t.Fatalf("capacity = %d, want unlimited 0", user.Capacity)
	}
}

func TestRegisterFallsBackToGlobalCapacityWhenDefaultGroupMissing(t *testing.T) {
	env := newRegistrationQuotaEnv(t, registrationDefaultCapacity)
	firstGroup := seedRegistrationGroup(t, env.db, "User", quotaTwoGB)
	seedDefaultGroupSetting(t, env.db, "Missing")

	user := registerWithCode(t, env, "missing-default-user", "missing-default-user@example.com")
	if user.UserGroupID != firstGroup.ID {
		t.Fatalf("existing default group fallback changed: user_group_id = %d, want first group %d", user.UserGroupID, firstGroup.ID)
	}
	if user.Capacity != registrationDefaultCapacity {
		t.Fatalf("capacity = %d, want global default %d", user.Capacity, registrationDefaultCapacity)
	}
}

func seedRegistrationGroup(t *testing.T, db *gorm.DB, name string, maxCapacity int64) model.UserGroup {
	t.Helper()
	group := model.UserGroup{Name: name, MaxCapacity: maxCapacity}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed user group %s: %v", name, err)
	}
	return group
}

func seedDefaultGroupSetting(t *testing.T, db *gorm.DB, defaultGroup string) {
	t.Helper()
	if err := db.Create(&model.SiteSetting{DefaultGroup: defaultGroup}).Error; err != nil {
		t.Fatalf("seed site setting: %v", err)
	}
}

func registerWithCode(t *testing.T, env *registrationQuotaEnv, username, email string) *model.User {
	t.Helper()
	code := "123456"
	if err := env.redisClient.Set(context.Background(), registerCodeKey(email), code, time.Minute).Err(); err != nil {
		t.Fatalf("seed register code: %v", err)
	}
	user, err := env.service.Register(username, "secret1", email, code)
	if err != nil {
		t.Fatalf("register user: %v", err)
	}
	return user
}
