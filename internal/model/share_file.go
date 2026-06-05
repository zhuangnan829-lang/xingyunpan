// 路径: internal/model/share_file.go
package model

// ShareFile 分享文件关联模型
type ShareFile struct {
	ShareID uint `gorm:"primaryKey;comment:分享ID" json:"share_id"`
	FileID  uint `gorm:"primaryKey;comment:文件ID" json:"file_id"`

	// 关联
	Share    Share    `gorm:"foreignKey:ShareID;constraint:OnDelete:CASCADE" json:"-"`
	UserFile UserFile `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName 指定表名
func (ShareFile) TableName() string {
	return "share_files"
}
