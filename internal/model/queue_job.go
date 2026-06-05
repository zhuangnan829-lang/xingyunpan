package model

import "time"

const (
	QueueJobStatusPending    = "pending"
	QueueJobStatusProcessing = "processing"
	QueueJobStatusCompleted  = "completed"
	QueueJobStatusFailed     = "failed"
)

// QueueJob stores one unified async queue task.
type QueueJob struct {
	BaseModel
	QueueKey         string     `gorm:"index:idx_queue_jobs_lookup,priority:1;size:32;not null" json:"queue_key"`
	JobType          string     `gorm:"size:64;not null" json:"job_type"`
	ResourceType     string     `gorm:"size:64;not null" json:"resource_type"`
	ResourceID       string     `gorm:"size:128;not null" json:"resource_id"`
	DedupeKey        string     `gorm:"uniqueIndex;size:191;not null" json:"dedupe_key"`
	Payload          string     `gorm:"type:text;not null" json:"payload"`
	DispatchNodeID   *uint      `gorm:"index" json:"dispatch_node_id"`
	DispatchNodeName string     `gorm:"size:255" json:"dispatch_node_name"`
	DispatchNodeType string     `gorm:"size:32" json:"dispatch_node_type"`
	NodeCapability   string     `gorm:"size:64;index" json:"node_capability"`
	Status           string     `gorm:"index:idx_queue_jobs_lookup,priority:2;size:20;default:'pending';not null" json:"status"`
	Attempts         int        `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts      int        `gorm:"not null;default:0" json:"max_attempts"`
	ScheduledAt      time.Time  `gorm:"index:idx_queue_jobs_lookup,priority:3;not null" json:"scheduled_at"`
	StartedAt        *time.Time `json:"started_at"`
	FinishedAt       *time.Time `json:"finished_at"`
	LastError        string     `gorm:"type:text" json:"last_error"`
	Result           string     `gorm:"type:text" json:"result"`
}

// TableName specifies the DB table name.
func (QueueJob) TableName() string {
	return "queue_jobs"
}
