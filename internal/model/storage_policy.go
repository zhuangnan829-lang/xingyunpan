package model

// StoragePolicy stores admin-managed storage policy definitions.
type StoragePolicy struct {
	BaseModel
	Name               string `gorm:"size:255;not null" json:"name"`
	Type               string `gorm:"size:32;not null;default:'local'" json:"type"`
	GroupsJSON         string `gorm:"type:text" json:"groups_json"`
	BlobPath           string `gorm:"type:text" json:"blob_path"`
	BlobNamePattern    string `gorm:"type:text" json:"blob_name_pattern"`
	MaxFileSize        int    `gorm:"not null;default:0" json:"max_file_size"`
	MaxFileSizeUnit    string `gorm:"size:8;not null;default:'MB'" json:"max_file_size_unit"`
	ExtensionMode      string `gorm:"size:16;not null;default:'allow'" json:"extension_mode"`
	Extensions         string `gorm:"type:text" json:"extensions"`
	NameRuleMode       string `gorm:"size:16;not null;default:'allow'" json:"name_rule_mode"`
	NameRegex          string `gorm:"type:text" json:"name_regex"`
	ChunkSize          int    `gorm:"not null;default:25" json:"chunk_size"`
	ChunkSizeUnit      string `gorm:"size:8;not null;default:'MB'" json:"chunk_size_unit"`
	PreAllocate        bool   `gorm:"not null;default:true" json:"pre_allocate"`
	ParallelChunkCount int    `gorm:"not null;default:1" json:"parallel_chunk_count"`
	EnableCDN          bool   `gorm:"not null;default:false" json:"enable_cdn"`
	DownloadCDN        string `gorm:"type:text" json:"download_cdn"`
	EnableEncryption   bool   `gorm:"not null;default:false" json:"enable_encryption"`
	EncryptionKeyID    string `gorm:"size:255" json:"encryption_key_id"`
}

// TableName specifies the DB table name.
func (StoragePolicy) TableName() string {
	return "storage_policies"
}
