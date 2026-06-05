package model

// MediaSetting stores admin-managed media processing preferences.
type MediaSetting struct {
	BaseModel
	ImageMode          string  `gorm:"size:32;not null;default:'quality'" json:"image_mode"`
	ImageMaxSizeGB     int     `gorm:"not null;default:8" json:"image_max_size_gb"`
	ImageQuality       int     `gorm:"not null;default:88" json:"image_quality"`
	VideoPreviewSecond float64 `gorm:"not null;default:1" json:"video_preview_second"`
	VideoStrategy      string  `gorm:"size:32;not null;default:'balanced'" json:"video_strategy"`
	MetadataDeepScan   bool    `gorm:"not null;default:true" json:"metadata_deep_scan"`
	LibreOfficePath    string  `gorm:"size:255;not null;default:'soffice'" json:"libreoffice_path"`
}

// TableName specifies the DB table name.
func (MediaSetting) TableName() string {
	return "media_settings"
}
