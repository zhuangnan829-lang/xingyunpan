package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// EmailSettingRepository handles persistence for singleton email settings.
type EmailSettingRepository interface {
	Get() (*model.EmailSetting, error)
	Save(setting *model.EmailSetting) error
}

type emailSettingRepository struct {
	db *gorm.DB
}

// NewEmailSettingRepository creates an email settings repository.
func NewEmailSettingRepository(db *gorm.DB) EmailSettingRepository {
	return &emailSettingRepository{db: db}
}

func (r *emailSettingRepository) Get() (*model.EmailSetting, error) {
	var setting model.EmailSetting
	if err := r.db.Order("id asc").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询邮件设置失败: %w", err)
	}

	return &setting, nil
}

func (r *emailSettingRepository) Save(setting *model.EmailSetting) error {
	if setting == nil {
		return fmt.Errorf("邮件设置不能为空")
	}

	if setting.ID == 0 {
		if err := r.db.Create(setting).Error; err != nil {
			return fmt.Errorf("创建邮件设置失败: %w", err)
		}
		return nil
	}

	if err := r.db.Save(setting).Error; err != nil {
		return fmt.Errorf("更新邮件设置失败: %w", err)
	}

	return nil
}
