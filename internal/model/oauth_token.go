package model

import "time"

type OAuthAuthorizationCode struct {
	BaseModel
	CodeHash    string     `gorm:"uniqueIndex;size:128;not null" json:"-"`
	AppID       uint       `gorm:"index;not null" json:"app_id"`
	UserID      uint       `gorm:"index;not null" json:"user_id"`
	RedirectURI string     `gorm:"type:text;not null" json:"redirect_uri"`
	ScopesJSON  string     `gorm:"type:text" json:"scopes_json"`
	State       string     `gorm:"type:text" json:"state"`
	ExpiresAt   time.Time  `gorm:"index;not null" json:"expires_at"`
	UsedAt      *time.Time `gorm:"index" json:"used_at,omitempty"`
	App         OAuthApp   `gorm:"foreignKey:AppID" json:"app,omitempty"`
	User        User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (OAuthAuthorizationCode) TableName() string {
	return "oauth_authorization_codes"
}

type OAuthAccessToken struct {
	BaseModel
	TokenHash  string     `gorm:"uniqueIndex;size:128;not null" json:"-"`
	AppID      uint       `gorm:"index;not null" json:"app_id"`
	UserID     uint       `gorm:"index;not null" json:"user_id"`
	ScopesJSON string     `gorm:"type:text" json:"scopes_json"`
	ExpiresAt  time.Time  `gorm:"index;not null" json:"expires_at"`
	RevokedAt  *time.Time `gorm:"index" json:"revoked_at,omitempty"`
	App        OAuthApp   `gorm:"foreignKey:AppID" json:"app,omitempty"`
	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (OAuthAccessToken) TableName() string {
	return "oauth_access_tokens"
}

type OAuthRefreshToken struct {
	BaseModel
	TokenHash  string     `gorm:"uniqueIndex;size:128;not null" json:"-"`
	AppID      uint       `gorm:"index;not null" json:"app_id"`
	UserID     uint       `gorm:"index;not null" json:"user_id"`
	ScopesJSON string     `gorm:"type:text" json:"scopes_json"`
	ExpiresAt  time.Time  `gorm:"index;not null" json:"expires_at"`
	RevokedAt  *time.Time `gorm:"index" json:"revoked_at,omitempty"`
	UsedAt     *time.Time `gorm:"index" json:"used_at,omitempty"`
	App        OAuthApp   `gorm:"foreignKey:AppID" json:"app,omitempty"`
	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (OAuthRefreshToken) TableName() string {
	return "oauth_refresh_tokens"
}

type OAuthGrant struct {
	BaseModel
	AppID      uint       `gorm:"uniqueIndex:idx_oauth_grants_app_user;not null" json:"app_id"`
	UserID     uint       `gorm:"uniqueIndex:idx_oauth_grants_app_user;not null" json:"user_id"`
	ScopesJSON string     `gorm:"type:text" json:"scopes_json"`
	RevokedAt  *time.Time `gorm:"index" json:"revoked_at,omitempty"`
	App        OAuthApp   `gorm:"foreignKey:AppID" json:"app,omitempty"`
	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (OAuthGrant) TableName() string {
	return "oauth_grants"
}

type OAuthAuditLog struct {
	BaseModel
	ActorUserID  uint   `gorm:"index;not null;default:0" json:"actor_user_id"`
	AppID        uint   `gorm:"index;not null;default:0" json:"app_id"`
	UserID       uint   `gorm:"index;not null;default:0" json:"user_id"`
	Action       string `gorm:"size:80;index;not null" json:"action"`
	MetadataJSON string `gorm:"type:text" json:"metadata_json"`
}

func (OAuthAuditLog) TableName() string {
	return "oauth_audit_logs"
}
