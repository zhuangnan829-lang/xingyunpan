package model

import "time"

// Share stores a public share link and its access limits.
type Share struct {
	BaseModel
	UserID         uint       `gorm:"not null;index:idx_user_id;comment:creator user id" json:"user_id"`
	ShareToken     string     `gorm:"uniqueIndex:idx_share_token;size:64;not null;comment:share token" json:"share_token"`
	AccessCodeHash string     `gorm:"size:255;comment:access code hash" json:"-"`
	ExpiresAt      *time.Time `gorm:"index:idx_expires_at;comment:expiration time" json:"expires_at"`
	MaxDownloads   *int       `gorm:"comment:max download count before expiration" json:"max_downloads"`
	DownloadCount  int        `gorm:"default:0;comment:download count" json:"download_count"`
	AccessCount    int        `gorm:"default:0;comment:access count" json:"access_count"`

	User       User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ShareFiles []ShareFile `gorm:"foreignKey:ShareID;constraint:OnDelete:CASCADE" json:"share_files,omitempty"`
}

func (Share) TableName() string {
	return "shares"
}
