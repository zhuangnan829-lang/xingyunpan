package model

import "time"

const (
	DavAccountPermissionRead  = "read"
	DavAccountPermissionWrite = "write"
	DavAccountStatusActive    = "active"
	DavAccountStatusDisabled  = "disabled"
)

// DavAccount stores one user-owned external mount/WebDAV account.
type DavAccount struct {
	BaseModel
	UserID       uint       `gorm:"not null;index:idx_dav_accounts_user_status,priority:1" json:"user_id"`
	AccountToken string     `gorm:"uniqueIndex;size:64;not null" json:"account_token"`
	Name         string     `gorm:"size:120;not null" json:"name"`
	RootPath     string     `gorm:"size:1000;not null;default:'/'" json:"root_path"`
	Permission   string     `gorm:"size:16;not null;default:'write'" json:"permission"`
	ReverseProxy bool       `gorm:"not null;default:false" json:"reverse_proxy"`
	Status       string     `gorm:"size:20;not null;default:'active';index:idx_dav_accounts_user_status,priority:2" json:"status"`
	SecretHash   string     `gorm:"size:255;not null" json:"-"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	LastUsedIP   string     `gorm:"size:64" json:"last_used_ip"`
	Description  string     `gorm:"type:text" json:"description"`
}

func (DavAccount) TableName() string {
	return "dav_accounts"
}
