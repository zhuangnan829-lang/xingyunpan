// 路径: internal/model/base.go
package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含所有表的公共字段
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
