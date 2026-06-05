// 路径: internal/model/collaboration.go
package model

// Collaboration 协作模型
type Collaboration struct {
	BaseModel
	FileID         uint   `gorm:"not null;uniqueIndex:unique_collaboration,priority:1;index:idx_file_owner,priority:1;comment:文件ID" json:"file_id"`
	OwnerID        uint   `gorm:"not null;index:idx_file_owner,priority:2;comment:所有者ID" json:"owner_id"`
	CollaboratorID uint   `gorm:"not null;uniqueIndex:unique_collaboration,priority:2;index:idx_collaborator;comment:协作者ID" json:"collaborator_id"`
	Permission     string `gorm:"size:20;not null;comment:权限(view/download/edit)" json:"permission"`

	// 关联
	UserFile     UserFile `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	Owner        User     `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Collaborator User     `gorm:"foreignKey:CollaboratorID" json:"collaborator,omitempty"`
}

// TableName 指定表名
func (Collaboration) TableName() string {
	return "collaborations"
}
