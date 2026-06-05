// 路径: internal/model/file_version.go
package model

// FileVersion 文件版本模型
type FileVersion struct {
	BaseModel
	FileID         uint  `gorm:"not null;index:idx_file_versions,priority:1;comment:文件ID" json:"file_id"`
	VersionNumber  int   `gorm:"not null;index:idx_file_versions,priority:2;comment:版本号" json:"version_number"`
	PhysicalFileID uint  `gorm:"not null;comment:物理文件ID" json:"physical_file_id"`
	FileSize       int64 `gorm:"not null;comment:文件大小" json:"file_size"`
	UploaderID     uint  `gorm:"not null;comment:上传者ID" json:"uploader_id"`
	IsCurrent      bool  `gorm:"default:false;index:idx_current_version;comment:是否当前版本" json:"is_current"`

	// 关联
	UserFile     UserFile     `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	PhysicalFile PhysicalFile `gorm:"foreignKey:PhysicalFileID" json:"physical_file,omitempty"`
	Uploader     User         `gorm:"foreignKey:UploaderID" json:"uploader,omitempty"`
}

// TableName 指定表名
func (FileVersion) TableName() string {
	return "file_versions"
}
