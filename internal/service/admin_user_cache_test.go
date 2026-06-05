package service

import (
	"path/filepath"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	sqlite "github.com/glebarez/sqlite"
	goredis "github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/cache"
	"xingyunpan-v2/pkg/crypto"
)

func TestAdminUserUpdateInvalidatesUserInfoCache(t *testing.T) {
	env := newAdminUserCacheEnv(t)
	user, groupA, groupB := seedAdminUserCacheUser(t, env)

	cached, err := env.userService.GetUserInfo(user.ID)
	if err != nil {
		t.Fatalf("prime user info cache: %v", err)
	}
	if cached.Role != "user" || !cached.Enabled || cached.Capacity != 1024 || cached.UserGroupID != groupA.ID {
		t.Fatalf("initial cached user = %+v, want role=user enabled=true capacity=1024 group=%d", cached, groupA.ID)
	}

	_, err = env.adminUserService.Update(user.ID, &AdminUserUpsertPayload{
		Username:            user.Username,
		Email:               user.Email,
		Role:                "admin",
		Enabled:             false,
		UserGroupID:         groupB.ID,
		Capacity:            4096,
		FollowGroupCapacity: false,
	}, 0)
	if err != nil {
		t.Fatalf("admin update user: %v", err)
	}

	fresh, err := env.userService.GetUserInfo(user.ID)
	if err != nil {
		t.Fatalf("get user info after admin update: %v", err)
	}
	if fresh.Role != "admin" {
		t.Fatalf("role = %q, want admin", fresh.Role)
	}
	if fresh.Enabled {
		t.Fatalf("enabled = true, want false")
	}
	if fresh.Capacity != 4096 {
		t.Fatalf("capacity = %d, want 4096", fresh.Capacity)
	}
	if fresh.UserGroupID != groupB.ID {
		t.Fatalf("user_group_id = %d, want %d", fresh.UserGroupID, groupB.ID)
	}
}

func TestAdminUserBatchUpdatesInvalidateUserInfoCache(t *testing.T) {
	env := newAdminUserCacheEnv(t)
	user, groupA, groupB := seedAdminUserCacheUser(t, env)

	if _, err := env.userService.GetUserInfo(user.ID); err != nil {
		t.Fatalf("prime user info cache: %v", err)
	}

	if _, err := env.adminUserService.BatchUpdateRole(&AdminUserBatchRolePayload{
		IDs:  []uint{user.ID},
		Role: "admin",
	}, 0); err != nil {
		t.Fatalf("batch update role: %v", err)
	}
	afterRole, err := env.userService.GetUserInfo(user.ID)
	if err != nil {
		t.Fatalf("get user info after role update: %v", err)
	}
	if afterRole.Role != "admin" || afterRole.UserGroupID != groupA.ID {
		t.Fatalf("after role update = %+v, want role=admin group=%d", afterRole, groupA.ID)
	}

	if _, err := env.adminUserService.BatchUpdateStatus(&AdminUserBatchStatusPayload{
		IDs:     []uint{user.ID},
		Enabled: false,
	}, 0); err != nil {
		t.Fatalf("batch update status: %v", err)
	}
	afterStatus, err := env.userService.GetUserInfo(user.ID)
	if err != nil {
		t.Fatalf("get user info after status update: %v", err)
	}
	if afterStatus.Enabled {
		t.Fatalf("enabled = true, want false")
	}

	if _, err := env.adminUserService.BatchUpdateGroup(&AdminUserBatchGroupPayload{
		IDs:         []uint{user.ID},
		UserGroupID: groupB.ID,
	}, 0); err != nil {
		t.Fatalf("batch update group: %v", err)
	}
	afterGroup, err := env.userService.GetUserInfo(user.ID)
	if err != nil {
		t.Fatalf("get user info after group update: %v", err)
	}
	if afterGroup.UserGroupID != groupB.ID {
		t.Fatalf("user_group_id = %d, want %d", afterGroup.UserGroupID, groupB.ID)
	}
	if afterGroup.Capacity != groupB.MaxCapacity {
		t.Fatalf("capacity = %d, want group max capacity %d", afterGroup.Capacity, groupB.MaxCapacity)
	}
}

type adminUserCacheEnv struct {
	db               *gorm.DB
	userService      UserService
	adminUserService AdminUserService
}

func newAdminUserCacheEnv(t *testing.T) *adminUserCacheEnv {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admin-user-cache.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserGroup{}); err != nil {
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

	userRepo := repository.NewUserRepository(db)
	userGroupRepo := repository.NewUserGroupRepository(db)
	cacheService := cache.NewCacheService(redisClient)

	return &adminUserCacheEnv{
		db: db,
		userService: NewUserService(
			userRepo,
			userGroupRepo,
			nil,
			"admin-user-cache-secret",
			24,
			168,
			0,
			cacheService,
			redisClient,
			nil,
			300,
			60,
		),
		adminUserService: NewAdminUserService(userRepo, userGroupRepo, cacheService),
	}
}

func seedAdminUserCacheUser(t *testing.T, env *adminUserCacheEnv) (model.User, model.UserGroup, model.UserGroup) {
	t.Helper()

	groupA := model.UserGroup{Name: "Group A", MaxCapacity: 1024}
	groupB := model.UserGroup{Name: "Group B", MaxCapacity: 8192}
	if err := env.db.Create(&groupA).Error; err != nil {
		t.Fatalf("seed group A: %v", err)
	}
	if err := env.db.Create(&groupB).Error; err != nil {
		t.Fatalf("seed group B: %v", err)
	}

	password, err := crypto.HashPassword("secret1")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := model.User{
		Username:    "cached-user",
		Email:       "cached@example.com",
		Password:    password,
		Role:        "user",
		Enabled:     true,
		UserGroupID: groupA.ID,
		Capacity:    1024,
	}
	if err := env.db.Select("*").Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	return user, groupA, groupB
}
