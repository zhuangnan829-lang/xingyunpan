package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// SiteSettingRepository handles persistence for singleton site settings.
type SiteSettingRepository interface {
	Get() (*model.SiteSetting, error)
	Save(setting *model.SiteSetting) error
}

type siteSettingRepository struct {
	db *gorm.DB
}

// NewSiteSettingRepository creates a site settings repository.
func NewSiteSettingRepository(db *gorm.DB) SiteSettingRepository {
	return &siteSettingRepository{db: db}
}

// Get returns the current singleton site settings row if it exists.
func (r *siteSettingRepository) Get() (*model.SiteSetting, error) {
	var setting model.SiteSetting
	if err := r.db.Order("id asc").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询站点设置失败: %w", err)
	}

	return &setting, nil
}

// Save creates or updates the singleton site settings row.
func (r *siteSettingRepository) Save(setting *model.SiteSetting) error {
	if setting == nil {
		return fmt.Errorf("站点设置不能为空")
	}

	if setting.ID == 0 {
		if err := r.db.Create(setting).Error; err != nil {
			return fmt.Errorf("创建站点设置失败: %w", err)
		}
		return nil
	}

	if err := r.db.Save(setting).Error; err != nil {
		return fmt.Errorf("更新站点设置失败: %w", err)
	}

	return nil
}
