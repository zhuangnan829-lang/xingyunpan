package model

// EmailSetting stores admin-managed email delivery settings.
type EmailSetting struct {
	BaseModel
	Enabled             bool   `gorm:"not null;default:true" json:"enabled"`
	Provider            string `gorm:"size:32;not null;default:'qq'" json:"provider"`
	Host                string `gorm:"size:255" json:"host"`
	Port                int    `gorm:"not null;default:587" json:"port"`
	Username            string `gorm:"size:255;not null" json:"username"`
	Password            string `gorm:"size:255;not null" json:"password"`
	FromName            string `gorm:"size:120;not null;default:'星云盘'" json:"from_name"`
	FromAddress         string `gorm:"size:255;not null" json:"from_address"`
	ReplyTo             string `gorm:"size:255" json:"reply_to"`
	ForceSSL            bool   `gorm:"not null;default:false" json:"force_ssl"`
	ConnectionTimeout   int    `gorm:"not null;default:30" json:"connection_timeout"`
	CodeTTLSeconds      int    `gorm:"not null;default:300" json:"code_ttl_seconds"`
	SendIntervalSeconds int    `gorm:"not null;default:60" json:"send_interval_seconds"`
}

// TableName specifies the DB table name.
func (EmailSetting) TableName() string {
	return "email_settings"
}
