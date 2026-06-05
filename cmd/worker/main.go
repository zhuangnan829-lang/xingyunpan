package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/queue"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/storage"

	"go.uber.org/zap"
)

func main() {
	fmt.Println("=== Xingyunpan V2 Unified Queue Worker ===")

	if err := logger.Init(&logger.Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}); err != nil {
		fmt.Printf("init logger failed: %v\n", err)
		os.Exit(1)
	}

	if err := config.LoadDefault(); err != nil {
		logger.Fatal("load config failed", zap.Error(err))
	}
	if !config.Config.Worker.IsEnabled() {
		logger.Info("unified queue worker disabled by config", zap.Bool("worker_enabled", false))
		return
	}
	if err := config.InitDatabase(); err != nil {
		logger.Fatal("init database failed", zap.Error(err))
	}
	defer config.CloseDatabase()

	if err := config.InitRedis(); err != nil {
		logger.Fatal("init redis failed", zap.Error(err))
	}
	defer config.CloseRedis()

	var stor storage.MultipartStorage
	if config.Config.Storage.Type == "local" {
		stor = storage.NewLocalStorage(config.Config.Storage.BasePath)
	} else {
		logger.Fatal("unsupported storage type", zap.String("type", config.Config.Storage.Type))
	}

	db := config.GetDB()
	queueSettingRepo := repository.NewQueueSettingRepository(db)
	queueJobRepo := repository.NewQueueJobRepository(db)
	fileSystemSettingRepo := repository.NewFileSystemSettingRepository(db)
	physicalFileRepo := repository.NewPhysicalFileRepository(db)
	userFileRepo := repository.NewUserFileRepository(db)
	multipartRepo := repository.NewMultipartUploadRepository(db)
	recycleRepo := repository.NewRecycleRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	offlineDownloadRepo := repository.NewOfflineDownloadTaskRepository(db)
	nodeDispatchService := service.NewNodeDispatchService(nodeRepo)
	fileSystemSettingService := service.NewFileSystemSettingService(fileSystemSettingRepo, nil)
	fileSystemRuntimeService := service.NewFileSystemRuntimeService(db, fileSystemSettingRepo, userFileRepo, stor, config.Config.Server.BaseURL, nodeDispatchService)
	offlineDownloadService := service.NewOfflineDownloadServiceWithSettings(offlineDownloadRepo, nil, db, nil, fileSystemSettingService, nodeDispatchService)

	executor := queue.NewExecutor(
		db,
		stor,
		physicalFileRepo,
		userFileRepo,
		multipartRepo,
		recycleRepo,
		offlineDownloadService,
	)
	executor.SetArchiveExecutor(fileSystemRuntimeService)
	runner := queue.NewRunner(queueSettingRepo, queueJobRepo, executor, 5*time.Second, fileSystemSettingRepo)

	ctx, cancelHeartbeat := context.WithCancel(context.Background())
	defer cancelHeartbeat()
	queue.StartRunnerHeartbeat(ctx, config.GetRedis(), "worker", "cmd/worker", 5*time.Second)
	go runner.Start()

	logger.Info("unified queue worker started",
		zap.Bool("worker_enabled", true),
		zap.Strings("queues", queue.DescribeImplementedQueues()),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("stopping unified queue worker")
	cancelHeartbeat()
	runner.Stop()
	time.Sleep(2 * time.Second)

	logger.Info("unified queue worker stopped")
}
