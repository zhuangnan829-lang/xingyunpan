package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// EmailTemplateRepository handles persistence for email templates.
type EmailTemplateRepository interface {
	List() ([]model.EmailTemplate, error)
	GetByKey(templateKey string) (*model.EmailTemplate, error)
	Save(template *model.EmailTemplate) error
}

type emailTemplateRepository struct {
	db *gorm.DB
}

// NewEmailTemplateRepository creates an email template repository.
func NewEmailTemplateRepository(db *gorm.DB) EmailTemplateRepository {
	return &emailTemplateRepository{db: db}
}

func (r *emailTemplateRepository) List() ([]model.EmailTemplate, error) {
	var templates []model.EmailTemplate
	if err := r.db.Order("id asc").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("查询邮件模板失败: %w", err)
	}
	return templates, nil
}

func (r *emailTemplateRepository) GetByKey(templateKey string) (*model.EmailTemplate, error) {
	var template model.EmailTemplate
	if err := r.db.Where("template_key = ?", templateKey).First(&template).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询邮件模板失败: %w", err)
	}
	return &template, nil
}

func (r *emailTemplateRepository) Save(template *model.EmailTemplate) error {
	if template == nil {
		return fmt.Errorf("邮件模板不能为空")
	}

	if template.ID == 0 {
		if err := r.db.Create(template).Error; err != nil {
			return fmt.Errorf("创建邮件模板失败: %w", err)
		}
		return nil
	}

	if err := r.db.Save(template).Error; err != nil {
		return fmt.Errorf("更新邮件模板失败: %w", err)
	}

	return nil
}
