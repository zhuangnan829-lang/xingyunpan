package model

// QueueSetting stores admin-managed configuration for a named queue.
type QueueSetting struct {
	BaseModel
	QueueKey      string `gorm:"uniqueIndex:idx_queue_settings_queue_key;size:32;not null" json:"queue_key"`
	WorkerNum     int    `gorm:"not null;default:5" json:"worker_num"`
	MaxExecution  int    `gorm:"not null;default:3600" json:"max_execution"`
	BackoffFactor int    `gorm:"not null;default:2" json:"backoff_factor"`
	MaxBackoff    int    `gorm:"not null;default:60" json:"max_backoff"`
	MaxRetry      int    `gorm:"not null;default:0" json:"max_retry"`
	RetryDelay    int    `gorm:"not null;default:0" json:"retry_delay"`
}

// TableName specifies the DB table name.
func (QueueSetting) TableName() string {
	return "queue_settings"
}
