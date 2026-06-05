package model

// CaptchaSetting stores admin-managed captcha configuration.
type CaptchaSetting struct {
	BaseModel
	LoginEnabled         bool   `gorm:"not null;default:true" json:"login_enabled"`
	RegisterEnabled      bool   `gorm:"not null;default:true" json:"register_enabled"`
	ResetPasswordEnabled bool   `gorm:"not null;default:true" json:"reset_password_enabled"`
	Provider             string `gorm:"size:32;not null;default:'image'" json:"provider"`
	SecurityLevel        string `gorm:"size:32;not null;default:'balanced'" json:"security_level"`
	SiteKey              string `gorm:"size:255" json:"site_key"`
	SecretKey            string `gorm:"size:255" json:"secret_key"`
	FailureThreshold     int    `gorm:"not null;default:3" json:"failure_threshold"`
	CooldownSeconds      int    `gorm:"not null;default:60" json:"cooldown_seconds"`
	WhitelistPathsJSON   string `gorm:"type:text" json:"-"`
}

// TableName specifies the DB table name.
func (CaptchaSetting) TableName() string {
	return "captcha_settings"
}
