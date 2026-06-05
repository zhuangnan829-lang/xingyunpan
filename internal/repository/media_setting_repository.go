package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// MediaSettingRepository handles persistence for singleton media settings.
type MediaSettingRepository interface {
	Get() (*model.MediaSetting, error)
	Save(setting *model.MediaSetting) error
}

type mediaSettingRepository struct {
	db *gorm.DB
}

// NewMediaSettingRepository creates a media settings repository.
func NewMediaSettingRepository(db *gorm.DB) MediaSettingRepository {
	return &mediaSettingRepository{db: db}
}

// Get returns the current singleton media settings row if it exists.
func (r *mediaSettingRepository) Get() (*model.MediaSetting, error) {
	var setting model.MediaSetting
	if err := r.db.Order("id asc").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query media settings failed: %w", err)
	}

	return &setting, nil
}

// Save creates or updates the singleton media settings row.
func (r *mediaSettingRepository) Save(setting *model.MediaSetting) error {
	if setting == nil {
		return fmt.Errorf("media settings cannot be nil")
	}

	if setting.ID == 0 {
		if err := r.db.Create(setting).Error; err != nil {
			return fmt.Errorf("create media settings failed: %w", err)
		}
		return nil
	}

	if err := r.db.Save(setting).Error; err != nil {
		return fmt.Errorf("update media settings failed: %w", err)
	}

	return nil
}
