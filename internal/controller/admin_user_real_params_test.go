package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	sqlite "github.com/glebarez/sqlite"
	goredis "github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/middleware"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/cache"
	"xingyunpan-v2/pkg/crypto"
	applogger "xingyunpan-v2/pkg/logger"
	xredis "xingyunpan-v2/pkg/redis"
	"xingyunpan-v2/pkg/response"
	"xingyunpan-v2/pkg/storage"
)

func TestAdminUserManagementInterfacesAffectDatabaseAndRuntime(t *testing.T) {
	env := newAdminUserRealParamsEnv(t)

	adminToken := loginToken(t, env.router, "root", "rootpass1", response.CodeSuccess)

	createResp := apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users", adminToken, map[string]interface{}{
		"username":              "managed",
		"email":                 "managed@example.com",
		"password":              "initial1",
		"role":                  "user",
		"enabled":               true,
		"user_group_id":         env.groupSmall.ID,
		"capacity":              int64(4096),
		"follow_group_capacity": false,
	})
	requireResponseCode(t, createResp, response.CodeSuccess)

	var user model.User
	if err := env.db.Where("username = ?", "managed").First(&user).Error; err != nil {
		t.Fatalf("created user not found in users table: %v", err)
	}
	if user.Password == "initial1" {
		t.Fatalf("users.password stored raw password")
	}
	if err := crypto.VerifyPassword(user.Password, "initial1"); err != nil {
		t.Fatalf("stored password hash does not match initial password: %v", err)
	}
	oldToken := loginToken(t, env.router, "managed", "initial1", response.CodeSuccess)

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", user.ID), adminToken, map[string]interface{}{
		"username":              "managed-new",
		"email":                 "managed-new@example.com",
		"role":                  "user",
		"enabled":               true,
		"user_group_id":         env.groupSmall.ID,
		"capacity":              int64(4096),
		"follow_group_capacity": false,
	}), response.CodeSuccess)
	loginToken(t, env.router, "managed-new", "initial1", response.CodeSuccess)
	loginToken(t, env.router, "managed-new@example.com", "initial1", response.CodeSuccess)
	loginToken(t, env.router, "managed", "initial1", response.CodeUnauthorized)
	loginToken(t, env.router, "managed@example.com", "initial1", response.CodeUnauthorized)

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/reset-password", user.ID), adminToken, map[string]interface{}{
		"password": "resetpass1",
	}), response.CodeSuccess)
	loginToken(t, env.router, "managed-new", "initial1", response.CodeUnauthorized)
	userToken := loginToken(t, env.router, "managed-new", "resetpass1", response.CodeSuccess)

	adminListAsUser := apiJSON(t, env.router, http.MethodGet, "/api/v1/admin/users", userToken, nil)
	if adminListAsUser.Code != response.CodeForbidden {
		t.Fatalf("user admin access code = %d, want %d", adminListAsUser.Code, response.CodeForbidden)
	}
	requireResponseCode(t, apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users/batch-role", adminToken, map[string]interface{}{
		"ids":  []uint{user.ID},
		"role": "admin",
	}), response.CodeSuccess)
	requireResponseCode(t, apiJSON(t, env.router, http.MethodGet, "/api/v1/admin/users", userToken, nil), response.CodeSuccess)

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d/status", user.ID), adminToken, map[string]interface{}{
		"enabled": false,
	}), response.CodeSuccess)
	loginToken(t, env.router, "managed-new", "resetpass1", response.CodeUnauthorized)
	oldTokenResp := apiJSON(t, env.router, http.MethodGet, "/api/v1/user/info", oldToken, nil)
	if oldTokenResp.Code != response.CodeForbidden {
		t.Fatalf("disabled old-token /user/info code = %d, want %d", oldTokenResp.Code, response.CodeForbidden)
	}
	requireResponseCode(t, apiJSON(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d/status", user.ID), adminToken, map[string]interface{}{
		"enabled": true,
	}), response.CodeSuccess)
	userToken = loginToken(t, env.router, "managed-new", "resetpass1", response.CodeSuccess)

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", user.ID), adminToken, map[string]interface{}{
		"username":              "managed-new",
		"email":                 "managed-new@example.com",
		"role":                  "admin",
		"enabled":               true,
		"user_group_id":         env.groupEncrypted.ID,
		"capacity":              int64(8192),
		"follow_group_capacity": false,
	}), response.CodeSuccess)
	uploadResp := uploadFile(t, env.router, userToken, "new-encrypted.bin", bytes.Repeat([]byte("a"), 2048))
	requireResponseCode(t, uploadResp, response.CodeSuccess)
	encryptedFileID := responseDataUint(t, uploadResp, "id")
	var encryptedFile model.UserFile
	if err := env.db.Preload("PhysicalFile").First(&encryptedFile, encryptedFileID).Error; err != nil {
		t.Fatalf("load encrypted upload: %v", err)
	}
	if encryptedFile.PhysicalFile == nil {
		t.Fatalf("uploaded file missed physical file")
	}
	if !strings.HasPrefix(encryptedFile.PhysicalFile.StoragePath, fmt.Sprintf("encrypted/%d/", user.ID)) {
		t.Fatalf("storage_path = %q, want encrypted/%d prefix", encryptedFile.PhysicalFile.StoragePath, user.ID)
	}
	if !encryptedFile.PhysicalFile.Encrypted || encryptedFile.PhysicalFile.EncryptionKeyID != "managed-key" {
		t.Fatalf("encryption metadata = encrypted:%v key:%q, want true managed-key", encryptedFile.PhysicalFile.Encrypted, encryptedFile.PhysicalFile.EncryptionKeyID)
	}
	oversizedByPolicy := uploadFile(t, env.router, userToken, "new-too-large.bin", bytes.Repeat([]byte("b"), 9000))
	if oversizedByPolicy.Code != response.CodeInvalidParams {
		t.Fatalf("oversized storage-policy upload code = %d, want %d", oversizedByPolicy.Code, response.CodeInvalidParams)
	}

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", user.ID), adminToken, map[string]interface{}{
		"username":              "managed-new",
		"email":                 "managed-new@example.com",
		"role":                  "admin",
		"enabled":               true,
		"user_group_id":         env.groupCDN.ID,
		"capacity":              int64(8192),
		"follow_group_capacity": false,
	}), response.CodeSuccess)
	cdnResp := uploadFile(t, env.router, userToken, "cdn-download.bin", bytes.Repeat([]byte("c"), 1024))
	requireResponseCode(t, cdnResp, response.CodeSuccess)
	cdnFileID := responseDataUint(t, cdnResp, "id")
	downloadReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/file/%d/download", cdnFileID), nil)
	downloadReq.Header.Set("Authorization", "Bearer "+userToken)
	downloadRec := httptest.NewRecorder()
	env.router.ServeHTTP(downloadRec, downloadReq)
	if downloadRec.Code != http.StatusFound {
		t.Fatalf("cdn download HTTP status = %d, want %d; body=%s", downloadRec.Code, http.StatusFound, downloadRec.Body.String())
	}
	if location := downloadRec.Header().Get("Location"); !strings.HasPrefix(location, "https://cdn.example.test/cdn/") {
		t.Fatalf("cdn Location = %q, want cdn prefix", location)
	}

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", user.ID), adminToken, map[string]interface{}{
		"username":              "managed-new",
		"email":                 "managed-new@example.com",
		"role":                  "admin",
		"enabled":               true,
		"user_group_id":         env.groupCDN.ID,
		"capacity":              int64(2500),
		"follow_group_capacity": false,
	}), response.CodeSuccess)
	quotaFail := uploadFile(t, env.router, userToken, "cdn-quota-fail.bin", bytes.Repeat([]byte("d"), 1024))
	if quotaFail.Code != response.CodeInternalError || !strings.Contains(quotaFail.Message, "upload failed") {
		t.Fatalf("quota-lowered upload response = code:%d message:%q, want upload failure", quotaFail.Code, quotaFail.Message)
	}
	requireResponseCode(t, apiJSON(t, env.router, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", user.ID), adminToken, map[string]interface{}{
		"username":              "managed-new",
		"email":                 "managed-new@example.com",
		"role":                  "admin",
		"enabled":               true,
		"user_group_id":         env.groupCDN.ID,
		"capacity":              int64(20000),
		"follow_group_capacity": false,
	}), response.CodeSuccess)
	requireResponseCode(t, uploadFile(t, env.router, userToken, "cdn-quota-ok.bin", bytes.Repeat([]byte("e"), 1024)), response.CodeSuccess)

	deletePreview := apiJSON(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d/delete-preview", user.ID), adminToken, nil)
	requireResponseCode(t, deletePreview, response.CodeSuccess)
	if !responseDataBool(t, deletePreview, "has_blocking_assets") {
		t.Fatalf("delete preview has_blocking_assets = false, want true")
	}
	deleteResp := apiJSON(t, env.router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", user.ID), adminToken, nil)
	if deleteResp.Code != response.CodeInvalidParams {
		t.Fatalf("asset user delete code = %d, want %d", deleteResp.Code, response.CodeInvalidParams)
	}
}

func TestAdminUserBatchInterfacesMatchSingleOperationSemantics(t *testing.T) {
	env := newAdminUserRealParamsEnv(t)
	adminToken := loginToken(t, env.router, "root", "rootpass1", response.CodeSuccess)

	firstID := createManagedUserViaAPI(t, env, adminToken, "batch-one", env.groupSmall.ID, 4096)
	secondID := createManagedUserViaAPI(t, env, adminToken, "batch-two", env.groupSmall.ID, 4096)

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users/batch-group", adminToken, map[string]interface{}{
		"ids":           []uint{firstID, secondID},
		"user_group_id": env.groupEncrypted.ID,
	}), response.CodeSuccess)
	assertUserFields(t, env.db, firstID, map[string]interface{}{"user_group_id": env.groupEncrypted.ID, "capacity": env.groupEncrypted.MaxCapacity})
	assertUserFields(t, env.db, secondID, map[string]interface{}{"user_group_id": env.groupEncrypted.ID, "capacity": env.groupEncrypted.MaxCapacity})

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users/batch-role", adminToken, map[string]interface{}{
		"ids":  []uint{firstID, secondID},
		"role": "admin",
	}), response.CodeSuccess)
	assertUserFields(t, env.db, firstID, map[string]interface{}{"role": "admin"})
	assertUserFields(t, env.db, secondID, map[string]interface{}{"role": "admin"})

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users/batch-status", adminToken, map[string]interface{}{
		"ids":     []uint{firstID, secondID},
		"enabled": false,
	}), response.CodeSuccess)
	loginToken(t, env.router, "batch-one", "initial1", response.CodeUnauthorized)
	loginToken(t, env.router, "batch-two", "initial1", response.CodeUnauthorized)

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users/batch-status", adminToken, map[string]interface{}{
		"ids":     []uint{firstID, secondID},
		"enabled": true,
	}), response.CodeSuccess)

	assetUserID := createManagedUserViaAPI(t, env, adminToken, "batch-asset", env.groupCDN.ID, 4096)
	assetToken := loginToken(t, env.router, "batch-asset", "initial1", response.CodeSuccess)
	requireResponseCode(t, uploadFile(t, env.router, assetToken, "cdn-asset.bin", bytes.Repeat([]byte("x"), 256)), response.CodeSuccess)
	batchDeleteBlocked := apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users/batch-delete", adminToken, map[string]interface{}{
		"ids": []uint{firstID, assetUserID},
	})
	if batchDeleteBlocked.Code != response.CodeInvalidParams || !strings.Contains(batchDeleteBlocked.Message, "assets") {
		t.Fatalf("batch delete blocked response = code:%d message:%q, want asset error", batchDeleteBlocked.Code, batchDeleteBlocked.Message)
	}
	var firstStillThere int64
	if err := env.db.Model(&model.User{}).Where("id = ?", firstID).Count(&firstStillThere).Error; err != nil {
		t.Fatalf("count first user after blocked batch delete: %v", err)
	}
	if firstStillThere != 1 {
		t.Fatalf("batch delete partially deleted empty user when asset user blocked")
	}

	requireResponseCode(t, apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users/batch-delete", adminToken, map[string]interface{}{
		"ids": []uint{firstID, secondID},
	}), response.CodeSuccess)
	assertUserDeleted(t, env.db, firstID)
	assertUserDeleted(t, env.db, secondID)
}

type adminUserRealParamsEnv struct {
	db             *gorm.DB
	router         *gin.Engine
	groupSmall     model.UserGroup
	groupEncrypted model.UserGroup
	groupCDN       model.UserGroup
}

func newAdminUserRealParamsEnv(t *testing.T) *adminUserRealParamsEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "admin-real-params.db")), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.StoragePolicy{},
		&model.StoragePolicyHitLog{},
		&model.UserGroup{},
		&model.User{},
		&model.UserFile{},
		&model.PhysicalFile{},
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

	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(redisServer.Close)
	redisClient := goredis.NewClient(&goredis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	previousDB := config.DB
	previousRedis := config.RDB
	previousConfig := config.Config
	config.DB = db
	config.RDB = redisClient
	config.Config = &config.AppConfig{}
	config.Config.JWT.Secret = "admin-real-params-secret"
	previousLogger := applogger.Logger
	if err := applogger.Init(&applogger.Config{Level: "error", Format: "console", Output: "stderr"}); err != nil {
		t.Fatalf("init test logger: %v", err)
	}
	t.Cleanup(func() {
		config.DB = previousDB
		config.RDB = previousRedis
		config.Config = previousConfig
		applogger.Logger = previousLogger
	})

	policySmall := model.StoragePolicy{
		Name:            "Small text policy",
		Type:            "local",
		BlobPath:        "small/{uid}",
		BlobNamePattern: "{originname}",
		MaxFileSize:     1,
		MaxFileSizeUnit: "KB",
		ExtensionMode:   "allow",
		Extensions:      "txt",
		NameRuleMode:    "allow",
		NameRegex:       `^ok-[a-z0-9-]+\.txt$`,
	}
	policyEncrypted := model.StoragePolicy{
		Name:               "Encrypted bin policy",
		Type:               "local",
		BlobPath:           "encrypted/{uid}",
		BlobNamePattern:    "enc-{originname}",
		MaxFileSize:        8,
		MaxFileSizeUnit:    "KB",
		ExtensionMode:      "allow",
		Extensions:         "bin",
		NameRuleMode:       "allow",
		NameRegex:          `^new-[a-z0-9-]+\.bin$`,
		EnableEncryption:   true,
		EncryptionKeyID:    "managed-key",
		ParallelChunkCount: 2,
	}
	policyCDN := model.StoragePolicy{
		Name:            "CDN bin policy",
		Type:            "local",
		BlobPath:        "cdn/{uid}",
		BlobNamePattern: "{originname}",
		MaxFileSize:     16,
		MaxFileSizeUnit: "KB",
		ExtensionMode:   "allow",
		Extensions:      "bin",
		NameRuleMode:    "allow",
		NameRegex:       `^cdn-[a-z0-9-]+\.bin$`,
		EnableCDN:       true,
		DownloadCDN:     "https://cdn.example.test",
	}
	if err := db.Create(&policySmall).Error; err != nil {
		t.Fatalf("seed small policy: %v", err)
	}
	if err := db.Create(&policyEncrypted).Error; err != nil {
		t.Fatalf("seed encrypted policy: %v", err)
	}
	if err := db.Create(&policyCDN).Error; err != nil {
		t.Fatalf("seed cdn policy: %v", err)
	}

	groupSmall := model.UserGroup{Name: "Small", StoragePolicyID: policySmall.ID, MaxCapacity: 4096}
	groupEncrypted := model.UserGroup{Name: "Encrypted", StoragePolicyID: policyEncrypted.ID, MaxCapacity: 8192}
	groupCDN := model.UserGroup{Name: "CDN", StoragePolicyID: policyCDN.ID, MaxCapacity: 16384}
	if err := db.Create(&groupSmall).Error; err != nil {
		t.Fatalf("seed small group: %v", err)
	}
	if err := db.Create(&groupEncrypted).Error; err != nil {
		t.Fatalf("seed encrypted group: %v", err)
	}
	if err := db.Create(&groupCDN).Error; err != nil {
		t.Fatalf("seed cdn group: %v", err)
	}

	adminPassword, err := crypto.HashPassword("rootpass1")
	if err != nil {
		t.Fatalf("hash admin password: %v", err)
	}
	admin := model.User{
		Username:    "root",
		Email:       "root@example.com",
		Password:    adminPassword,
		Role:        "admin",
		Enabled:     true,
		UserGroupID: groupSmall.ID,
		Capacity:    0,
	}
	if err := db.Select("*").Create(&admin).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	if err := db.Create(&model.FileSystemSetting{StaticCacheTTL: 0, MimeMap: `{"bin":"application/octet-stream","txt":"text/plain"}`}).Error; err != nil {
		t.Fatalf("seed file settings: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewUserGroupRepository(db)
	cacheService := cache.NewCacheService(redisClient)
	userService := service.NewUserService(userRepo, groupRepo, nil, config.Config.JWT.Secret, 24, 168, 0, cacheService, redisClient, nil, 300, 60)
	adminUserService := service.NewAdminUserService(userRepo, groupRepo, cacheService)
	fileService := service.NewFileService(
		db,
		repository.NewPhysicalFileRepository(db),
		repository.NewUserFileRepository(db),
		userRepo,
		repository.NewFileSystemSettingRepository(db),
		nil,
		storage.NewLocalStorage(t.TempDir()),
		redisClient,
		cacheService,
		nil,
	)

	userController := NewUserController(userService)
	adminController := NewAdminUserController(adminUserService)
	fileController := NewFileController(fileService, service.NewFileSystemSettingService(repository.NewFileSystemSettingRepository(db), xredis.NewMultipartRedis(redisClient)), nil)

	router := gin.New()
	api := router.Group("/api/v1")
	api.POST("/login", userController.Login)
	auth := api.Group("")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/user/info", userController.GetUserInfo)
	auth.GET("/file", fileController.List)
	auth.POST("/file/upload", fileController.Upload)
	auth.GET("/file/:id/download", fileController.Download)
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	adminGroup.GET("/users", adminController.List)
	adminGroup.POST("/users", adminController.Create)
	adminGroup.PUT("/users/:id", adminController.Update)
	adminGroup.PUT("/users/:id/status", adminController.UpdateStatus)
	adminGroup.POST("/users/:id/reset-password", adminController.ResetPassword)
	adminGroup.GET("/users/:id/delete-preview", adminController.DeletePreview)
	adminGroup.DELETE("/users/:id", adminController.Delete)
	adminGroup.POST("/users/batch-group", adminController.BatchUpdateGroup)
	adminGroup.POST("/users/batch-role", adminController.BatchUpdateRole)
	adminGroup.POST("/users/batch-status", adminController.BatchUpdateStatus)
	adminGroup.POST("/users/batch-delete", adminController.BatchDelete)

	return &adminUserRealParamsEnv{db: db, router: router, groupSmall: groupSmall, groupEncrypted: groupEncrypted, groupCDN: groupCDN}
}

func createManagedUserViaAPI(t *testing.T, env *adminUserRealParamsEnv, adminToken, username string, groupID uint, capacity int64) uint {
	t.Helper()
	resp := apiJSON(t, env.router, http.MethodPost, "/api/v1/admin/users", adminToken, map[string]interface{}{
		"username":              username,
		"email":                 username + "@example.com",
		"password":              "initial1",
		"role":                  "user",
		"enabled":               true,
		"user_group_id":         groupID,
		"capacity":              capacity,
		"follow_group_capacity": false,
	})
	requireResponseCode(t, resp, response.CodeSuccess)
	var user model.User
	if err := env.db.Where("username = ?", username).First(&user).Error; err != nil {
		t.Fatalf("load created %s: %v", username, err)
	}
	return user.ID
}

func loginToken(t *testing.T, router *gin.Engine, username, password string, wantCode int) string {
	t.Helper()
	resp := apiJSON(t, router, http.MethodPost, "/api/v1/login", "", map[string]interface{}{
		"username": username,
		"password": password,
	})
	if resp.Code != wantCode {
		t.Fatalf("login(%s) code = %d message=%q, want %d", username, resp.Code, resp.Message, wantCode)
	}
	if wantCode != response.CodeSuccess {
		return ""
	}
	token, _ := resp.Data.(map[string]interface{})["access_token"].(string)
	if token == "" {
		t.Fatalf("login(%s) missed access_token in data %#v", username, resp.Data)
	}
	return token
}

func apiJSON(t *testing.T, router *gin.Engine, method, path, token string, body interface{}) response.Response {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(data)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return decodeAPIResponse(t, rec)
}

func uploadFile(t *testing.T, router *gin.Engine, token, fileName string, content []byte) response.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/file/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return decodeAPIResponse(t, rec)
}

func decodeAPIResponse(t *testing.T, rec *httptest.ResponseRecorder) response.Response {
	t.Helper()
	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response HTTP %d body=%q: %v", rec.Code, rec.Body.String(), err)
	}
	return resp
}

func requireResponseCode(t *testing.T, resp response.Response, want int) {
	t.Helper()
	if resp.Code != want {
		t.Fatalf("response code = %d message=%q data=%#v, want %d", resp.Code, resp.Message, resp.Data, want)
	}
}

func responseDataUint(t *testing.T, resp response.Response, key string) uint {
	t.Helper()
	value, ok := resp.Data.(map[string]interface{})[key].(float64)
	if !ok {
		t.Fatalf("response data missing numeric %q: %#v", key, resp.Data)
	}
	return uint(value)
}

func responseDataBool(t *testing.T, resp response.Response, key string) bool {
	t.Helper()
	value, ok := resp.Data.(map[string]interface{})[key].(bool)
	if !ok {
		t.Fatalf("response data missing bool %q: %#v", key, resp.Data)
	}
	return value
}

func assertUserFields(t *testing.T, db *gorm.DB, userID uint, fields map[string]interface{}) {
	t.Helper()
	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		t.Fatalf("load user %d: %v", userID, err)
	}
	for field, want := range fields {
		switch field {
		case "role":
			if user.Role != want {
				t.Fatalf("user %d role = %q, want %q", userID, user.Role, want)
			}
		case "user_group_id":
			if user.UserGroupID != want {
				t.Fatalf("user %d user_group_id = %d, want %d", userID, user.UserGroupID, want)
			}
		case "capacity":
			if user.Capacity != want {
				t.Fatalf("user %d capacity = %d, want %d", userID, user.Capacity, want)
			}
		}
	}
}

func assertUserDeleted(t *testing.T, db *gorm.DB, userID uint) {
	t.Helper()
	var count int64
	if err := db.Model(&model.User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		t.Fatalf("count user %d: %v", userID, err)
	}
	if count != 0 {
		t.Fatalf("user %d still exists after delete", userID)
	}
}
