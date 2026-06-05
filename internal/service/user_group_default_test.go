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

type userGroupDefaultEnv struct {
	db               *gorm.DB
	redisClient      *goredis.Client
	userGroupService UserGroupService
	userService      UserService
	policy           model.StoragePolicy
}

func newUserGroupDefaultEnv(t *testing.T) *userGroupDefaultEnv {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "user-group-default.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.UserGroup{},
		&model.SiteSetting{},
		&model.UserPreference{},
		&model.StoragePolicy{},
	); err != nil {
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

	policy := model.StoragePolicy{Name: "Default policy", Type: "local"}
	if err := db.Create(&policy).Error; err != nil {
		t.Fatalf("seed storage policy: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	userGroupRepo := repository.NewUserGroupRepository(db)
	siteSettingRepo := repository.NewSiteSettingRepository(db)
	storagePolicyRepo := repository.NewStoragePolicyRepository(db)

	return &userGroupDefaultEnv{
		db:          db,
		redisClient: redisClient,
		policy:      policy,
		userGroupService: NewUserGroupService(
			userGroupRepo,
			userRepo,
			storagePolicyRepo,
			siteSettingRepo,
		),
		userService: NewUserService(
			userRepo,
			userGroupRepo,
			siteSettingRepo,
			"test-secret",
			24,
			168,
			registrationDefaultCapacity,
			nil,
			redisClient,
			nil,
			300,
			60,
		),
	}
}

func TestUserGroupRenameSyncsDefaultGroupAndRegistration(t *testing.T) {
	env := newUserGroupDefaultEnv(t)
	userGroup := seedDefaultTestGroup(t, env.db, "User", env.policy.ID)
	seedDefaultGroupSetting(t, env.db, " User ")

	updated, err := env.userGroupService.Update(userGroup.ID, &UserGroupPayload{
		Name:            "Member",
		Description:     "Renamed default group",
		StoragePolicyID: env.policy.ID,
		MaxCapacity:     quotaTwoGB,
	})
	if err != nil {
		t.Fatalf("rename default user group: %v", err)
	}
	if updated.Name != "Member" {
		t.Fatalf("updated group name = %q, want Member", updated.Name)
	}

	var setting model.SiteSetting
	if err := env.db.First(&setting).Error; err != nil {
		t.Fatalf("load site setting: %v", err)
	}
	if setting.DefaultGroup != "Member" {
		t.Fatalf("default_group = %q, want Member", setting.DefaultGroup)
	}

	user := registerDefaultTestUser(t, env, "member-user", "member-user@example.com")
	if user.UserGroupID != userGroup.ID {
		t.Fatalf("registered user_group_id = %d, want renamed group id %d", user.UserGroupID, userGroup.ID)
	}
	if user.Capacity != quotaTwoGB {
		t.Fatalf("registered capacity = %d, want %d", user.Capacity, quotaTwoGB)
	}
}

func TestCannotDeleteCurrentDefaultUserGroupAfterTrim(t *testing.T) {
	env := newUserGroupDefaultEnv(t)
	userGroup := seedDefaultTestGroup(t, env.db, "User", env.policy.ID)
	seedDefaultTestGroup(t, env.db, "Other", env.policy.ID)
	seedDefaultGroupSetting(t, env.db, " User ")

	err := env.userGroupService.Delete(userGroup.ID)
	if err == nil {
		t.Fatal("expected deleting current default user group to fail")
	}
}

func seedDefaultTestGroup(t *testing.T, db *gorm.DB, name string, policyID uint) model.UserGroup {
	t.Helper()
	group := model.UserGroup{
		Name:            name,
		StoragePolicyID: policyID,
		MaxCapacity:     quotaTwoGB,
	}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed user group %s: %v", name, err)
	}
	return group
}

func registerDefaultTestUser(t *testing.T, env *userGroupDefaultEnv, username, email string) *model.User {
	t.Helper()
	code := "123456"
	if err := env.redisClient.Set(context.Background(), registerCodeKey(email), code, time.Minute).Err(); err != nil {
		t.Fatalf("seed register code: %v", err)
	}
	user, err := env.userService.Register(username, "secret1", email, code)
	if err != nil {
		t.Fatalf("register user: %v", err)
	}
	return user
}
