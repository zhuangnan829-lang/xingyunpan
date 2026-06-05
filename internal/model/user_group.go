package model

// UserGroup stores admin-managed user group definitions.
type UserGroup struct {
	BaseModel
	Name            string `gorm:"size:120;not null;uniqueIndex" json:"name"`
	Description     string `gorm:"type:text" json:"description"`
	StoragePolicyID uint   `gorm:"not null;default:0;index" json:"storage_policy_id"`
	MaxCapacity     int64  `gorm:"not null;default:0" json:"max_capacity"`
}

// TableName specifies the DB table name.
func (UserGroup) TableName() string {
	return "user_groups"
}
