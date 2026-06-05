package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// FullTextSearchSettingRepository handles persistence for singleton full text search settings.
type FullTextSearchSettingRepository interface {
	Get() (*model.FullTextSearchSetting, error)
	Save(setting *model.FullTextSearchSetting) error
}

type fullTextSearchSettingRepository struct {
	db *gorm.DB
}

// NewFullTextSearchSettingRepository creates a repository instance.
func NewFullTextSearchSettingRepository(db *gorm.DB) FullTextSearchSettingRepository {
	return &fullTextSearchSettingRepository{db: db}
}

// Get returns the current singleton row if it exists.
func (r *fullTextSearchSettingRepository) Get() (*model.FullTextSearchSetting, error) {
	var setting model.FullTextSearchSetting
	if err := r.db.Order("id asc").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query full text search settings failed: %w", err)
	}

	return &setting, nil
}

// Save creates or updates the singleton row.
func (r *fullTextSearchSettingRepository) Save(setting *model.FullTextSearchSetting) error {
	if setting == nil {
		return fmt.Errorf("full text search settings cannot be nil")
	}

	if setting.ID == 0 {
		if err := r.db.Create(setting).Error; err != nil {
			return fmt.Errorf("create full text search settings failed: %w", err)
		}
		return nil
	}

	if err := r.db.Save(setting).Error; err != nil {
		return fmt.Errorf("update full text search settings failed: %w", err)
	}

	return nil
}
