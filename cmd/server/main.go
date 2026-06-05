package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/controller"
	"xingyunpan-v2/internal/middleware"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/cache"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/redis"
	"xingyunpan-v2/pkg/storage"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "xingyunpan-v2/docs/swagger"
)

func shouldIgnoreSettingMigrationError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "can't drop 'uni_email_templates_template_key'") ||
		isLegacyQueueSettingsDropMissingMessage(message) ||
		strings.Contains(message, "can't drop foreign key") ||
		(strings.Contains(message, "email_templates") && strings.Contains(message, "already exists")) ||
		(strings.Contains(message, "table 'email_templates' already exists"))
}

func isLegacyQueueSettingsDropMissingMessage(message string) bool {
	return strings.Contains(message, "uni_queue_settings_queue_key") &&
		strings.Contains(message, "drop") &&
		(strings.Contains(message, "check that column/key exists") ||
			strings.Contains(message, "can't drop") ||
			strings.Contains(message, "error 1091") ||
			strings.Contains(message, "1091"))
}

func shouldIgnoreLegacyConstraintMigrationError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "can't drop") && strings.Contains(message, "check that column/key exists")
}

func ensureEssentialFileTables() error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS user_files (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL,
			deleted_at DATETIME(3) NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			parent_id BIGINT UNSIGNED NULL,
			file_name VARCHAR(255) NOT NULL,
			physical_file_id BIGINT UNSIGNED NULL,
			is_folder TINYINT(1) NOT NULL DEFAULT 0,
			file_size BIGINT NOT NULL DEFAULT 0,
			file_path VARCHAR(1000) NULL,
			PRIMARY KEY (id),
			INDEX idx_user_files_deleted_at (deleted_at),
			INDEX idx_user_parent (user_id, parent_id),
			INDEX idx_user_folder (is_folder),
			INDEX idx_user_files_physical_file_id (physical_file_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS shares (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL,
			deleted_at DATETIME(3) NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			share_token VARCHAR(64) NOT NULL,
			access_code_hash VARCHAR(255) NULL,
			expires_at DATETIME(3) NULL,
			max_downloads BIGINT NULL,
			download_count BIGINT NOT NULL DEFAULT 0,
			access_count BIGINT NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			UNIQUE KEY idx_share_token (share_token),
			INDEX idx_shares_deleted_at (deleted_at),
			INDEX idx_user_id (user_id),
			INDEX idx_expires_at (expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS share_files (
			share_id BIGINT UNSIGNED NOT NULL,
			file_id BIGINT UNSIGNED NOT NULL,
			PRIMARY KEY (share_id, file_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS file_versions (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL,
			deleted_at DATETIME(3) NULL,
			file_id BIGINT UNSIGNED NOT NULL,
			version_number INT NOT NULL,
			physical_file_id BIGINT UNSIGNED NOT NULL,
			file_size BIGINT NOT NULL,
			uploader_id BIGINT UNSIGNED NOT NULL,
			is_current TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (id),
			INDEX idx_file_versions_deleted_at (deleted_at),
			INDEX idx_file_versions (file_id, version_number),
			INDEX idx_current_version (is_current)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS multipart_uploads (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL,
			deleted_at DATETIME(3) NULL,
			upload_id VARCHAR(64) NOT NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			file_name VARCHAR(255) NOT NULL,
			file_hash VARCHAR(64) NOT NULL,
			file_size BIGINT NOT NULL,
			total_chunks INT NOT NULL,
			chunk_size INT NOT NULL,
			storage_type VARCHAR(20) NOT NULL,
			storage_path VARCHAR(500) NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'uploading',
			completed_at DATETIME(3) NULL,
			PRIMARY KEY (id),
			UNIQUE KEY idx_multipart_uploads_upload_id (upload_id),
			INDEX idx_multipart_uploads_deleted_at (deleted_at),
			INDEX idx_user_status (user_id, status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS collaborations (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL,
			deleted_at DATETIME(3) NULL,
			file_id BIGINT UNSIGNED NOT NULL,
			owner_id BIGINT UNSIGNED NOT NULL,
			collaborator_id BIGINT UNSIGNED NOT NULL,
			permission VARCHAR(20) NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY unique_collaboration (file_id, collaborator_id),
			INDEX idx_collaborations_deleted_at (deleted_at),
			INDEX idx_file_owner (file_id, owner_id),
			INDEX idx_collaborator (collaborator_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS recycle_bin (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			file_id BIGINT UNSIGNED NOT NULL,
			file_name VARCHAR(255) NOT NULL,
			file_size BIGINT NOT NULL,
			file_type VARCHAR(100) NULL,
			original_path VARCHAR(1000) NOT NULL,
			deleted_at DATETIME(3) NOT NULL,
			expires_at DATETIME(3) NOT NULL,
			PRIMARY KEY (id),
			INDEX idx_user_expires (user_id, expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS user_preferences (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL,
			deleted_at DATETIME(3) NULL,
			user_id BIGINT UNSIGNED NOT NULL,
			language VARCHAR(32) NOT NULL DEFAULT 'zh-CN',
			timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai',
			mode VARCHAR(16) NOT NULL DEFAULT 'light',
			theme VARCHAR(32) NOT NULL DEFAULT 'sky',
			keep_versions TINYINT(1) NOT NULL DEFAULT 1,
			version_extensions VARCHAR(512) NOT NULL DEFAULT '',
			max_versions BIGINT NOT NULL DEFAULT 10,
			view_sync VARCHAR(16) NOT NULL DEFAULT 'server',
			expand_tree TINYINT(1) NOT NULL DEFAULT 1,
			folder_action VARCHAR(16) NOT NULL DEFAULT 'open',
			home_visibility VARCHAR(32) NOT NULL DEFAULT 'passwordless',
			PRIMARY KEY (id),
			UNIQUE KEY idx_user_preferences_user_id (user_id),
			INDEX idx_user_preferences_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS oauth_apps (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			created_at DATETIME(3) NULL,
			updated_at DATETIME(3) NULL,
			deleted_at DATETIME(3) NULL,
			slug VARCHAR(80) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT NULL,
			app_name VARCHAR(255) NOT NULL,
			icon_path TEXT NULL,
			client_id VARCHAR(128) NOT NULL,
			client_secret VARCHAR(255) NULL,
			redirect_uris_json TEXT NULL,
			scopes_json TEXT NULL,
			permissions_json TEXT NULL,
			is_system TINYINT(1) NOT NULL DEFAULT 0,
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			token_ttl VARCHAR(64) NOT NULL DEFAULT '7 天',
			refresh_token_ttl_seconds BIGINT NOT NULL DEFAULT 604800,
			PRIMARY KEY (id),
			UNIQUE KEY idx_oauth_apps_slug (slug),
			UNIQUE KEY idx_oauth_apps_client_id (client_id),
			INDEX idx_oauth_apps_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	if !db.Migrator().HasColumn(&model.Share{}, "max_downloads") {
		if err := db.Exec(`ALTER TABLE shares ADD COLUMN max_downloads BIGINT NULL AFTER expires_at`).Error; err != nil {
			return fmt.Errorf("add shares.max_downloads failed: %w", err)
		}
	}
	if db.Migrator().HasTable(&model.User{}) && !db.Migrator().HasColumn(&model.User{}, "last_seen_at") {
		if err := db.Exec(`ALTER TABLE users ADD COLUMN last_seen_at DATETIME(3) NULL`).Error; err != nil {
			return fmt.Errorf("add users.last_seen_at failed: %w", err)
		}
		if err := db.Exec(`CREATE INDEX idx_users_last_seen_at ON users (last_seen_at)`).Error; err != nil {
			logger.Warn("create users.last_seen_at index failed", zap.Error(err))
		}
	}
	if err := ensureUserAvatarColumn(); err != nil {
		return err
	}

	return nil
}

func ensureUserAvatarColumn() error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}

	if db.Migrator().HasColumn(&model.User{}, "AvatarURL") {
		return nil
	}
	if err := db.Migrator().AddColumn(&model.User{}, "AvatarURL"); err != nil {
		return fmt.Errorf("add users.avatar_url column failed: %w", err)
	}
	return nil
}

func startOAuthRefreshScheduler(runtimeService service.FileSystemRuntimeService) {
	if runtimeService == nil {
		return
	}

	go func() {
		for {
			interval, err := runtimeService.OAuthRefreshInterval()
			if err != nil || interval <= 0 {
				logger.Warn("oauth refresh scheduler interval unavailable", zap.Error(err))
				interval = time.Hour
			}

			time.Sleep(interval)
			if _, err := runtimeService.RunOAuthRefresh(); err != nil {
				logger.Warn("oauth credential refresh pass failed", zap.Error(err))
			}
		}
	}()
}

func main() {
	if err := config.LoadDefault(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := initLogger(); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	logger.Info("日志初始化成功")

	if err := config.InitDatabase(); err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer config.CloseDatabase()

	if err := config.AutoMigrate(
		&model.SiteSetting{},
		&model.CaptchaSetting{},
		&model.MediaSetting{},
		&model.FileSystemSetting{},
		&model.FullTextSearchSetting{},
		&model.EmailSetting{},
		&model.EmailTemplate{},
		&model.EventSetting{},
		&model.QueueJob{},
		&model.BlobScanTask{},
		&model.ArchiveExtractTask{},
		&model.TrafficEvent{},
		&model.FileCustomPropertyValue{},
		&model.StoragePolicy{},
		&model.StoragePolicyAudit{},
		&model.StoragePolicyHitLog{},
		&model.Node{},
		&model.DavAccount{},
		&model.OAuthApp{},
		&model.OAuthCredential{},
		&model.OAuthAuthorizationCode{},
		&model.OAuthAccessToken{},
		&model.OAuthRefreshToken{},
		&model.OAuthGrant{},
		&model.OAuthAuditLog{},
		&model.UserGroup{},
		&model.User{},
		&model.UserPreference{},
	); err != nil {
		if shouldIgnoreSettingMigrationError(err) {
			logger.Warn("管理设置表迁移已跳过 - 使用现有表结构", zap.Error(err))
		} else {
			logger.Fatal("管理设置表迁移失败", zap.Error(err))
		}
	}

	if err := config.AutoMigrate(&model.EventSetting{}); err != nil {
		logger.Fatal("event settings table migration failed", zap.Error(err))
	}
	if err := repository.NewQueueSettingRepository(config.GetDB()).EnsureSchema(); err != nil {
		logger.Fatal("queue settings table migration failed", zap.Error(err))
	}

	fileModels := []struct {
		name  string
		model interface{}
	}{
		{name: "physical_files", model: &model.PhysicalFile{}},
		{name: "blob_scan_tasks", model: &model.BlobScanTask{}},
		{name: "user_files", model: &model.UserFile{}},
		{name: "shares", model: &model.Share{}},
		{name: "share_files", model: &model.ShareFile{}},
		{name: "file_versions", model: &model.FileVersion{}},
		{name: "multipart_uploads", model: &model.MultipartUpload{}},
		{name: "collaborations", model: &model.Collaboration{}},
		{name: "recycle_bin", model: &model.RecycleBin{}},
		{name: "file_deletions", model: &model.FileDeletion{}},
		{name: "traffic_events", model: &model.TrafficEvent{}},
		{name: "storage_policy_hit_logs", model: &model.StoragePolicyHitLog{}},
	}

	for _, fileModel := range fileModels {
		if err := config.AutoMigrate(fileModel.model); err != nil {
			if shouldIgnoreLegacyConstraintMigrationError(err) {
				logger.Warn("文件业务表迁移已跳过旧约束冲突", zap.String("table", fileModel.name), zap.Error(err))
				continue
			}
			logger.Fatal("文件业务表迁移失败", zap.String("table", fileModel.name), zap.Error(err))
		}
	}

	if err := ensureEssentialFileTables(); err != nil {
		logger.Fatal("文件业务表兜底建表失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	if err := config.InitRedis(); err != nil {
		logger.Warn("Redis 连接失败，部分功能可能受限", zap.Error(err))
	} else {
		defer config.CloseRedis()
		logger.Info("Redis 连接成功")
	}

	gin.SetMode(config.Config.Server.Mode)

	if config.Config.Server.Mode != "release" {
		go func() {
			logger.Info("pprof 性能分析已启用", zap.String("addr", "localhost:6060"))
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				logger.Error("pprof 服务启动失败", zap.Error(err))
			}
		}()
	}

	router := setupRouter()

	srv := &http.Server{
		Addr:         config.Config.Server.GetAddr(),
		Handler:      router,
		ReadTimeout:  config.Config.Server.GetReadTimeout(),
		WriteTimeout: config.Config.Server.GetWriteTimeout(),
	}

	go func() {
		logger.Info("服务启动", zap.String("addr", srv.Addr), zap.String("mode", config.Config.Server.Mode))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("服务启动失败", zap.Error(err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务强制关闭", zap.Error(err))
	}

	logger.Info("服务已关闭")
}

func initLogger() error {
	cfg := &logger.Config{
		Level:      config.Config.Log.Level,
		Format:     config.Config.Log.Format,
		Output:     config.Config.Log.Output,
		FilePath:   config.Config.Log.FilePath,
		MaxSize:    config.Config.Log.MaxSize,
		MaxBackups: config.Config.Log.MaxBackups,
		MaxAge:     config.Config.Log.MaxAge,
		Compress:   config.Config.Log.Compress,
	}

	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "console"
	}
	if cfg.Output == "" {
		cfg.Output = "stdout"
	}

	return logger.Init(cfg)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.MetricsMiddleware())

	rateLimitConfig := middleware.NewRateLimitConfig()
	router.Use(middleware.RateLimitMiddleware(rateLimitConfig))

	db := config.GetDB()

	physicalFileRepo := repository.NewPhysicalFileRepository(db)
	userFileRepo := repository.NewUserFileRepository(db)
	userRepo := repository.NewUserRepository(db)
	multipartRepo := repository.NewMultipartUploadRepository(db)
	shareRepo := repository.NewShareRepository(db)
	recycleRepo := repository.NewRecycleRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	collaborationRepo := repository.NewCollaborationRepository(db)
	siteSettingRepo := repository.NewSiteSettingRepository(db)
	captchaSettingRepo := repository.NewCaptchaSettingRepository(db)
	mediaSettingRepo := repository.NewMediaSettingRepository(db)
	fileSystemSettingRepo := repository.NewFileSystemSettingRepository(db)
	fileCustomPropertyValueRepo := repository.NewFileCustomPropertyValueRepository(db)
	fullTextSearchSettingRepo := repository.NewFullTextSearchSettingRepository(db)
	storagePolicyRepo := repository.NewStoragePolicyRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	davAccountRepo := repository.NewDavAccountRepository(db)
	offlineDownloadRepo := repository.NewOfflineDownloadTaskRepository(db)
	userGroupRepo := repository.NewUserGroupRepository(db)
	emailSettingRepo := repository.NewEmailSettingRepository(db)
	emailTemplateRepo := repository.NewEmailTemplateRepository(db)
	eventSettingRepo := repository.NewEventSettingRepository(db)
	queueSettingRepo := repository.NewQueueSettingRepository(db)
	queueJobRepo := repository.NewQueueJobRepository(db)

	localStorage := storage.NewLocalStorage(config.Config.Storage.BasePath)

	redisClient := config.GetRedis()
	redisMultipart := redis.NewMultipartRedis(redisClient)
	cacheService := cache.NewCacheService(redisClient)
	queueDispatchService := service.NewQueueDispatchService(queueSettingRepo, queueJobRepo, fileSystemSettingRepo)
	fileEventService := service.NewFileEventService(fileSystemSettingRepo)

	fileService := service.NewFileService(
		db,
		physicalFileRepo,
		userFileRepo,
		userRepo,
		fileSystemSettingRepo,
		collaborationRepo,
		localStorage,
		redisClient,
		cacheService,
		queueDispatchService,
		fileEventService,
	)

	folderService := service.NewFolderService(
		db,
		userFileRepo,
		cacheService,
		fileEventService,
	)

	emailConfigResolver := func() (service.SMTPConfig, error) {
		setting, err := emailSettingRepo.Get()
		if err != nil {
			return service.SMTPConfig{}, err
		}
		if setting != nil {
			return service.SMTPConfig{
				Enabled:           setting.Enabled,
				Provider:          setting.Provider,
				Host:              setting.Host,
				Port:              setting.Port,
				Username:          setting.Username,
				Password:          setting.Password,
				FromName:          setting.FromName,
				FromAddress:       setting.FromAddress,
				ReplyTo:           setting.ReplyTo,
				ForceSSL:          setting.ForceSSL,
				ConnectionTimeout: setting.ConnectionTimeout,
			}, nil
		}

		return service.SMTPConfig{
			Enabled:           config.Config.Email.Enabled,
			Provider:          config.Config.Email.Provider,
			Host:              config.Config.Email.Host,
			Port:              config.Config.Email.Port,
			Username:          config.Config.Email.Username,
			Password:          config.Config.Email.Password,
			FromName:          config.Config.Email.FromName,
			FromAddress:       config.Config.Email.FromAddress,
			ReplyTo:           config.Config.Email.FromAddress,
			ForceSSL:          false,
			ConnectionTimeout: 30,
		}, nil
	}

	emailTemplateResolver := func(templateKey string) (*service.ResolvedEmailTemplate, error) {
		return service.ResolveEmailTemplate(emailTemplateRepo, templateKey)
	}

	siteContextResolver := func() (service.EmailTemplateCommonContext, error) {
		setting, err := siteSettingRepo.Get()
		if err != nil {
			return service.EmailTemplateCommonContext{}, err
		}

		frontendBaseURL := strings.TrimRight(strings.TrimSpace(config.Config.Server.BaseURL), "/")
		if setting != nil && strings.TrimSpace(setting.PrimaryURL) != "" {
			frontendBaseURL = strings.TrimRight(strings.TrimSpace(setting.PrimaryURL), "/")
		}
		if frontendBaseURL == "" {
			frontendBaseURL = "http://127.0.0.1:4173"
		}
		if strings.HasSuffix(frontendBaseURL, ":8080") {
			frontendBaseURL = strings.TrimSuffix(frontendBaseURL, ":8080") + ":4173"
		}

		siteName := "星云盘"
		logoPath := "/logo192.png"
		if setting != nil {
			if strings.TrimSpace(setting.SiteName) != "" {
				siteName = strings.TrimSpace(setting.SiteName)
			}
			if strings.TrimSpace(setting.LogoLight) != "" {
				logoPath = strings.TrimSpace(setting.LogoLight)
			} else if strings.TrimSpace(setting.Logo192) != "" {
				logoPath = strings.TrimSpace(setting.Logo192)
			}
		}

		logoURL := logoPath
		if strings.HasPrefix(logoURL, "/") {
			logoURL = frontendBaseURL + logoURL
		}

		return service.EmailTemplateCommonContext{
			SiteBasic: service.EmailTemplateSiteBasic{Name: siteName},
			SiteUrl:   frontendBaseURL,
			Logo:      service.EmailTemplateLogo{Normal: logoURL},
		}, nil
	}

	emailSender := service.NewSMTPEmailSenderWithDependencies(emailConfigResolver, emailTemplateResolver, siteContextResolver)

	userService := service.NewUserService(
		userRepo,
		userGroupRepo,
		siteSettingRepo,
		config.Config.JWT.Secret,
		int64(config.Config.JWT.ExpireHours),
		int64(config.Config.JWT.RefreshExpireHours),
		config.Config.User.DefaultCapacity,
		cacheService,
		redisClient,
		emailSender,
		config.Config.Email.CodeTTLSeconds,
		config.Config.Email.SendIntervalSeconds,
	)

	multipartService := service.NewMultipartService(
		db,
		multipartRepo,
		physicalFileRepo,
		userFileRepo,
		userRepo,
		fileSystemSettingRepo,
		localStorage,
		redisMultipart,
		config.Config.Storage.Type,
		5242880,
		24,
		queueDispatchService,
	)

	shareService := service.NewShareService(
		shareRepo,
		userFileRepo,
		redisClient,
		config.Config.JWT.Secret,
		config.Config.Server.BaseURL,
		localStorage,
		db,
	)

	searchService := service.NewSearchService(db, cacheService, fileSystemSettingRepo, fullTextSearchSettingRepo)
	videoService := service.NewVideoService(db, fileSystemSettingRepo)
	musicService := service.NewMusicService(db, fileSystemSettingRepo)
	documentService := service.NewDocumentService(db, fileSystemSettingRepo)

	recycleService := service.NewRecycleService(
		recycleRepo,
		userFileRepo,
		localStorage,
		db,
		queueDispatchService,
	)

	versionService := service.NewVersionService(
		versionRepo,
		userFileRepo,
		physicalFileRepo,
		userRepo,
		cacheService,
		db,
	)

	collaborationService := service.NewCollaborationService(
		collaborationRepo,
		userFileRepo,
		userRepo,
		cacheService,
	)

	siteSettingService := service.NewSiteSettingService(siteSettingRepo)
	avatarService := service.NewAvatarService(siteSettingRepo)
	appearanceSettingService := service.NewAppearanceSettingService(siteSettingRepo)
	captchaSettingService := service.NewCaptchaSettingService(captchaSettingRepo)
	captchaRuntimeService := service.NewCaptchaRuntimeService(captchaSettingRepo, nil)
	mediaSettingService := service.NewMediaSettingService(mediaSettingRepo)
	fileSystemSettingService := service.NewFileSystemSettingService(fileSystemSettingRepo, redisMultipart)
	oauthCredentialService := service.NewOAuthCredentialService(db)
	nodeDispatchService := service.NewNodeDispatchService(nodeRepo)
	fileSystemRuntimeService := service.NewFileSystemRuntimeService(db, fileSystemSettingRepo, userFileRepo, localStorage, config.Config.Server.BaseURL, nodeDispatchService, oauthCredentialService)
	fileCustomPropertyService := service.NewFileCustomPropertyService(userFileRepo, fileSystemSettingRepo, fileCustomPropertyValueRepo)
	fullTextSearchSettingService := service.NewFullTextSearchSettingService(fullTextSearchSettingRepo, queueDispatchService)
	storagePolicyService := service.NewStoragePolicyService(storagePolicyRepo, db)
	if repaired, err := storagePolicyService.RepairLegacyDefaults(service.StoragePolicyActor{Name: "system"}); err != nil {
		logger.Warn("存储策略历史数据修复失败", zap.Error(err))
	} else if len(repaired) > 0 {
		logger.Info("存储策略历史数据已修复", zap.Int("count", len(repaired)))
	}
	nodeService := service.NewNodeService(nodeRepo)
	davAccountService := service.NewDavAccountService(davAccountRepo, config.Config.Server.BaseURL)
	offlineDownloadService := service.NewOfflineDownloadServiceWithSettings(offlineDownloadRepo, fileService, db, shareService, fileSystemSettingService, nodeDispatchService)
	queueExecutor := queue.NewExecutor(
		db,
		localStorage,
		physicalFileRepo,
		userFileRepo,
		multipartRepo,
		recycleRepo,
		offlineDownloadService,
	)
	queueExecutor.SetArchiveExecutor(fileSystemRuntimeService)
	if config.Config.Queue.IsEmbeddedRunnerEnabled() {
		queueRunner := queue.NewRunner(queueSettingRepo, queueJobRepo, queueExecutor, 5*time.Second, fileSystemSettingRepo)
		go queueRunner.Start()
		queue.StartRunnerHeartbeat(context.Background(), config.GetRedis(), "server_embedded", "cmd/server", 5*time.Second)
		logger.Info("内置后台任务 runner 已启动",
			zap.Strings("queues", queue.DescribeImplementedQueues()),
			zap.Bool("embedded_runner_enabled", true),
			zap.String("server_mode", config.Config.Server.Mode),
		)
	} else {
		logger.Info("内置后台任务 runner 已禁用",
			zap.Bool("embedded_runner_enabled", false),
			zap.String("server_mode", config.Config.Server.Mode),
			zap.String("hint", "生产环境建议启动 cmd/worker 独立 runner，避免 server 与 worker 并发叠加"),
		)
	}
	userGroupService := service.NewUserGroupService(userGroupRepo, userRepo, storagePolicyRepo, siteSettingRepo)
	adminUserService := service.NewAdminUserService(userRepo, userGroupRepo, cacheService)
	adminFileService := service.NewAdminFileService(db, userFileRepo, userRepo, fileService, shareService)
	adminBlobService := service.NewAdminBlobService(db, localStorage)
	adminShareService := service.NewAdminShareService(db, shareRepo, config.Config.Server.BaseURL)
	adminOAuthAppService := service.NewAdminOAuthAppService(db)
	oauthSessionService := service.NewOAuthSessionService(db)
	dashboardService := service.NewDashboardService(db)
	emailSettingService := service.NewEmailSettingService(emailSettingRepo, emailSender)
	emailTemplateService := service.NewEmailTemplateService(emailTemplateRepo)
	eventSettingService := service.NewEventSettingService(eventSettingRepo)
	queueSettingService := service.NewQueueSettingService(queueSettingRepo)
	queueStatsService := service.NewQueueStatsService(queueJobRepo)
	queueJobService := service.NewQueueJobService(queueJobRepo, queueSettingRepo)
	queueRuntimeService := service.NewQueueRuntimeService(config.GetRedis())

	healthController := controller.NewHealthController()
	fileController := controller.NewFileController(fileService, fileSystemSettingService, recycleService, fileEventService)
	fileCustomPropertyController := controller.NewFileCustomPropertyController(fileCustomPropertyService)
	folderController := controller.NewFolderController(folderService, recycleService)
	userController := controller.NewUserControllerWithRuntimeServices(userService, avatarService, captchaRuntimeService)
	multipartController := controller.NewMultipartController(multipartService)
	shareController := controller.NewShareController(shareService, fileSystemSettingService)
	searchController := controller.NewSearchController(searchService)
	videoController := controller.NewVideoController(videoService)
	musicController := controller.NewMusicController(musicService)
	documentController := controller.NewDocumentController(documentService)
	recycleController := controller.NewRecycleController(recycleService, fileSystemSettingService)
	versionController := controller.NewVersionController(versionService)
	collaborationController := controller.NewCollaborationController(collaborationService)
	siteSettingController := controller.NewSiteSettingController(siteSettingService)
	appearanceSettingController := controller.NewAppearanceSettingController(appearanceSettingService)
	captchaSettingController := controller.NewCaptchaSettingController(captchaSettingService)
	captchaController := controller.NewCaptchaController(captchaRuntimeService)
	mediaSettingController := controller.NewMediaSettingController(mediaSettingService)
	fileSystemSettingController := controller.NewFileSystemSettingController(fileSystemSettingService)
	fileSystemRuntimeController := controller.NewFileSystemRuntimeController(fileSystemRuntimeService)
	fullTextSearchSettingController := controller.NewFullTextSearchSettingController(fullTextSearchSettingService)
	storagePolicyController := controller.NewStoragePolicyController(storagePolicyService)
	nodeController := controller.NewNodeController(nodeService)
	davAccountController := controller.NewDavAccountController(davAccountService)
	offlineDownloadController := controller.NewOfflineDownloadController(offlineDownloadService, fileSystemSettingService)
	userGroupController := controller.NewUserGroupController(userGroupService)
	adminUserController := controller.NewAdminUserController(adminUserService)
	adminFileController := controller.NewAdminFileController(adminFileService, fileSystemSettingService)
	adminBlobController := controller.NewAdminBlobController(adminBlobService)
	adminShareController := controller.NewAdminShareController(adminShareService, fileSystemSettingService)
	adminOAuthAppController := controller.NewAdminOAuthAppController(adminOAuthAppService)
	adminOAuthCredentialController := controller.NewAdminOAuthCredentialController(oauthCredentialService)
	oauthSessionController := controller.NewOAuthSessionController(oauthSessionService)
	dashboardController := controller.NewDashboardController(dashboardService)
	emailSettingController := controller.NewEmailSettingController(emailSettingService)
	emailTemplateController := controller.NewEmailTemplateController(emailTemplateService)
	eventSettingController := controller.NewEventSettingController(eventSettingService)
	queueController := controller.NewQueueController(queueSettingService, queueStatsService, queueJobService, fileSystemSettingService)
	queueController.SetRuntimeService(queueRuntimeService)

	startOAuthRefreshScheduler(fileSystemRuntimeService)

	router.GET("/health", healthController.Check)
	router.GET("/ping", healthController.Ping)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/api/v1/avatars/*filepath", func(c *gin.Context) {
		localPath, err := service.AvatarFilePathFromURLPath(c.Param("filepath"))
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		absRoot, _ := filepath.Abs("uploads")
		absFile, _ := filepath.Abs(localPath)
		if !strings.HasPrefix(absFile, absRoot+string(os.PathSeparator)) && absFile != absRoot {
			c.Status(http.StatusBadRequest)
			return
		}
		if _, err := os.Stat(absFile); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(absFile)
	})
	router.Any("/dav/:accountToken", davAccountController.DavProbe)

	api := router.Group("/api/v1")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "星云盘 V2 API",
				"version": "1.0.0",
				"phase":   "Phase 4 Complete",
			})
		})

		api.POST("/user/register/email-code", userController.SendRegisterEmailCode)
		api.POST("/user/password/email-code", userController.SendResetPasswordEmailCode)
		api.POST("/user/password/reset", userController.ResetPasswordByEmailCode)
		api.POST("/user/register", userController.Register)
		api.POST("/user/login", userController.Login)
		api.GET("/captcha/config", captchaController.Config)
		api.POST("/captcha/challenge", captchaController.Challenge)
		api.POST("/session/oauth/token", oauthSessionController.Token)
		api.POST("/session/token/refresh", oauthSessionController.Refresh)
		api.GET("/session/oauth/userinfo", oauthSessionController.UserInfo)

		router.GET("/session/authorize", middleware.AuthMiddleware(), oauthSessionController.Authorize)
		apiV4Session := router.Group("/api/v4/session")
		{
			apiV4Session.POST("/oauth/token", oauthSessionController.Token)
			apiV4Session.POST("/token/refresh", oauthSessionController.Refresh)
			apiV4Session.GET("/oauth/userinfo", oauthSessionController.UserInfo)
		}

		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", middleware.RequireOAuthScope("UserInfo.Read", "UserInfo.Write"), userController.GetUserInfo)
			user.PUT("/profile", middleware.RequireOAuthScope("UserInfo.Write"), userController.UpdateProfile)
			user.POST("/avatar", middleware.RequireOAuthScope("UserInfo.Write"), userController.UploadAvatar)
			user.GET("/info", middleware.RequireOAuthScope("UserInfo.Read", "UserInfo.Write"), userController.GetUserInfo)
			user.POST("/change-password", middleware.RequireOAuthScope("UserSecurityInfo.Write"), userController.ChangePassword)
			user.GET("/preferences", middleware.RequireOAuthScope("UserInfo.Read", "UserInfo.Write"), userController.GetPreferences)
			user.PUT("/preferences", middleware.RequireOAuthScope("UserInfo.Write"), userController.UpdatePreferences)
		}

		multipart := api.Group("/files/multipart")
		multipart.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Files.Write"))
		{
			multipart.GET("", multipartController.ListUploadTasks)
			multipart.POST("/init", multipartController.InitMultipartUpload)
			multipart.GET("/:upload_id/urls", multipartController.GetPresignedURLs)
			multipart.POST("/chunk", multipartController.RecordChunkUpload)
			multipart.POST("/chunk/complete", multipartController.RecordChunkUpload)
			multipart.GET("/:upload_id/chunks", multipartController.GetCompletedChunks)
			multipart.POST("/complete", multipartController.CompleteMultipartUpload)
			multipart.DELETE("/:upload_id", multipartController.CancelMultipartUpload)
			multipart.GET("/:upload_id/progress", multipartController.GetUploadProgress)
		}

		multipartCompat := api.Group("/multipart")
		multipartCompat.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Files.Write"))
		{
			multipartCompat.GET("", multipartController.ListUploadTasks)
			multipartCompat.POST("/init", multipartController.InitMultipartUpload)
			multipartCompat.GET("/:upload_id/urls", multipartController.GetPresignedURLs)
			multipartCompat.POST("/chunk", multipartController.RecordChunkUpload)
			multipartCompat.POST("/chunk/complete", multipartController.RecordChunkUpload)
			multipartCompat.GET("/:upload_id/chunks", multipartController.GetCompletedChunks)
			multipartCompat.POST("/complete", multipartController.CompleteMultipartUpload)
			multipartCompat.DELETE("/:upload_id", multipartController.CancelMultipartUpload)
			multipartCompat.GET("/:upload_id/progress", multipartController.GetUploadProgress)
		}

		authorized := api.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			authorized.POST("/file/upload", middleware.RequireOAuthScope("Files.Write"), fileController.Upload)
			authorized.POST("/file/create", middleware.RequireOAuthScope("Files.Write"), fileController.Create)
			authorized.GET("/file/list", middleware.RequireOAuthScope("Files.Read", "Files.Write"), fileController.List)
			authorized.PUT("/file/:id", middleware.RequireOAuthScope("Files.Write"), fileController.Rename)
			authorized.GET("/file/:id/download", middleware.RequireOAuthScope("Files.Read", "Files.Write"), fileController.Download)
			authorized.GET("/file/:id/preview-pdf", middleware.RequireOAuthScope("Files.Read", "Files.Write"), fileController.PreviewPDF)
			authorized.GET("/file/:id/custom-properties", middleware.RequireOAuthScope("Files.Read", "Files.Write"), fileCustomPropertyController.Get)
			authorized.PUT("/file/:id/custom-properties", middleware.RequireOAuthScope("Files.Write"), fileCustomPropertyController.Update)
			authorized.GET("/file/:id/thumbnail", middleware.RequireOAuthScope("Files.Read", "Files.Write"), fileController.Thumbnail)
			authorized.PUT("/file/:id/move", middleware.RequireOAuthScope("Files.Write"), fileController.Move)
			authorized.POST("/file/:id/copy", middleware.RequireOAuthScope("Files.Write"), fileController.Copy)
			authorized.DELETE("/file/:id", middleware.RequireOAuthScope("Files.Write"), fileController.Delete)
			authorized.POST("/file/check", middleware.RequireOAuthScope("Files.Read", "Files.Write"), fileController.Check)
			authorized.GET("/file/events", middleware.RequireOAuthScope("Files.Read", "Files.Write"), fileController.Events)
			authorized.POST("/file/package-downloads", middleware.RequireOAuthScope("Workflow.Write"), fileSystemRuntimeController.CreatePackageDownloadSession)
			authorized.GET("/file/package-downloads/:sessionId/download", middleware.RequireOAuthScope("Workflow.Read", "Workflow.Write"), fileSystemRuntimeController.DownloadPackage)
			authorized.POST("/file/archive/extract", middleware.RequireOAuthScope("Workflow.Write"), fileSystemRuntimeController.ExtractArchive)
			authorized.POST("/file/wopi/sessions", middleware.RequireOAuthScope("Files.Write"), fileSystemRuntimeController.CreateWOPISession)
			authorized.GET("/file/wopi/sessions/:sessionId", middleware.RequireOAuthScope("Files.Read", "Files.Write"), fileSystemRuntimeController.GetWOPISession)
			authorized.GET("/file-system-settings/client", fileSystemSettingController.GetClientSettings)
			authorized.GET("/file-system-settings/map-provider", fileSystemRuntimeController.MapConfig)

			authorized.POST("/folder", middleware.RequireOAuthScope("Files.Write"), folderController.Create)
			authorized.PUT("/folder/:id", middleware.RequireOAuthScope("Files.Write"), folderController.Rename)
			authorized.PUT("/folder/:id/rename", middleware.RequireOAuthScope("Files.Write"), folderController.Rename)
			authorized.DELETE("/folder/:id", middleware.RequireOAuthScope("Files.Write"), folderController.Delete)
			authorized.PUT("/folder/:id/move", middleware.RequireOAuthScope("Files.Write"), folderController.Move)
			authorized.POST("/folder/:id/copy", middleware.RequireOAuthScope("Files.Write"), folderController.Copy)
			authorized.GET("/folder/:id/path", middleware.RequireOAuthScope("Files.Read", "Files.Write"), folderController.Path)
		}

		shares := api.Group("/shares")
		{
			shares.GET("/:shareId", shareController.GetShareInfo)
			shares.POST("/:shareId/verify", middleware.SharePasswordRateLimitMiddleware(), shareController.VerifyPassword)
			shares.GET("/:shareId/download", shareController.Download)
			shares.POST("/:shareId/download", shareController.IncrementDownload)

			shares.Use(middleware.AuthMiddleware())
			shares.POST("", middleware.RequireOAuthScope("Shares.Write"), shareController.CreateShare)
			shares.GET("/me", middleware.RequireOAuthScope("Shares.Read", "Shares.Write"), shareController.GetMyShares)
			shares.DELETE("/:shareId", middleware.RequireOAuthScope("Shares.Write"), shareController.DeleteShare)
		}

		search := api.Group("/search")
		search.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Files.Read", "Files.Write"))
		{
			search.POST("", searchController.SearchFiles)
			search.GET("/files", searchController.SearchFiles)
			search.GET("/suggestions", searchController.GetSuggestions)
		}

		videos := api.Group("/videos")
		videos.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Files.Read", "Files.Write"))
		{
			videos.GET("", videoController.List)
		}

		music := api.Group("/music")
		music.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Files.Read", "Files.Write"))
		{
			music.GET("", musicController.List)
		}

		documents := api.Group("/documents")
		documents.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Files.Read", "Files.Write"))
		{
			documents.GET("", documentController.List)
		}

		recycle := api.Group("/recycle")
		recycle.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Files.Write"))
		{
			recycle.POST("", recycleController.MoveToRecycle)
			recycle.GET("", recycleController.GetRecycleList)
			recycle.POST("/restore", recycleController.RestoreFiles)
			recycle.DELETE("", recycleController.PermanentDelete)
			recycle.DELETE("/all", recycleController.EmptyRecycleBin)
		}

		files := api.Group("/files")
		files.Use(middleware.AuthMiddleware())
		{
			files.GET("/:fileId/versions", middleware.RequireOAuthScope("Files.Read", "Files.Write"), versionController.GetVersionHistory)
			files.GET("/:fileId/versions/:versionId/download", middleware.RequireOAuthScope("Files.Read", "Files.Write"), versionController.DownloadVersion)
			files.POST("/:fileId/versions/:versionId/restore", middleware.RequireOAuthScope("Files.Write"), versionController.RestoreVersion)
			files.DELETE("/:fileId/versions/:versionId", middleware.RequireOAuthScope("Files.Write"), versionController.DeleteVersion)
			files.GET("/:fileId/permissions", middleware.RequireOAuthScope("Shares.Read", "Shares.Write"), collaborationController.CheckFilePermission)
			files.GET("/:fileId/collaborators", middleware.RequireOAuthScope("Shares.Read", "Shares.Write"), collaborationController.GetCollaborators)
			files.PUT("/:fileId/collaborators/:userId", middleware.RequireOAuthScope("Shares.Write"), collaborationController.UpdateCollaboratorPermission)
			files.DELETE("/:fileId/collaborators/:userId", middleware.RequireOAuthScope("Shares.Write"), collaborationController.RemoveCollaborator)
		}

		collaborations := api.Group("/collaborations")
		collaborations.Use(middleware.AuthMiddleware())
		{
			collaborations.POST("", middleware.RequireOAuthScope("Shares.Write"), collaborationController.AddCollaborator)
			collaborations.GET("/me", middleware.RequireOAuthScope("Shares.Read", "Shares.Write"), collaborationController.GetMyCollaborations)
		}

		sharedWithMe := api.Group("/shared-with-me")
		sharedWithMe.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Shares.Read", "Shares.Write"))
		{
			sharedWithMe.GET("", collaborationController.GetSharedWithMe)
		}

		davAccounts := api.Group("/dav/accounts")
		davAccounts.Use(middleware.AuthMiddleware())
		{
			davAccounts.GET("", middleware.RequireOAuthScope("DavAccount.Read", "DavAccount.Write"), davAccountController.List)
			davAccounts.POST("", middleware.RequireOAuthScope("DavAccount.Write"), davAccountController.Create)
			davAccounts.PUT("/:id", middleware.RequireOAuthScope("DavAccount.Write"), davAccountController.Update)
			davAccounts.DELETE("/:id", middleware.RequireOAuthScope("DavAccount.Write"), davAccountController.Delete)
			davAccounts.POST("/:id/secret", middleware.RequireOAuthScope("DavAccount.Write"), davAccountController.ResetSecret)
		}

		offlineDownloads := api.Group("/offline-downloads")
		offlineDownloads.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Workflow.Write"))
		{
			offlineDownloads.GET("", offlineDownloadController.List)
			offlineDownloads.POST("", offlineDownloadController.Create)
			offlineDownloads.POST("/refresh", offlineDownloadController.Refresh)
			offlineDownloads.POST("/batch-delete", offlineDownloadController.BatchDelete)
			offlineDownloads.POST("/:id/pause", offlineDownloadController.Pause)
			offlineDownloads.POST("/:id/resume", offlineDownloadController.Resume)
			offlineDownloads.POST("/:id/retry", offlineDownloadController.Retry)
			offlineDownloads.DELETE("/:id", offlineDownloadController.Delete)
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireOAuthScope("Admin.Read"), middleware.AdminMiddleware())
		{
			admin.GET("/dashboard/overview", dashboardController.Overview)
			admin.GET("/site-settings", siteSettingController.Get)
			admin.PUT("/site-settings", siteSettingController.Update)
			admin.GET("/appearance-settings", appearanceSettingController.Get)
			admin.PUT("/appearance-settings", appearanceSettingController.Update)
			admin.GET("/media-settings", mediaSettingController.Get)
			admin.PUT("/media-settings", mediaSettingController.Update)
			admin.GET("/file-system-settings", fileSystemSettingController.Get)
			admin.PUT("/file-system-settings", fileSystemSettingController.Update)
			admin.PATCH("/file-system-settings/icons", fileSystemSettingController.UpdateIcons)
			admin.POST("/file-system-settings/blob-url-cache/clear", fileSystemSettingController.ClearBlobURLCache)
			admin.GET("/file-system-settings/browser-apps", fileSystemSettingController.GetBrowserApps)
			admin.GET("/file-system-settings/browser-apps/resolve", fileSystemSettingController.ResolveBrowserApp)
			admin.GET("/file-system-settings/master-key/status", fileSystemRuntimeController.MasterKeyStatus)
			admin.POST("/file-system-settings/slave-signature", fileSystemRuntimeController.SignSlaveRequest)
			admin.POST("/file-system-settings/slave-signature/verify", fileSystemRuntimeController.VerifySlaveRequest)
			admin.GET("/file-system-settings/oauth-refresh/status", fileSystemRuntimeController.OAuthRefreshStatus)
			admin.POST("/file-system-settings/oauth-refresh/run", fileSystemRuntimeController.RunOAuthRefresh)
			admin.GET("/full-text-search-settings", fullTextSearchSettingController.Get)
			admin.PUT("/full-text-search-settings", fullTextSearchSettingController.Update)
			admin.POST("/full-text-search-settings/rebuild-index", fullTextSearchSettingController.RebuildIndex)
			admin.GET("/storage-policies", storagePolicyController.List)
			admin.POST("/storage-policies", storagePolicyController.Create)
			admin.POST("/storage-policies/import", storagePolicyController.Import)
			admin.POST("/storage-policies/repair-legacy", storagePolicyController.RepairLegacyDefaults)
			admin.GET("/storage-policies/:id/preview", storagePolicyController.Preview)
			admin.GET("/storage-policies/:id/audits", storagePolicyController.History)
			admin.GET("/storage-policies/:id/hits", storagePolicyController.RecentHits)
			admin.GET("/storage-policies/:id/audits/:auditId", storagePolicyController.AuditDetail)
			admin.POST("/storage-policies/:id/audits/:auditId/rollback", storagePolicyController.Rollback)
			admin.POST("/storage-policies/:id/copy", storagePolicyController.Copy)
			admin.GET("/storage-policies/:id/export", storagePolicyController.Export)
			admin.POST("/storage-policies/:id/migrate-groups", storagePolicyController.MigrateGroups)
			admin.GET("/storage-policies/:id", storagePolicyController.Get)
			admin.PUT("/storage-policies/:id", storagePolicyController.Update)
			admin.DELETE("/storage-policies/:id", storagePolicyController.Delete)
			admin.GET("/nodes", nodeController.List)
			admin.POST("/nodes", nodeController.Create)
			admin.POST("/nodes/test-offline-connectivity", nodeController.TestOfflineConnectivity)
			admin.POST("/nodes/:id/check-health", nodeController.CheckHealth)
			admin.GET("/nodes/:id", nodeController.Get)
			admin.PUT("/nodes/:id", nodeController.Update)
			admin.DELETE("/nodes/:id", nodeController.Delete)
			admin.GET("/user-groups", userGroupController.List)
			admin.GET("/user-groups/summary", userGroupController.Summary)
			admin.GET("/user-groups/:id/users", userGroupController.ListMembers)
			admin.POST("/user-groups", userGroupController.Create)
			admin.PUT("/user-groups/:id", userGroupController.Update)
			admin.DELETE("/user-groups/:id", userGroupController.Delete)
			admin.GET("/users", adminUserController.List)
			admin.POST("/users", adminUserController.Create)
			admin.POST("/users/batch-delete", adminUserController.BatchDelete)
			admin.POST("/users/batch-group", adminUserController.BatchUpdateGroup)
			admin.POST("/users/batch-role", adminUserController.BatchUpdateRole)
			admin.POST("/users/batch-status", adminUserController.BatchUpdateStatus)
			admin.GET("/users/:id/delete-preview", adminUserController.DeletePreview)
			admin.PUT("/users/:id", adminUserController.Update)
			admin.PUT("/users/:id/status", adminUserController.UpdateStatus)
			admin.DELETE("/users/:id", adminUserController.Delete)
			admin.POST("/users/:id/reset-password", adminUserController.ResetPassword)
			admin.GET("/files", adminFileController.List)
			admin.GET("/files/:id", adminFileController.Get)
			admin.POST("/files/import", adminFileController.Import)
			admin.PUT("/files/:id", adminFileController.Rename)
			admin.GET("/files/:id/download", adminFileController.Download)
			admin.POST("/files/:id/share", adminFileController.CreateShare)
			admin.DELETE("/files/:id", adminFileController.Delete)
			admin.GET("/blobs", adminBlobController.List)
			admin.POST("/blobs/scan", adminBlobController.Scan)
			admin.GET("/blobs/scan/latest", adminBlobController.LatestScan)
			admin.POST("/blobs/batch-delete", adminBlobController.BatchDelete)
			admin.GET("/blobs/:id", adminBlobController.Get)
			admin.GET("/blobs/:id/download", adminBlobController.Download)
			admin.POST("/blobs/:id/lock", adminBlobController.Lock)
			admin.POST("/blobs/:id/unlock", adminBlobController.Unlock)
			admin.DELETE("/blobs/:id", adminBlobController.Delete)
			admin.GET("/shares/metrics", adminShareController.Metrics)
			admin.GET("/shares", adminShareController.List)
			admin.POST("/shares/batch-delete", adminShareController.BatchDelete)
			admin.DELETE("/shares/:id", adminShareController.Delete)
			admin.GET("/oauth-apps", adminOAuthAppController.List)
			admin.POST("/oauth-apps", adminOAuthAppController.Create)
			admin.GET("/oauth-apps/:id", adminOAuthAppController.Get)
			admin.PUT("/oauth-apps/:id", adminOAuthAppController.Update)
			admin.PUT("/oauth-apps/:id/status", adminOAuthAppController.UpdateStatus)
			admin.POST("/oauth-apps/:id/secret", adminOAuthAppController.RegenerateSecret)
			admin.DELETE("/oauth-apps/:id", adminOAuthAppController.Delete)
			admin.GET("/oauth-credentials", adminOAuthCredentialController.List)
			admin.POST("/oauth-credentials", adminOAuthCredentialController.Create)
			admin.POST("/oauth-credentials/:id/refresh", adminOAuthCredentialController.Refresh)
			admin.GET("/captcha-settings", captchaSettingController.Get)
			admin.PUT("/captcha-settings", captchaSettingController.Update)
			admin.GET("/email-settings", emailSettingController.Get)
			admin.PUT("/email-settings", emailSettingController.Update)
			admin.POST("/email-settings/test", emailSettingController.SendTestEmail)
			admin.GET("/email-templates", emailTemplateController.List)
			admin.PUT("/email-templates/:templateKey", emailTemplateController.Update)
			admin.GET("/event-settings", eventSettingController.Get)
			admin.PUT("/event-settings", eventSettingController.Update)
			admin.POST("/event-settings/reset", eventSettingController.Reset)
			admin.POST("/event-settings/toggle-all", eventSettingController.ToggleAll)
			admin.POST("/event-settings/categories/:categoryKey", eventSettingController.ToggleCategory)
			admin.PATCH("/event-settings/events/:eventKey", eventSettingController.ToggleEvent)
			admin.GET("/queue-settings", queueController.GetSettings)
			admin.PUT("/queue-settings", queueController.UpdateSettings)
			admin.GET("/queue-stats", queueController.GetStats)
			admin.GET("/queue-runtime", queueController.GetRuntime)
			admin.GET("/queue-jobs", queueController.GetJobs)
			admin.POST("/queue-jobs/recover-stale", queueController.RecoverStaleJobs)
			admin.POST("/queue-jobs/batch-delete", queueController.BatchDeleteJobs)
			admin.POST("/queue-jobs/clear", queueController.ClearJobs)
			admin.GET("/queue-jobs/:id", queueController.GetJob)
			admin.POST("/queue-jobs/:id/retry", queueController.RetryJob)
			admin.DELETE("/queue-jobs/:id", queueController.DeleteJob)
		}
	}

	legacyMultipart := router.Group("/api/multipart")
	legacyMultipart.Use(middleware.AuthMiddleware())
	{
		legacyMultipart.GET("", multipartController.ListUploadTasks)
		legacyMultipart.POST("/init", multipartController.InitMultipartUpload)
		legacyMultipart.GET("/:upload_id/urls", multipartController.GetPresignedURLs)
		legacyMultipart.POST("/chunk", multipartController.RecordChunkUpload)
		legacyMultipart.POST("/chunk/complete", multipartController.RecordChunkUpload)
		legacyMultipart.GET("/:upload_id/chunks", multipartController.GetCompletedChunks)
		legacyMultipart.POST("/complete", multipartController.CompleteMultipartUpload)
		legacyMultipart.DELETE("/:upload_id", multipartController.CancelMultipartUpload)
		legacyMultipart.GET("/:upload_id/progress", multipartController.GetUploadProgress)
	}

	legacyFiles := router.Group("/api/files")
	legacyFiles.Use(middleware.AuthMiddleware())
	{
		legacyFiles.GET("/:fileId/permissions", collaborationController.CheckFilePermission)
		legacyFiles.GET("/:fileId/collaborators", collaborationController.GetCollaborators)
		legacyFiles.PUT("/:fileId/collaborators/:userId", collaborationController.UpdateCollaboratorPermission)
		legacyFiles.DELETE("/:fileId/collaborators/:userId", collaborationController.RemoveCollaborator)
	}

	legacyCollaborations := router.Group("/api/collaborations")
	legacyCollaborations.Use(middleware.AuthMiddleware())
	{
		legacyCollaborations.POST("", collaborationController.AddCollaborator)
		legacyCollaborations.GET("/me", collaborationController.GetMyCollaborations)
	}

	return router
}
