// 路径: internal/model/recycle_bin.go
package model

import (
	"time"
)

// RecycleBin 回收站模型
// 注意：此模型不使用 BaseModel，因为它有自己的业务 deleted_at 字段
type RecycleBin struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	UserID        uint      `gorm:"not null;index:idx_user_expires,priority:1;comment:用户ID" json:"user_id"`
	FileID        uint      `gorm:"not null;comment:文件ID" json:"file_id"`
	FileName      string    `gorm:"size:255;not null;comment:文件名" json:"file_name"`
	FileSize      int64     `gorm:"not null;comment:文件大小" json:"file_size"`
	FileType      string    `gorm:"size:100;comment:文件类型" json:"file_type"`
	OriginalPath  string    `gorm:"size:1000;not null;comment:原始路径" json:"original_path"`
	FileDeletedAt time.Time `gorm:"column:deleted_at;not null;comment:删除时间" json:"deleted_at"` // 业务字段：文件被删除的时间
	ExpiresAt     time.Time `gorm:"not null;index:idx_user_expires,priority:2;comment:过期时间" json:"expires_at"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (RecycleBin) TableName() string {
	return "recycle_bin"
}
