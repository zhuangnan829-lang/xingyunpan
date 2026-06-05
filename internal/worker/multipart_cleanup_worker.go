// 路径: internal/worker/multipart_cleanup_worker.go
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

// MultipartCleanupWorker 分片清理 Worker
type MultipartCleanupWorker struct {
	multipartRepo repository.MultipartUploadRepository
	storage       storage.MultipartStorage
	interval      time.Duration
	expireHours   int
	stopChan      chan struct{}
}

// NewMultipartCleanupWorker 创建分片清理 Worker 实例
func NewMultipartCleanupWorker(
	multipartRepo repository.MultipartUploadRepository,
	storage storage.MultipartStorage,
	interval time.Duration,
	expireHours int,
) *MultipartCleanupWorker {
	return &MultipartCleanupWorker{
		multipartRepo: multipartRepo,
		storage:       storage,
		interval:      interval,
		expireHours:   expireHours,
		stopChan:      make(chan struct{}),
	}
}

// Start 启动 Worker
func (w *MultipartCleanupWorker) Start() {
	logger.Info("分片清理 Worker 已启动",
		zap.Duration("interval", w.interval),
		zap.Int("expire_hours", w.expireHours))
	
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// 立即执行一次
	w.process()

	for {
		select {
		case <-ticker.C:
			w.process()
		case <-w.stopChan:
			logger.Info("分片清理 Worker 已停止")
			return
		}
	}
}

// Stop 停止 Worker
func (w *MultipartCleanupWorker) Stop() {
	close(w.stopChan)
}

// process 处理过期的分片上传任务
func (w *MultipartCleanupWorker) process() {
	// 获取过期的上传任务
	uploads, err := w.multipartRepo.GetExpiredUploads(w.expireHours)
	if err != nil {
		logger.Error("获取过期上传任务失败", zap.Error(err))
		return
	}

	if len(uploads) == 0 {
		return
	}

	logger.Info("开始清理过期分片", zap.Int("count", len(uploads)))

	successCount := 0
	failCount := 0

	for _, upload := range uploads {
		if err := w.cleanupUpload(&upload); err != nil {
			logger.Error("清理分片失败",
				zap.String("upload_id", upload.UploadID),
				zap.Error(err))
			failCount++
		} else {
			successCount++
		}
	}

	logger.Info("分片清理完成",
		zap.Int("success", successCount),
		zap.Int("fail", failCount))
}

// cleanupUpload 清理单个上传任务
func (w *MultipartCleanupWorker) cleanupUpload(upload *model.MultipartUpload) error {
	// 删除所有分片文件
	chunkPaths := make([]string, upload.TotalChunks)
	for i := 1; i <= upload.TotalChunks; i++ {
		chunkPaths[i-1] = fmt.Sprintf("multipart/%s/chunk_%d", upload.UploadID, i)
	}

	if err := w.storage.DeleteChunks(chunkPaths); err != nil {
		logger.Warn("删除分片文件失败",
			zap.String("upload_id", upload.UploadID),
			zap.Error(err))
		// 继续执行,不返回错误
	}

	// 删除数据库记录
	if err := w.multipartRepo.Delete(upload.UploadID); err != nil {
		return fmt.Errorf("删除数据库记录失败: %w", err)
	}

	logger.Info("清理分片成功",
		zap.String("upload_id", upload.UploadID),
		zap.String("file_name", upload.FileName))

	return nil
}
