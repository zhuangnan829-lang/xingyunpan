package model

// StoragePolicyAudit records admin changes to storage policies.
type StoragePolicyAudit struct {
	BaseModel
	StoragePolicyID uint   `gorm:"index;not null;default:0" json:"storage_policy_id"`
	Action          string `gorm:"size:32;index;not null" json:"action"`
	OperatorID      uint   `gorm:"index;not null;default:0" json:"operator_id"`
	OperatorName    string `gorm:"size:120;not null;default:''" json:"operator_name"`
	SourceAuditID   uint   `gorm:"index;not null;default:0" json:"source_audit_id"`
	BeforeJSON      string `gorm:"type:longtext" json:"before_json"`
	AfterJSON       string `gorm:"type:longtext" json:"after_json"`
	GroupsJSON      string `gorm:"type:longtext" json:"groups_json"`
	UserCount       int64  `gorm:"not null;default:0" json:"user_count"`
}

func (StoragePolicyAudit) TableName() string {
	return "storage_policy_audits"
}
