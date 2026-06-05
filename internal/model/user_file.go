// 路径: internal/model/user_file.go
package model

// UserFile 用户文件模型（用户视角的文件/文件夹）
type UserFile struct {
	BaseModel
	UserID         uint   `gorm:"index:idx_user_parent,priority:1;not null;comment:用户ID" json:"user_id"` // ✅ 任务 8: 复合索引
	ParentID       *uint  `gorm:"index:idx_user_parent,priority:2;comment:父目录ID(NULL表示根目录)" json:"parent_id"` // ✅ 任务 8: 复合索引
	FileName       string `gorm:"size:255;not null;comment:文件/文件夹名称" json:"file_name"`
	PhysicalFileID *uint  `gorm:"index;comment:物理文件ID(文件夹为NULL)" json:"physical_file_id"`
	IsFolder       bool   `gorm:"index:idx_user_folder;default:false;not null;comment:是否为文件夹" json:"is_folder"` // ✅ 任务 8: 用于排序的索引
	FileSize       int64  `gorm:"default:0;comment:文件大小(字节,文件夹为0)" json:"file_size"`
	FilePath       string `gorm:"size:1000;comment:文件完整路径" json:"file_path"`

	// 关联
	User         User          `gorm:"foreignKey:UserID" json:"-"`
	PhysicalFile *PhysicalFile `gorm:"foreignKey:PhysicalFileID" json:"physical_file,omitempty"`
	Parent       *UserFile     `gorm:"foreignKey:ParentID" json:"-"`
	Children     []UserFile    `gorm:"foreignKey:ParentID" json:"-"`
}

// TableName 指定表名
func (UserFile) TableName() string {
	return "user_files"
}
