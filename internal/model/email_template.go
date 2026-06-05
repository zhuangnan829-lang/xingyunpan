package model

// EmailTemplate stores editable admin email templates.
type EmailTemplate struct {
	BaseModel
	TemplateKey   string `gorm:"size:64;uniqueIndex;not null" json:"template_key"`
	Name          string `gorm:"size:120;not null" json:"name"`
	Description   string `gorm:"size:255" json:"description"`
	Status        string `gorm:"size:32;not null;default:'enabled'" json:"status"`
	StatusTone    string `gorm:"size:32;not null;default:'success'" json:"status_tone"`
	Pro           bool   `gorm:"not null;default:false" json:"pro"`
	LanguagesJSON string `gorm:"type:longtext;not null" json:"-"`
}

// TableName specifies the DB table name.
func (EmailTemplate) TableName() string {
	return "email_templates"
}
