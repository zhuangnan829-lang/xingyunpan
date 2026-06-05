package model

// UserPreference stores per-user UI and behavior preferences.
type UserPreference struct {
	BaseModel
	UserID            uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	Language          string `gorm:"size:32;not null;default:'zh-CN'" json:"language"`
	Timezone          string `gorm:"size:64;not null;default:'Asia/Shanghai'" json:"timezone"`
	Mode              string `gorm:"size:16;not null;default:'light'" json:"mode"`
	Theme             string `gorm:"size:32;not null;default:'sky'" json:"theme"`
	KeepVersions      bool   `gorm:"not null;default:true" json:"keep_versions"`
	VersionExtensions string `gorm:"size:512;not null;default:''" json:"version_extensions"`
	MaxVersions       int    `gorm:"not null;default:10" json:"max_versions"`
	ViewSync          string `gorm:"size:16;not null;default:'server'" json:"view_sync"`
	ExpandTree        bool   `gorm:"not null;default:true" json:"expand_tree"`
	FolderAction      string `gorm:"size:16;not null;default:'open'" json:"folder_action"`
	HomeVisibility    string `gorm:"size:32;not null;default:'passwordless'" json:"home_visibility"`
}

func (UserPreference) TableName() string {
	return "user_preferences"
}
