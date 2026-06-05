package model

import "time"

type StoragePolicyHitLog struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	StoragePolicyID   uint      `gorm:"index:idx_policy_hit_time,priority:1;not null;default:0" json:"storage_policy_id"`
	StoragePolicyName string    `gorm:"size:255;not null;default:''" json:"storage_policy_name"`
	HitType           string    `gorm:"size:32;index;not null;default:'global_default'" json:"hit_type"`
	Action            string    `gorm:"size:32;index;not null" json:"action"`
	UserID            uint      `gorm:"index;not null;default:0" json:"user_id"`
	Username          string    `gorm:"size:120;not null;default:''" json:"username"`
	UserGroupID       uint      `gorm:"index;not null;default:0" json:"user_group_id"`
	UserGroupName     string    `gorm:"size:120;not null;default:''" json:"user_group_name"`
	FileID            uint      `gorm:"index;not null;default:0" json:"file_id"`
	FileName          string    `gorm:"size:255;not null;default:''" json:"file_name"`
	FileSize          int64     `gorm:"not null;default:0" json:"file_size"`
	ResourceType      string    `gorm:"size:64;not null;default:''" json:"resource_type"`
	ResourceID        string    `gorm:"size:128;not null;default:''" json:"resource_id"`
	ConfigJSON        string    `gorm:"type:longtext" json:"config_json"`
	CreatedAt         time.Time `gorm:"index:idx_policy_hit_time,priority:2" json:"created_at"`
}

func (StoragePolicyHitLog) TableName() string {
	return "storage_policy_hit_logs"
}
