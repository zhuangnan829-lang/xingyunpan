// 路径: internal/repository/file_deletion_repository.go
package repository

import (
	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// FileDeletionRepository 文件删除队列仓库接口
type FileDeletionRepository interface {
	Create(deletion *model.FileDeletion) error
	GetPendingDeletions(limit int) ([]model.FileDeletion, error)
	UpdateStatus(id uint, status string) error
	IncrementRetryCount(id uint) error
	Delete(id uint) error
}

type fileDeletionRepository struct {
	db *gorm.DB
}

// NewFileDeletionRepository 创建文件删除队列仓库实例
func NewFileDeletionRepository(db *gorm.DB) FileDeletionRepository {
	return &fileDeletionRepository{db: db}
}

// Create 创建删除任务
func (r *fileDeletionRepository) Create(deletion *model.FileDeletion) error {
	return r.db.Create(deletion).Error
}

// GetPendingDeletions 获取待处理的删除任务
func (r *fileDeletionRepository) GetPendingDeletions(limit int) ([]model.FileDeletion, error) {
	var deletions []model.FileDeletion
	err := r.db.Where("status = ?", model.DeletionStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&deletions).Error
	return deletions, err
}

// UpdateStatus 更新删除任务状态
func (r *fileDeletionRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.FileDeletion{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// IncrementRetryCount 增加重试次数
func (r *fileDeletionRepository) IncrementRetryCount(id uint) error {
	return r.db.Model(&model.FileDeletion{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error
}

// Delete 删除任务记录
func (r *fileDeletionRepository) Delete(id uint) error {
	return r.db.Delete(&model.FileDeletion{}, id).Error
}
