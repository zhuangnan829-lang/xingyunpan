// 路径: internal/repository/multipart_upload_repository.go
package repository

import (
	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// MultipartUploadRepository 分片上传仓库接口
type MultipartUploadRepository interface {
	Create(upload *model.MultipartUpload) error
	GetByUploadID(uploadID string) (*model.MultipartUpload, error)
	GetByUserID(userID uint, status string, page, pageSize int) ([]model.MultipartUpload, int64, error)
	UpdateStatus(uploadID string, status string) error
	Delete(uploadID string) error
	GetExpiredUploads(hours int) ([]model.MultipartUpload, error)
}

type multipartUploadRepository struct {
	db *gorm.DB
}

// NewMultipartUploadRepository 创建分片上传仓库实例
func NewMultipartUploadRepository(db *gorm.DB) MultipartUploadRepository {
	return &multipartUploadRepository{db: db}
}

// Create 创建分片上传任务
func (r *multipartUploadRepository) Create(upload *model.MultipartUpload) error {
	return r.db.Create(upload).Error
}

// GetByUploadID 根据上传ID获取任务
func (r *multipartUploadRepository) GetByUploadID(uploadID string) (*model.MultipartUpload, error) {
	var upload model.MultipartUpload
	err := r.db.Where("upload_id = ?", uploadID).First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}

// GetByUserID 根据用户ID获取任务列表
func (r *multipartUploadRepository) GetByUserID(userID uint, status string, page, pageSize int) ([]model.MultipartUpload, int64, error) {
	var uploads []model.MultipartUpload
	var total int64

	query := r.db.Model(&model.MultipartUpload{}).Where("user_id = ?", userID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&uploads).Error

	return uploads, total, err
}

// UpdateStatus 更新任务状态
func (r *multipartUploadRepository) UpdateStatus(uploadID string, status string) error {
	return r.db.Model(&model.MultipartUpload{}).
		Where("upload_id = ?", uploadID).
		Update("status", status).Error
}

// Delete 删除任务
func (r *multipartUploadRepository) Delete(uploadID string) error {
	return r.db.Where("upload_id = ?", uploadID).Delete(&model.MultipartUpload{}).Error
}

// GetExpiredUploads 获取过期的上传任务
func (r *multipartUploadRepository) GetExpiredUploads(hours int) ([]model.MultipartUpload, error) {
	var uploads []model.MultipartUpload
	err := r.db.Where("status = ? AND created_at < DATE_SUB(NOW(), INTERVAL ? HOUR)", 
		model.MultipartStatusUploading, hours).
		Find(&uploads).Error
	return uploads, err
}
