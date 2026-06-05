package model

import "time"

const (
	OfflineTaskStatusQueued      = "queued"
	OfflineTaskStatusDownloading = "downloading"
	OfflineTaskStatusPaused      = "paused"
	OfflineTaskStatusCompleted   = "completed"
	OfflineTaskStatusFailed      = "failed"
	OfflineTaskStatusExpired     = "expired"
)

// OfflineDownloadTask stores one user-owned offline download task.
type OfflineDownloadTask struct {
	BaseModel
	UserID           uint       `gorm:"not null;index:idx_offline_tasks_user_status,priority:1" json:"user_id"`
	TaskToken        string     `gorm:"size:64;not null" json:"task_token"`
	Name             string     `gorm:"size:255;not null" json:"name"`
	SourceURL        string     `gorm:"type:text;not null" json:"source_url"`
	SavePath         string     `gorm:"size:1000;not null;default:'/offline-downloads'" json:"save_path"`
	Status           string     `gorm:"size:20;not null;default:'queued';index:idx_offline_tasks_user_status,priority:2" json:"status"`
	Progress         int        `gorm:"not null;default:0" json:"progress"`
	SpeedText        string     `gorm:"size:80;not null;default:'waiting'" json:"speed_text"`
	SizeText         string     `gorm:"size:80;not null;default:'unknown'" json:"size_text"`
	DownloadedBytes  int64      `gorm:"not null;default:0" json:"downloaded_bytes"`
	TotalBytes       int64      `gorm:"not null;default:0" json:"total_bytes"`
	ErrorMessage     string     `gorm:"type:text" json:"error_message"`
	QueueJobID       *uint      `gorm:"index" json:"queue_job_id"`
	NodeID           *uint      `gorm:"index" json:"node_id"`
	Downloader       string     `gorm:"size:64" json:"downloader"`
	RPCURL           string     `gorm:"type:text" json:"rpc_url"`
	RPCSecret        string     `gorm:"type:text" json:"rpc_secret"`
	TaskOptions      string     `gorm:"type:text" json:"task_options"`
	TempDir          string     `gorm:"type:text" json:"temp_dir"`
	RefreshInterval  int        `gorm:"not null;default:5" json:"refresh_interval"`
	WaitForSeeding   bool       `gorm:"not null;default:false" json:"wait_for_seeding"`
	RemoteTaskID     string     `gorm:"size:191" json:"remote_task_id"`
	DispatchNodeID   *uint      `gorm:"index" json:"dispatch_node_id"`
	DispatchNodeName string     `gorm:"size:255" json:"dispatch_node_name"`
	DispatchNodeType string     `gorm:"size:32" json:"dispatch_node_type"`
	SavedFileID      *uint      `gorm:"index" json:"saved_file_id"`
	SavedFolderID    *uint      `gorm:"index" json:"saved_folder_id"`
	CompletedAt      *time.Time `json:"completed_at"`
}

func (OfflineDownloadTask) TableName() string {
	return "offline_download_tasks"
}
