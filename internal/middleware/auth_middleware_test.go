package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	goredis "github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/cache"
	"xingyunpan-v2/pkg/crypto"
	"xingyunpan-v2/pkg/jwt"
)

func TestAuthMiddlewareRejectsDisabledOldTokenAndAllowsAfterReenable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newAuthMiddlewareDB(t)
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(redisServer.Close)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	ctx := context.Background()

	previousDB := config.DB
	previousRedis := config.RDB
	previousConfig := config.Config
	if config.Config == nil {
		config.Config = &config.AppConfig{}
	}
	config.DB = db
	config.RDB = redisClient
	config.Config.JWT.Secret = "middleware-test-secret"
	t.Cleanup(func() {
		config.DB = previousDB
		config.RDB = previousRedis
		config.Config = previousConfig
	})

	group := model.UserGroup{Name: "User", MaxCapacity: 1024}
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("seed group: %v", err)
	}
	password, err := crypto.HashPassword("secret1")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := model.User{
		Username:    "normal-user",
		Email:       "normal@example.com",
		Password:    password,
		Role:        "user",
		Enabled:     true,
		UserGroupID: group.ID,
		Capacity:    1024,
	}
	if err := db.Select("*").Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, config.Config.JWT.Secret, 1)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	router := protectedTestRouter()

	if status := requestProtected(router, token); status != http.StatusOK {
		t.Fatalf("initial request status = %d, want %d", status, http.StatusOK)
	}

	adminService := service.NewAdminUserService(
		repository.NewUserRepository(db),
		repository.NewUserGroupRepository(db),
		cache.NewCacheService(redisClient),
	)
	if _, err := adminService.UpdateStatus(user.ID, &service.AdminUserStatusPayload{Enabled: false}, 0); err != nil {
		t.Fatalf("disable user: %v", err)
	}
	var disabledUser model.User
	if err := db.First(&disabledUser, user.ID).Error; err != nil {
		t.Fatalf("reload disabled user: %v", err)
	}
	if disabledUser.Enabled {
		t.Fatalf("user remained enabled after admin disable")
	}
	if exists, err := redisClient.Exists(ctx, "user:auth-status:"+strconv.Itoa(int(user.ID))).Result(); err != nil {
		t.Fatalf("check auth cache key: %v", err)
	} else if exists != 0 {
		t.Fatalf("auth status cache key still exists after admin disable")
	}
	if status := requestProtected(router, token); status != http.StatusForbidden {
		t.Fatalf("disabled old-token request status = %d, want %d", status, http.StatusForbidden)
	}

	if _, err := adminService.UpdateStatus(user.ID, &service.AdminUserStatusPayload{Enabled: true}, 0); err != nil {
		t.Fatalf("reenable user: %v", err)
	}
	if status := requestProtected(router, token); status != http.StatusOK {
		t.Fatalf("reenabled request status = %d, want %d", status, http.StatusOK)
	}
}

func newAuthMiddlewareDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "auth-middleware.db")), &gorm.Config{
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
	return db
}

func protectedTestRouter() *gin.Engine {
	router := gin.New()
	router.GET("/files", AuthMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func requestProtected(router http.Handler, token string) int {
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}
