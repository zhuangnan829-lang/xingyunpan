// 路径: internal/worker/file_deletion_worker.go
package worker

import (
	"fmt"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/storage"

	"go.uber.org/zap"
)

// FileDeletionWorker 文件删除 Worker
type FileDeletionWorker struct {
	deletionRepo repository.FileDeletionRepository
	storage      storage.Storage
	interval     time.Duration
	batchSize    int
	maxRetries   int
	stopChan     chan struct{}
}

// NewFileDeletionWorker 创建文件删除 Worker 实例
func NewFileDeletionWorker(
	deletionRepo repository.FileDeletionRepository,
	storage storage.Storage,
	interval time.Duration,
	batchSize int,
	maxRetries int,
) *FileDeletionWorker {
	return &FileDeletionWorker{
		deletionRepo: deletionRepo,
		storage:      storage,
		interval:     interval,
		batchSize:    batchSize,
		maxRetries:   maxRetries,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动 Worker
func (w *FileDeletionWorker) Start() {
	logger.Info("文件删除 Worker 已启动", zap.Duration("interval", w.interval))
	
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// 立即执行一次
	w.process()

	for {
		select {
		case <-ticker.C:
			w.process()
		case <-w.stopChan:
			logger.Info("文件删除 Worker 已停止")
			return
		}
	}
}

// Stop 停止 Worker
func (w *FileDeletionWorker) Stop() {
	close(w.stopChan)
}

// process 处理删除队列
func (w *FileDeletionWorker) process() {
	// 获取待处理的删除任务
	deletions, err := w.deletionRepo.GetPendingDeletions(w.batchSize)
	if err != nil {
		logger.Error("获取待删除文件失败", zap.Error(err))
		return
	}

	if len(deletions) == 0 {
		return
	}

	logger.Info("开始处理文件删除", zap.Int("count", len(deletions)))

	successCount := 0
	failCount := 0

	for _, deletion := range deletions {
		if err := w.processDeletion(&deletion); err != nil {
			logger.Error("删除文件失败",
				zap.Uint("id", deletion.ID),
				zap.String("path", deletion.StoragePath),
				zap.Error(err))
			failCount++
		} else {
			successCount++
		}
	}

	logger.Info("文件删除完成",
		zap.Int("success", successCount),
		zap.Int("fail", failCount))
}

// processDeletion 处理单个删除任务
func (w *FileDeletionWorker) processDeletion(deletion *model.FileDeletion) error {
	// 更新状态为处理中
	if err := w.deletionRepo.UpdateStatus(deletion.ID, model.DeletionStatusProcessing); err != nil {
		return fmt.Errorf("更新状态失败: %w", err)
	}

	// 删除物理文件
	if err := w.storage.Delete(deletion.StoragePath); err != nil {
		// 删除失败,增加重试次数
		if err := w.deletionRepo.IncrementRetryCount(deletion.ID); err != nil {
			logger.Error("增加重试次数失败", zap.Uint("id", deletion.ID), zap.Error(err))
		}

		// 检查是否超过最大重试次数
		if deletion.RetryCount+1 >= w.maxRetries {
			// 标记为失败
			if err := w.deletionRepo.UpdateStatus(deletion.ID, model.DeletionStatusFailed); err != nil {
				logger.Error("更新失败状态失败", zap.Uint("id", deletion.ID), zap.Error(err))
			}
			return fmt.Errorf("删除失败,已达最大重试次数: %w", err)
		}

		// 重置为待处理状态,等待下次重试
		if err := w.deletionRepo.UpdateStatus(deletion.ID, model.DeletionStatusPending); err != nil {
			logger.Error("重置状态失败", zap.Uint("id", deletion.ID), zap.Error(err))
		}
		return fmt.Errorf("删除失败,将重试: %w", err)
	}

	// 删除成功,更新状态并删除记录
	if err := w.deletionRepo.UpdateStatus(deletion.ID, model.DeletionStatusCompleted); err != nil {
		logger.Error("更新完成状态失败", zap.Uint("id", deletion.ID), zap.Error(err))
	}

	// 删除任务记录
	if err := w.deletionRepo.Delete(deletion.ID); err != nil {
		logger.Error("删除任务记录失败", zap.Uint("id", deletion.ID), zap.Error(err))
	}

	return nil
}
