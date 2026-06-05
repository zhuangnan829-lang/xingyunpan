package model

import "time"

// Node stores admin-managed compute node definitions.
type Node struct {
	BaseModel
	Name                   string     `gorm:"size:255;not null" json:"name"`
	Type                   string     `gorm:"size:32;not null;default:'worker'" json:"type"`
	Enabled                bool       `gorm:"not null;default:true" json:"enabled"`
	Weight                 int        `gorm:"not null;default:1" json:"weight"`
	IsBuiltIn              bool       `gorm:"not null;default:false" json:"is_built_in"`
	CreateArchive          bool       `gorm:"not null;default:false" json:"create_archive"`
	ExtractArchive         bool       `gorm:"not null;default:false" json:"extract_archive"`
	OfflineDownload        bool       `gorm:"not null;default:false" json:"offline_download"`
	OfflineDownloader      string     `gorm:"size:64;not null;default:'Aria2'" json:"offline_downloader"`
	OfflineRPCURL          string     `gorm:"type:text" json:"offline_rpc_url"`
	OfflineRPCSecret       string     `gorm:"type:text" json:"offline_rpc_secret"`
	OfflineTaskOptions     string     `gorm:"type:text" json:"offline_task_options"`
	OfflineTempDir         string     `gorm:"type:text" json:"offline_temp_dir"`
	OfflineRefreshInterval int        `gorm:"not null;default:5" json:"offline_refresh_interval"`
	OfflineWaitForSeeding  bool       `gorm:"not null;default:false" json:"offline_wait_for_seeding"`
	HealthStatus           string     `gorm:"size:32;not null;default:'unknown'" json:"health_status"`
	HealthMessage          string     `gorm:"type:text" json:"health_message"`
	LastHeartbeatAt        *time.Time `json:"last_heartbeat_at"`
	LastCheckedAt          *time.Time `json:"last_checked_at"`
}

// TableName specifies the DB table name.
func (Node) TableName() string {
	return "nodes"
}
