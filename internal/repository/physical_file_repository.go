// 路径: internal/repository/physical_file_repository.go
package repository

import (
	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// PhysicalFileRepository 物理文件仓储接口
type PhysicalFileRepository interface {
	Create(file *model.PhysicalFile) error
	GetByID(id uint) (*model.PhysicalFile, error)
	GetByHash(hash string) (*model.PhysicalFile, error)
	UpdateRefCount(id uint, delta int) error
	IncrementRefCount(id uint) error
	Delete(id uint) error
}

type physicalFileRepository struct {
	db *gorm.DB
}

func NewPhysicalFileRepository(db *gorm.DB) PhysicalFileRepository {
	return &physicalFileRepository{db: db}
}

func (r *physicalFileRepository) Create(file *model.PhysicalFile) error {
	return r.db.Create(file).Error
}

func (r *physicalFileRepository) GetByID(id uint) (*model.PhysicalFile, error) {
	var file model.PhysicalFile
	if err := r.db.First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *physicalFileRepository) GetByHash(hash string) (*model.PhysicalFile, error) {
	var file model.PhysicalFile
	if err := r.db.Where("file_hash = ?", hash).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *physicalFileRepository) UpdateRefCount(id uint, delta int) error {
	return r.db.Model(&model.PhysicalFile{}).Where("id = ?", id).
		UpdateColumn("ref_count", gorm.Expr("ref_count + ?", delta)).Error
}

func (r *physicalFileRepository) IncrementRefCount(id uint) error {
	return r.UpdateRefCount(id, 1)
}

func (r *physicalFileRepository) Delete(id uint) error {
	return r.db.Delete(&model.PhysicalFile{}, id).Error
}
