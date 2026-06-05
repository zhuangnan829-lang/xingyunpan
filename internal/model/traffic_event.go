package model

// TrafficEvent records file transfer bytes for dashboard traffic charts.
type TrafficEvent struct {
	BaseModel
	UserID       uint   `gorm:"index;not null;default:0" json:"user_id"`
	Direction    string `gorm:"index:idx_traffic_direction_time,priority:1;size:16;not null" json:"direction"`
	Bytes        int64  `gorm:"not null;default:0" json:"bytes"`
	Source       string `gorm:"size:32;not null;default:''" json:"source"`
	ResourceType string `gorm:"size:64;not null;default:''" json:"resource_type"`
	ResourceID   string `gorm:"size:128;not null;default:''" json:"resource_id"`
}

func (TrafficEvent) TableName() string {
	return "traffic_events"
}
