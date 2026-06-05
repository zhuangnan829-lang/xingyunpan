package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// FileSystemSettingRepository handles persistence for singleton file system settings.
type FileSystemSettingRepository interface {
	Get() (*model.FileSystemSetting, error)
	Save(setting *model.FileSystemSetting) error
}

type fileSystemSettingRepository struct {
	db *gorm.DB
}

// NewFileSystemSettingRepository creates a file system settings repository.
func NewFileSystemSettingRepository(db *gorm.DB) FileSystemSettingRepository {
	return &fileSystemSettingRepository{db: db}
}

// Get returns the current singleton file system settings row if it exists.
func (r *fileSystemSettingRepository) Get() (*model.FileSystemSetting, error) {
	var setting model.FileSystemSetting
	if err := r.db.Order("id asc").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query file system settings failed: %w", err)
	}

	return &setting, nil
}

// Save creates or updates the singleton file system settings row.
func (r *fileSystemSettingRepository) Save(setting *model.FileSystemSetting) error {
	if setting == nil {
		return fmt.Errorf("file system settings cannot be nil")
	}

	if setting.ID == 0 {
		if err := r.db.Create(setting).Error; err != nil {
			return fmt.Errorf("create file system settings failed: %w", err)
		}
		return nil
	}

	if err := r.db.Save(setting).Error; err != nil {
		return fmt.Errorf("update file system settings failed: %w", err)
	}

	return nil
}
