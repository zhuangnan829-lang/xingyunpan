package model

// EventSetting stores admin-managed event recording switches.
type EventSetting struct {
	BaseModel
	EnabledEventsJSON string `gorm:"type:longtext" json:"-"`
}

// TableName specifies the DB table name.
func (EventSetting) TableName() string {
	return "event_settings"
}
