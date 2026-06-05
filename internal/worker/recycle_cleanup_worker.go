// 路径: internal/worker/recycle_cleanup_worker.go
package worker

import (
	"context"
	"time"

	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/storage"

	"go.uber.org/zap"
)

// RecycleCleanupWorker 回收站自动清理 Worker
// 定时清理 expires_at 已过期的回收站项目
type RecycleCleanupWorker struct {
	recycleRepo repository.RecycleRepository
	fileRepo    repository.UserFileRepository
	storage     storage.Storage
	cronHour    int // 每天执行的小时（0-23），默认 2（凌晨 2 点）
	batchSize   int // 每次处理的最大数量
	stopChan    chan struct{}
}

// NewRecycleCleanupWorker 创建回收站清理 Worker 实例
// cronHour: 每天执行的小时（0-23），默认 2
// batchSize: 每次批量处理的最大数量，默认 100
func NewRecycleCleanupWorker(
	recycleRepo repository.RecycleRepository,
	fileRepo repository.UserFileRepository,
	stor storage.Storage,
	cronHour int,
	batchSize int,
) *RecycleCleanupWorker {
	if cronHour < 0 || cronHour > 23 {
		cronHour = 2 // 默认凌晨 2 点
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	return &RecycleCleanupWorker{
		recycleRepo: recycleRepo,
		fileRepo:    fileRepo,
		storage:     stor,
		cronHour:    cronHour,
		batchSize:   batchSize,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动 Worker，每天在指定小时执行一次清理
func (w *RecycleCleanupWorker) Start() {
	logger.Info("回收站清理 Worker 已启动",
		zap.Int("cron_hour", w.cronHour),
		zap.Int("batch_size", w.batchSize))

	for {
		// 计算距离下次执行的等待时间
		waitDuration := w.durationUntilNextRun()
		logger.Info("回收站清理 Worker 等待下次执行",
			zap.Duration("wait", waitDuration),
			zap.Time("next_run", time.Now().Add(waitDuration)))

		select {
		case <-time.After(waitDuration):
			w.process()
		case <-w.stopChan:
			logger.Info("回收站清理 Worker 已停止")
			return
		}
	}
}

// Stop 停止 Worker
func (w *RecycleCleanupWorker) Stop() {
	close(w.stopChan)
}

// durationUntilNextRun 计算距离下次执行的时间
func (w *RecycleCleanupWorker) durationUntilNextRun() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), w.cronHour, 0, 0, 0, now.Location())
	if !next.After(now) {
		// 今天的执行时间已过，等到明天
		next = next.Add(24 * time.Hour)
	}
	return next.Sub(now)
}

// process 执行一次清理，批量处理过期的回收站项目
func (w *RecycleCleanupWorker) process() {
	logger.Info("开始执行回收站过期项目清理")

	ctx := context.Background()
	totalDeleted := 0
	totalFailed := 0

	// 循环批量处理，直到没有更多过期项目
	for {
		items, err := w.recycleRepo.GetExpiredItems(ctx, w.batchSize)
		if err != nil {
			logger.Error("查询过期回收站项目失败", zap.Error(err))
			break
		}

		if len(items) == 0 {
			break
		}

		logger.Info("处理过期回收站批次", zap.Int("count", len(items)))

		successIDs := make([]uint, 0, len(items))
		for _, item := range items {
			// 删除物理文件
			if err := w.deletePhysicalFile(ctx, item.FileID); err != nil {
				logger.Warn("删除物理文件失败",
					zap.Uint("recycle_id", item.ID),
					zap.Uint("file_id", item.FileID),
					zap.Error(err))
				// 物理文件删除失败不阻止数据库记录清理
			}
			successIDs = append(successIDs, item.ID)
		}

		// 批量删除回收站记录
		if len(successIDs) > 0 {
			if err := w.recycleRepo.BatchDelete(ctx, successIDs); err != nil {
				logger.Error("批量删除回收站记录失败",
					zap.Int("count", len(successIDs)),
					zap.Error(err))
				totalFailed += len(successIDs)
			} else {
				totalDeleted += len(successIDs)
			}
		}

		// 如果本批次数量小于 batchSize，说明已处理完毕
		if len(items) < w.batchSize {
			break
		}
	}

	logger.Info("回收站过期项目清理完成",
		zap.Int("deleted", totalDeleted),
		zap.Int("failed", totalFailed))
}

// deletePhysicalFile 删除文件对应的物理文件和数据库记录
func (w *RecycleCleanupWorker) deletePhysicalFile(ctx context.Context, fileID uint) error {
	// 查询文件信息（含物理文件）
	file, err := w.fileRepo.GetByIDWithPhysicalFile(fileID)
	if err != nil {
		// 文件记录可能已不存在，忽略
		return nil
	}

	// 删除物理文件
	if file.PhysicalFile != nil && file.PhysicalFile.StoragePath != "" {
		if err := w.storage.Delete(file.PhysicalFile.StoragePath); err != nil {
			logger.Warn("删除物理存储文件失败",
				zap.Uint("file_id", fileID),
				zap.String("path", file.PhysicalFile.StoragePath),
				zap.Error(err))
		}
	}

	return nil
}
