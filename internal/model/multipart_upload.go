// 路径: internal/model/multipart_upload.go
package model

import "time"

// MultipartUpload 分片上传任务模型
type MultipartUpload struct {
	BaseModel
	UploadID    string     `gorm:"uniqueIndex;size:64;not null;comment:上传任务ID(UUID)" json:"upload_id"`
	UserID      uint       `gorm:"index:idx_user_status,priority:1;not null;comment:用户ID" json:"user_id"`
	FileName    string     `gorm:"size:255;not null;comment:文件名称" json:"file_name"`
	FileHash    string     `gorm:"size:64;not null;comment:文件哈希(SHA256)" json:"file_hash"`
	FileSize    int64      `gorm:"not null;comment:文件大小(字节)" json:"file_size"`
	TotalChunks int        `gorm:"not null;comment:总分片数" json:"total_chunks"`
	ChunkSize   int        `gorm:"not null;comment:分片大小(字节)" json:"chunk_size"`
	StorageType string     `gorm:"size:20;not null;comment:存储类型(local/minio/oss)" json:"storage_type"`
	StoragePath string     `gorm:"size:500;comment:存储路径" json:"storage_path"`
	Status      string     `gorm:"index:idx_user_status,priority:2;size:20;default:'uploading';not null;comment:状态(uploading/completed/cancelled)" json:"status"`
	CompletedAt *time.Time `gorm:"comment:完成时间" json:"completed_at"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName 指定表名
func (MultipartUpload) TableName() string {
	return "multipart_uploads"
}

// 状态常量
const (
	MultipartStatusUploading = "uploading"
	MultipartStatusCompleted = "completed"
	MultipartStatusCancelled = "cancelled"
)
