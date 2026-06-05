package controller

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/crypto"
	"xingyunpan-v2/pkg/response"
)

func TestUploadAvatarUsesSiteSettingsPathLimitAndDimension(t *testing.T) {
	env := newAvatarSettingsEnv(t, model.SiteSetting{
		AvatarPath:        "custom/avatars",
		AvatarSizeLimitMB: 1,
		AvatarDimension:   64,
		GravatarServer:    "https://gravatar.example.test/",
	})

	resp := uploadAvatarForTest(t, env.router, createPNGAvatar(t, 120, 80), "wide.png")
	requireResponseCode(t, resp, response.CodeSuccess)

	avatarURL := responseDataStringForAvatar(t, resp, "avatar_url")
	if !strings.HasPrefix(avatarURL, "/api/v1/avatars/custom/avatars/") {
		t.Fatalf("avatar_url = %q, want configured URL path", avatarURL)
	}
	localPath, err := service.AvatarFilePathFromURLPath(strings.TrimPrefix(avatarURL, "/api/v1/avatars/"))
	if err != nil {
		t.Fatalf("map avatar URL to file: %v", err)
	}
	if !strings.Contains(filepath.ToSlash(localPath), "uploads/custom/avatars/") {
		t.Fatalf("stored path = %q, want uploads/custom/avatars", localPath)
	}

	file, err := os.Open(localPath)
	if err != nil {
		t.Fatalf("open stored avatar: %v", err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("decode stored avatar: %v", err)
	}
	if got := img.Bounds().Dx(); got != 64 {
		t.Fatalf("stored avatar width = %d, want 64", got)
	}
	if got := img.Bounds().Dy(); got != 64 {
		t.Fatalf("stored avatar height = %d, want 64", got)
	}

	req := httptest.NewRequest(http.MethodGet, avatarURL, nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("static avatar HTTP status = %d, want 200; body=%q", rec.Code, rec.Body.String())
	}
}

func TestUploadAvatarRejectsConfiguredOversizeAndIllegalType(t *testing.T) {
	env := newAvatarSettingsEnv(t, model.SiteSetting{
		AvatarPath:        "avatar-tests",
		AvatarSizeLimitMB: 1,
		AvatarDimension:   80,
		GravatarServer:    "https://gravatar.example.test/",
	})

	oversized := uploadAvatarForTest(t, env.router, bytes.Repeat([]byte("x"), (1<<20)+10), "large.png")
	if oversized.Code != response.CodeInvalidParams || !strings.Contains(oversized.Message, "1MB") {
		t.Fatalf("oversized response = code:%d message:%q, want 1MB validation", oversized.Code, oversized.Message)
	}

	invalid := uploadAvatarForTest(t, env.router, []byte("not an image"), "avatar.txt")
	if invalid.Code != response.CodeInvalidParams || !strings.Contains(invalid.Message, "JPEG, PNG, WebP and GIF") {
		t.Fatalf("invalid type response = code:%d message:%q, want type validation", invalid.Code, invalid.Message)
	}
}

func TestUserInfoUsesConfiguredGravatarWhenAvatarIsEmpty(t *testing.T) {
	env := newAvatarSettingsEnv(t, model.SiteSetting{
		AvatarPath:        "avatar-tests",
		AvatarSizeLimitMB: 1,
		AvatarDimension:   96,
		GravatarServer:    "https://gravatar.example.test/base",
	})

	user, err := env.userService.GetUserInfo(env.userID)
	if err != nil {
		t.Fatalf("get user info: %v", err)
	}
	if user.AvatarURL == "" {
		t.Fatalf("avatar_url is empty, want gravatar URL")
	}
	if !strings.HasPrefix(user.AvatarURL, "https://gravatar.example.test/base/avatar/") {
		t.Fatalf("avatar_url = %q, want configured gravatar server", user.AvatarURL)
	}
	if !strings.Contains(user.AvatarURL, "s=96") {
		t.Fatalf("avatar_url = %q, want configured avatar dimension in gravatar size", user.AvatarURL)
	}
}

type avatarSettingsEnv struct {
	router      *gin.Engine
	db          *gorm.DB
	userID      uint
	userService service.UserService
}

func newAvatarSettingsEnv(t *testing.T, setting model.SiteSetting) *avatarSettingsEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(previousWD) })

	db, err := gorm.Open(sqlite.Open(filepath.Join(tempDir, "avatar-settings.db")), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.SiteSetting{}, &model.UserGroup{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	if err := db.Create(&setting).Error; err != nil {
		t.Fatalf("seed site setting: %v", err)
	}
	password, err := crypto.HashPassword("password1")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := model.User{
		Username: "avatar-user",
		Email:    "avatar-user@example.com",
		Password: password,
		Role:     "user",
		Enabled:  true,
	}
	if err := db.Select("*").Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewUserGroupRepository(db)
	siteRepo := repository.NewSiteSettingRepository(db)
	userService := service.NewUserService(userRepo, groupRepo, siteRepo, "avatar-secret", 24, 168, 0, nil, nil, nil, 300, 60)
	avatarService := service.NewAvatarService(siteRepo)
	userController := NewUserControllerWithAvatarService(userService, avatarService)

	router := gin.New()
	router.GET("/api/v1/avatars/*filepath", func(c *gin.Context) {
		localPath, err := service.AvatarFilePathFromURLPath(c.Param("filepath"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		if _, err := os.Stat(localPath); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(localPath)
	})
	router.POST("/api/v1/user/avatar", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		userController.UploadAvatar(c)
	})

	return &avatarSettingsEnv{router: router, db: db, userID: user.ID, userService: userService}
}

func createPNGAvatar(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 180, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func uploadAvatarForTest(t *testing.T, router *gin.Engine, content []byte, fileName string) response.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("avatar", fileName)
	if err != nil {
		t.Fatalf("create multipart avatar: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write multipart avatar: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/avatar", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode avatar response HTTP %d body=%q: %v", rec.Code, rec.Body.String(), err)
	}
	return resp
}

func responseDataStringForAvatar(t *testing.T, resp response.Response, key string) string {
	t.Helper()
	value, ok := resp.Data.(map[string]interface{})[key].(string)
	if !ok || value == "" {
		t.Fatalf("response data missing string %q: %#v", key, resp.Data)
	}
	return value
}
