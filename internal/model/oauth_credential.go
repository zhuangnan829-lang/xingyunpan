package model

import "time"

type OAuthCredential struct {
	BaseModel
	AppID                 uint       `gorm:"index;not null" json:"app_id"`
	UserID                uint       `gorm:"index;not null;default:0" json:"user_id"`
	Provider              string     `gorm:"size:64;not null;default:'xingyunpan'" json:"provider"`
	Subject               string     `gorm:"size:255" json:"subject"`
	AccessToken           string     `gorm:"type:text" json:"-"`
	RefreshToken          string     `gorm:"type:text" json:"-"`
	TokenType             string     `gorm:"size:32;not null;default:'Bearer'" json:"token_type"`
	ScopesJSON            string     `gorm:"type:text" json:"scopes_json"`
	TokenEndpoint         string     `gorm:"type:text" json:"token_endpoint"`
	AccessTokenExpiresAt  *time.Time `gorm:"index" json:"access_token_expires_at,omitempty"`
	RefreshTokenExpiresAt *time.Time `gorm:"index" json:"refresh_token_expires_at,omitempty"`
	LastRefreshedAt       *time.Time `json:"last_refreshed_at,omitempty"`
	NextRefreshAt         *time.Time `gorm:"index" json:"next_refresh_at,omitempty"`
	Status                string     `gorm:"size:32;not null;default:'active';index" json:"status"`
	RefreshError          string     `gorm:"type:text" json:"refresh_error,omitempty"`
	MetadataJSON          string     `gorm:"type:text" json:"metadata_json"`

	App OAuthApp `gorm:"foreignKey:AppID" json:"app,omitempty"`
}

func (OAuthCredential) TableName() string {
	return "oauth_credentials"
}
