// 路径: internal/model/file_deletion.go
package model

import "time"

// FileDeletion 文件删除队列模型
type FileDeletion struct {
	ID             uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	PhysicalFileID uint       `gorm:"not null;comment:物理文件ID" json:"physical_file_id"`
	StoragePath    string     `gorm:"size:500;not null;comment:存储路径" json:"storage_path"`
	StorageType    string     `gorm:"size:20;not null;comment:存储类型(local/minio/oss)" json:"storage_type"`
	Status         string     `gorm:"index;size:20;default:'pending';not null;comment:状态(pending/processing/completed/failed)" json:"status"`
	RetryCount     int        `gorm:"default:0;comment:重试次数" json:"retry_count"`
	CreatedAt      time.Time  `gorm:"not null;comment:创建时间" json:"created_at"`
	DeletedAt      *time.Time `gorm:"comment:删除时间" json:"deleted_at"`
}

// TableName 指定表名
func (FileDeletion) TableName() string {
	return "file_deletions"
}

// 状态常量
const (
	DeletionStatusPending    = "pending"
	DeletionStatusProcessing = "processing"
	DeletionStatusCompleted  = "completed"
	DeletionStatusFailed     = "failed"
)
