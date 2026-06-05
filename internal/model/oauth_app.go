package model

// OAuthApp stores administrator-managed OAuth client applications.
type OAuthApp struct {
	BaseModel
	Slug                   string `gorm:"uniqueIndex;size:80;not null" json:"slug"`
	Name                   string `gorm:"size:255;not null" json:"name"`
	Description            string `gorm:"type:text" json:"description"`
	AppName                string `gorm:"size:255;not null" json:"app_name"`
	IconPath               string `gorm:"type:text" json:"icon_path"`
	ClientID               string `gorm:"uniqueIndex;size:128;not null" json:"client_id"`
	ClientSecret           string `gorm:"size:255" json:"-"`
	RedirectURIsJSON       string `gorm:"type:text" json:"redirect_uris_json"`
	ScopesJSON             string `gorm:"type:text" json:"scopes_json"`
	PermissionsJSON        string `gorm:"type:text" json:"permissions_json"`
	IsSystem               bool   `gorm:"not null;default:false" json:"is_system"`
	Enabled                bool   `gorm:"not null;default:true" json:"enabled"`
	TokenTTL               string `gorm:"size:64;not null;default:'7 天'" json:"token_ttl"`
	RefreshTokenTTLSeconds int    `gorm:"not null;default:604800" json:"refresh_token_ttl_seconds"`
}

func (OAuthApp) TableName() string {
	return "oauth_apps"
}
