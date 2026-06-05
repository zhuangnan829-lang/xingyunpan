package model

// FileCustomPropertyValue stores per-file custom property values for one user file.
type FileCustomPropertyValue struct {
	BaseModel
	FileID     uint   `gorm:"not null;uniqueIndex" json:"file_id"`
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	ValuesJSON string `gorm:"type:longtext" json:"values_json"`
}

// TableName specifies the DB table name.
func (FileCustomPropertyValue) TableName() string {
	return "file_custom_property_values"
}
