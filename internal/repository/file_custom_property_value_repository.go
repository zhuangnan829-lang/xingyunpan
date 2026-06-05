package repository

import (
	"fmt"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// FileCustomPropertyValueRepository handles persistence of per-file custom property values.
type FileCustomPropertyValueRepository interface {
	GetByFileID(fileID uint) (*model.FileCustomPropertyValue, error)
	Save(value *model.FileCustomPropertyValue) error
}

type fileCustomPropertyValueRepository struct {
	db *gorm.DB
}

// NewFileCustomPropertyValueRepository creates a file custom property value repository.
func NewFileCustomPropertyValueRepository(db *gorm.DB) FileCustomPropertyValueRepository {
	return &fileCustomPropertyValueRepository{db: db}
}

func (r *fileCustomPropertyValueRepository) GetByFileID(fileID uint) (*model.FileCustomPropertyValue, error) {
	var value model.FileCustomPropertyValue
	if err := r.db.Where("file_id = ?", fileID).First(&value).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query file custom property values failed: %w", err)
	}

	return &value, nil
}

func (r *fileCustomPropertyValueRepository) Save(value *model.FileCustomPropertyValue) error {
	if value == nil {
		return fmt.Errorf("file custom property value cannot be nil")
	}

	if value.ID == 0 {
		if err := r.db.Create(value).Error; err != nil {
			return fmt.Errorf("create file custom property values failed: %w", err)
		}
		return nil
	}

	if err := r.db.Save(value).Error; err != nil {
		return fmt.Errorf("update file custom property values failed: %w", err)
	}

	return nil
}
