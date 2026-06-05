package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// CaptchaSettingRepository handles persistence for singleton captcha settings.
type CaptchaSettingRepository interface {
	Get() (*model.CaptchaSetting, error)
	Save(setting *model.CaptchaSetting) error
}

type captchaSettingRepository struct {
	db *gorm.DB
}

// NewCaptchaSettingRepository creates a captcha settings repository.
func NewCaptchaSettingRepository(db *gorm.DB) CaptchaSettingRepository {
	return &captchaSettingRepository{db: db}
}

// Get returns the current singleton captcha settings row if it exists.
func (r *captchaSettingRepository) Get() (*model.CaptchaSetting, error) {
	var setting model.CaptchaSetting
	if err := r.db.Order("id asc").First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询验证码设置失败: %w", err)
	}

	return &setting, nil
}

// Save creates or updates the singleton captcha settings row.
func (r *captchaSettingRepository) Save(setting *model.CaptchaSetting) error {
	if setting == nil {
		return fmt.Errorf("验证码设置不能为空")
	}

	if setting.ID == 0 {
		if err := r.db.Create(setting).Error; err != nil {
			return fmt.Errorf("创建验证码设置失败: %w", err)
		}
		return nil
	}

	if err := r.db.Save(setting).Error; err != nil {
		return fmt.Errorf("更新验证码设置失败: %w", err)
	}

	return nil
}
