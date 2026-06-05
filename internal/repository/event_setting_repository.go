package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// EventSettingRepository handles persistence for singleton event settings.
type EventSettingRepository interface {
	Get() (*model.EventSetting, error)
	Save(setting *model.EventSetting) error
}

type eventSettingRepository struct {
	db *gorm.DB
}

// NewEventSettingRepository creates an event settings repository.
func NewEventSettingRepository(db *gorm.DB) EventSettingRepository {
	return &eventSettingRepository{db: db}
}

// Get returns the current singleton event settings row if it exists.
func (r *eventSettingRepository) Get() (*model.EventSetting, error) {
	var setting model.EventSetting
	if err := r.db.Order("id asc").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询事件设置失败: %w", err)
	}

	return &setting, nil
}

// Save creates or updates the singleton event settings row.
func (r *eventSettingRepository) Save(setting *model.EventSetting) error {
	if setting == nil {
		return fmt.Errorf("事件设置不能为空")
	}

	if setting.ID == 0 {
		if err := r.db.Create(setting).Error; err != nil {
			return fmt.Errorf("创建事件设置失败: %w", err)
		}
		return nil
	}

	if err := r.db.Save(setting).Error; err != nil {
		return fmt.Errorf("更新事件设置失败: %w", err)
	}

	return nil
}
