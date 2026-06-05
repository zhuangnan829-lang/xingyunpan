package worker

import (
	"context"
	"log"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/metrics"

	"gorm.io/gorm"
)

// MetricsCollector 业务指标收集器
type MetricsCollector struct {
	db *gorm.DB
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(db *gorm.DB) *MetricsCollector {
	return &MetricsCollector{
		db: db,
	}
}

// Start 启动指标收集器（每分钟更新一次）
func (mc *MetricsCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// 立即执行一次
	mc.collectMetrics()

	for {
		select {
		case <-ticker.C:
			mc.collectMetrics()
		case <-ctx.Done():
			log.Println("Metrics collector stopped")
			return
		}
	}
}

// collectMetrics 收集业务指标
func (mc *MetricsCollector) collectMetrics() {
	// 收集总用户数
	var totalUsers int64
	if err := mc.db.Model(&model.User{}).Count(&totalUsers).Error; err != nil {
		log.Printf("Failed to count users: %v", err)
	} else {
		metrics.SetTotalUsers(float64(totalUsers))
	}

	// 收集总文件数（未删除的用户文件）
	var totalFiles int64
	if err := mc.db.Model(&model.UserFile{}).Where("deleted_at IS NULL").Count(&totalFiles).Error; err != nil {
		log.Printf("Failed to count files: %v", err)
	} else {
		metrics.SetTotalFiles(float64(totalFiles))
	}

	// 收集总存储使用量（物理文件大小总和）
	var totalStorage struct {
		Total int64
	}
	if err := mc.db.Model(&model.PhysicalFile{}).
		Select("COALESCE(SUM(file_size), 0) as total").
		Scan(&totalStorage).Error; err != nil {
		log.Printf("Failed to calculate total storage: %v", err)
	} else {
		metrics.SetTotalStorageBytes(float64(totalStorage.Total))
	}

	log.Printf("Metrics collected - Users: %d, Files: %d, Storage: %d bytes",
		totalUsers, totalFiles, totalStorage.Total)
}
