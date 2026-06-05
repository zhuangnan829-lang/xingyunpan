package model

import "time"

// PhysicalFile stores physical blob metadata for de-duplication and delivery.
type PhysicalFile struct {
	BaseModel
	FileHash        string     `gorm:"uniqueIndex;size:64;not null;comment:file sha256" json:"file_hash"`
	FileSize        int64      `gorm:"not null;comment:file size bytes" json:"file_size"`
	StoragePath     string     `gorm:"size:500;not null;comment:storage path" json:"storage_path"`
	RefCount        int        `gorm:"default:0;not null;comment:reference count" json:"ref_count"`
	StorageType     string     `gorm:"size:20;default:'local';not null;comment:storage type" json:"storage_type"`
	ContentType     string     `gorm:"size:100;comment:file mime type" json:"content_type"`
	ChunkCount      int        `gorm:"default:0;comment:multipart chunk count" json:"chunk_count"`
	IsMultipart     bool       `gorm:"default:false;comment:is multipart upload" json:"is_multipart"`
	Encrypted       bool       `gorm:"not null;default:false;comment:encrypted storage" json:"encrypted"`
	EncryptionKeyID string     `gorm:"size:255;comment:encryption key id" json:"encryption_key_id"`
	Locked          bool       `gorm:"not null;default:false;index;comment:admin deletion lock" json:"locked"`
	LockedReason    string     `gorm:"size:500;comment:admin lock reason" json:"locked_reason"`
	LockedAt        *time.Time `json:"locked_at"`
	LockedBy        uint       `gorm:"comment:admin user id that locked this blob" json:"locked_by"`
}

func (PhysicalFile) TableName() string {
	return "physical_files"
}
