package model

import "time"

const (
	BlobScanTaskStatusRunning   = "running"
	BlobScanTaskStatusCompleted = "completed"
	BlobScanTaskStatusFailed    = "failed"
)

// BlobScanTask stores the latest admin Blob asset inspection result.
type BlobScanTask struct {
	BaseModel
	Status               string     `gorm:"size:20;not null;default:'running';index" json:"status"`
	Progress             int        `gorm:"not null;default:0" json:"progress"`
	TotalPhysicalFiles   int64      `gorm:"not null;default:0" json:"total_physical_files"`
	ScannedPhysicalFiles int64      `gorm:"not null;default:0" json:"scanned_physical_files"`
	StorageFileCount     int64      `gorm:"not null;default:0" json:"storage_file_count"`
	OrphanCount          int64      `gorm:"not null;default:0" json:"orphan_count"`
	MissingOnStorage     int64      `gorm:"not null;default:0" json:"missing_on_storage"`
	RefCountMismatch     int64      `gorm:"not null;default:0" json:"ref_count_mismatch"`
	DuplicateHash        int64      `gorm:"not null;default:0" json:"duplicate_hash"`
	ZeroSize             int64      `gorm:"not null;default:0" json:"zero_size"`
	InvalidPath          int64      `gorm:"not null;default:0" json:"invalid_path"`
	ExtraStorageFiles    int64      `gorm:"not null;default:0" json:"extra_storage_files"`
	StartedAt            time.Time  `gorm:"not null" json:"started_at"`
	FinishedAt           *time.Time `json:"finished_at"`
	ResultJSON           string     `gorm:"type:longtext" json:"result_json"`
	LastError            string     `gorm:"type:text" json:"last_error"`
}

func (BlobScanTask) TableName() string {
	return "blob_scan_tasks"
}
