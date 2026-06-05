package model

import "time"

type ArchiveExtractTask struct {
	BaseModel
	TaskID         string     `gorm:"uniqueIndex;size:64;not null" json:"task_id"`
	QueueJobID     *uint      `gorm:"index" json:"queue_job_id"`
	NodeID         uint       `gorm:"index;not null" json:"node_id"`
	NodeName       string     `gorm:"size:255;not null" json:"node_name"`
	NodeCapability string     `gorm:"size:64;not null;default:'extract_archive'" json:"node_capability"`
	SourceFileID   uint       `gorm:"index;not null" json:"source_file_id"`
	TargetFolderID *uint      `gorm:"index" json:"target_folder_id"`
	Status         string     `gorm:"size:32;not null" json:"status"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message"`
	ExtractedFiles string     `gorm:"type:text" json:"extracted_files"`
	CompletedAt    *time.Time `json:"completed_at"`
}

func (ArchiveExtractTask) TableName() string {
	return "archive_extract_tasks"
}
