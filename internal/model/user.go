package model

import "time"

// User represents an application user account.
type User struct {
	BaseModel
	Username    string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password    string     `gorm:"size:255;not null" json:"-"`
	Email       string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Role        string     `gorm:"size:20;default:'user';not null" json:"role"`
	Enabled     bool       `gorm:"not null;default:true" json:"enabled"`
	UserGroupID uint       `gorm:"not null;default:0;index" json:"user_group_id"`
	AvatarURL   string     `gorm:"size:1000" json:"avatar_url"`
	Capacity    int64      `gorm:"default:10737418240;not null" json:"capacity"`
	UsedSize    int64      `gorm:"default:0;not null" json:"used_size"`
	LastSeenAt  *time.Time `gorm:"index" json:"last_seen_at"`
}

func (User) TableName() string {
	return "users"
}
