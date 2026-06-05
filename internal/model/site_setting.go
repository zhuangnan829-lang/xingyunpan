package model

// SiteSetting stores admin-managed site information settings.
type SiteSetting struct {
	BaseModel
	SiteName            string `gorm:"size:120;not null;default:'星云盘'" json:"site_name"`
	Tagline             string `gorm:"size:255;not null;default:'新一代私有云盘控制台'" json:"tagline"`
	Description         string `gorm:"type:text" json:"description"`
	TermsURL            string `gorm:"size:500" json:"terms_url"`
	PrivacyURL          string `gorm:"size:500" json:"privacy_url"`
	PrimaryURL          string `gorm:"size:500" json:"primary_url"`
	BackupURLsJSON      string `gorm:"type:text" json:"-"`
	LogoLight           string `gorm:"size:500" json:"logo_light"`
	LogoDark            string `gorm:"size:500" json:"logo_dark"`
	Favicon             string `gorm:"size:500" json:"favicon"`
	Logo192             string `gorm:"size:500" json:"logo_192"`
	InjectionCode       string `gorm:"type:text" json:"injection_code"`
	AppearanceSettingsJSON string `gorm:"type:longtext" json:"-"`
	MobileGuideEnabled  bool   `gorm:"not null;default:true" json:"mobile_guide_enabled"`
	MobileFeedbackURL   string `gorm:"size:500" json:"mobile_feedback_url"`
	DesktopGuideEnabled bool   `gorm:"not null;default:true" json:"desktop_guide_enabled"`
	DesktopCommunityURL string `gorm:"size:500" json:"desktop_community_url"`
	AllowRegistration   bool   `gorm:"not null;default:true" json:"allow_registration"`
	EmailActivation     bool   `gorm:"not null;default:false" json:"email_activation"`
	PasskeyLoginEnabled bool   `gorm:"not null;default:false" json:"passkey_login_enabled"`
	DefaultGroup        string `gorm:"size:64;not null;default:'User'" json:"default_group"`
	AvatarPath          string `gorm:"size:255;not null;default:'avatar'" json:"avatar_path"`
	AvatarSizeLimitMB   int    `gorm:"not null;default:4" json:"avatar_size_limit_mb"`
	AvatarDimension     int    `gorm:"not null;default:200" json:"avatar_dimension"`
	GravatarServer      string `gorm:"size:500;not null;default:'https://www.gravatar.com/'" json:"gravatar_server"`
}

// TableName specifies the DB table name.
func (SiteSetting) TableName() string {
	return "site_settings"
}
