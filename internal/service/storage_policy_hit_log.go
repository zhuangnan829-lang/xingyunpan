package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

type StoragePolicyHitLogPayload struct {
	ID                uint                   `json:"id"`
	StoragePolicyID   uint                   `json:"storage_policy_id"`
	StoragePolicyName string                 `json:"storage_policy_name"`
	HitType           string                 `json:"hit_type"`
	Action            string                 `json:"action"`
	UserID            uint                   `json:"user_id"`
	Username          string                 `json:"username"`
	UserGroupID       uint                   `json:"user_group_id"`
	UserGroupName     string                 `json:"user_group_name"`
	FileID            uint                   `json:"file_id"`
	FileName          string                 `json:"file_name"`
	FileSize          int64                  `json:"file_size"`
	ResourceType      string                 `json:"resource_type"`
	ResourceID        string                 `json:"resource_id"`
	Config            map[string]interface{} `json:"config"`
	CreatedAt         time.Time              `json:"created_at"`
}

type StoragePolicyHitLogInput struct {
	Action       string
	UserID       uint
	FileID       uint
	FileName     string
	FileSize     int64
	ResourceType string
	ResourceID   string
	Fallback     map[string]interface{}
}

func recordStoragePolicyHit(db *gorm.DB, input StoragePolicyHitLogInput) {
	if db == nil || input.UserID == 0 || strings.TrimSpace(input.Action) == "" {
		return
	}

	go func() {
		storagePolicyLogTableMu.Lock()
		if !db.Migrator().HasTable(&model.StoragePolicyHitLog{}) {
			if err := db.AutoMigrate(&model.StoragePolicyHitLog{}); err != nil {
				storagePolicyLogTableMu.Unlock()
				return
			}
		}
		storagePolicyLogTableMu.Unlock()

		row, err := loadStoragePolicyHitContext(db, input.UserID)
		if err != nil {
			return
		}

		config := input.Fallback
		if config == nil {
			config = map[string]interface{}{}
		}
		hitType := "global_default"
		if row.StoragePolicyID > 0 {
			hitType = "user_group_policy"
			config = storagePolicyHitConfig(&model.StoragePolicy{
				Name:               row.StoragePolicyName,
				Type:               row.PolicyType,
				BlobPath:           row.BlobPath,
				BlobNamePattern:    row.BlobNamePattern,
				MaxFileSize:        row.MaxFileSize,
				MaxFileSizeUnit:    row.MaxFileSizeUnit,
				ExtensionMode:      row.ExtensionMode,
				Extensions:         row.Extensions,
				NameRuleMode:       row.NameRuleMode,
				NameRegex:          row.NameRegex,
				ChunkSize:          row.ChunkSize,
				ChunkSizeUnit:      row.ChunkSizeUnit,
				PreAllocate:        row.PreAllocate,
				ParallelChunkCount: row.ParallelChunkCount,
				EnableCDN:          row.EnableCDN,
				DownloadCDN:        row.DownloadCDN,
				EnableEncryption:   row.EnableEncryption,
				EncryptionKeyID:    row.EncryptionKeyID,
			})
		}
		config["hit_type"] = hitType

		configJSON, _ := json.Marshal(config)
		log := model.StoragePolicyHitLog{
			StoragePolicyID:   row.StoragePolicyID,
			StoragePolicyName: strings.TrimSpace(row.StoragePolicyName),
			HitType:           hitType,
			Action:            strings.TrimSpace(input.Action),
			UserID:            input.UserID,
			Username:          strings.TrimSpace(row.Username),
			UserGroupID:       row.UserGroupID,
			UserGroupName:     strings.TrimSpace(row.UserGroupName),
			FileID:            input.FileID,
			FileName:          strings.TrimSpace(input.FileName),
			FileSize:          input.FileSize,
			ResourceType:      strings.TrimSpace(input.ResourceType),
			ResourceID:        strings.TrimSpace(input.ResourceID),
			ConfigJSON:        string(configJSON),
		}
		if err := db.Create(&log).Error; err == nil && log.ID%50 == 0 {
			cutoff := time.Now().AddDate(0, 0, -30)
			_ = db.Where("created_at < ?", cutoff).Delete(&model.StoragePolicyHitLog{}).Error
		}
	}()
}

type storagePolicyHitContextRow struct {
	UserID             uint
	Username           string
	UserGroupID        uint
	UserGroupName      string
	StoragePolicyID    uint
	StoragePolicyName  string
	PolicyType         string
	BlobPath           string
	BlobNamePattern    string
	MaxFileSize        int
	MaxFileSizeUnit    string
	ExtensionMode      string
	Extensions         string
	NameRuleMode       string
	NameRegex          string
	ChunkSize          int
	ChunkSizeUnit      string
	PreAllocate        bool
	ParallelChunkCount int
	EnableCDN          bool
	DownloadCDN        string
	EnableEncryption   bool
	EncryptionKeyID    string
}

func loadStoragePolicyHitContext(db *gorm.DB, userID uint) (*storagePolicyHitContextRow, error) {
	var row storagePolicyHitContextRow
	if err := db.Table("users").
		Select(`
			users.id AS user_id,
			users.username AS username,
			COALESCE(user_groups.id, 0) AS user_group_id,
			COALESCE(user_groups.name, '') AS user_group_name,
			COALESCE(storage_policies.id, 0) AS storage_policy_id,
			COALESCE(storage_policies.name, '') AS storage_policy_name,
			COALESCE(storage_policies.type, '') AS policy_type,
			COALESCE(storage_policies.blob_path, '') AS blob_path,
			COALESCE(storage_policies.blob_name_pattern, '') AS blob_name_pattern,
			COALESCE(storage_policies.max_file_size, 0) AS max_file_size,
			COALESCE(storage_policies.max_file_size_unit, '') AS max_file_size_unit,
			COALESCE(storage_policies.extension_mode, '') AS extension_mode,
			COALESCE(storage_policies.extensions, '') AS extensions,
			COALESCE(storage_policies.name_rule_mode, '') AS name_rule_mode,
			COALESCE(storage_policies.name_regex, '') AS name_regex,
			COALESCE(storage_policies.chunk_size, 0) AS chunk_size,
			COALESCE(storage_policies.chunk_size_unit, '') AS chunk_size_unit,
			COALESCE(storage_policies.pre_allocate, FALSE) AS pre_allocate,
			COALESCE(storage_policies.parallel_chunk_count, 0) AS parallel_chunk_count,
			COALESCE(storage_policies.enable_cdn, FALSE) AS enable_cdn,
			COALESCE(storage_policies.download_cdn, '') AS download_cdn,
			COALESCE(storage_policies.enable_encryption, FALSE) AS enable_encryption,
			COALESCE(storage_policies.encryption_key_id, '') AS encryption_key_id
		`).
		Joins("LEFT JOIN user_groups ON user_groups.id = users.user_group_id").
		Joins("LEFT JOIN storage_policies ON storage_policies.id = user_groups.storage_policy_id").
		Where("users.id = ?", userID).
		Scan(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func storagePolicyHitConfig(policy *model.StoragePolicy) map[string]interface{} {
	if policy == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"type":                 strings.TrimSpace(policy.Type),
		"blob_path":            strings.TrimSpace(policy.BlobPath),
		"blob_name_pattern":    strings.TrimSpace(policy.BlobNamePattern),
		"max_file_size":        policy.MaxFileSize,
		"max_file_size_unit":   strings.TrimSpace(policy.MaxFileSizeUnit),
		"extension_mode":       strings.TrimSpace(policy.ExtensionMode),
		"extensions":           strings.TrimSpace(policy.Extensions),
		"name_rule_mode":       strings.TrimSpace(policy.NameRuleMode),
		"name_regex":           strings.TrimSpace(policy.NameRegex),
		"chunk_size":           policy.ChunkSize,
		"chunk_size_unit":      strings.TrimSpace(policy.ChunkSizeUnit),
		"pre_allocate":         policy.PreAllocate,
		"parallel_chunk_count": policy.ParallelChunkCount,
		"enable_cdn":           policy.EnableCDN,
		"download_cdn":         strings.TrimSpace(policy.DownloadCDN),
		"enable_encryption":    policy.EnableEncryption,
		"encryption_key_id":    strings.TrimSpace(policy.EncryptionKeyID),
	}
}

func toStoragePolicyHitLogPayload(row *model.StoragePolicyHitLog) StoragePolicyHitLogPayload {
	config := map[string]interface{}{}
	if strings.TrimSpace(row.ConfigJSON) != "" {
		_ = json.Unmarshal([]byte(row.ConfigJSON), &config)
	}
	return StoragePolicyHitLogPayload{
		ID:                row.ID,
		StoragePolicyID:   row.StoragePolicyID,
		StoragePolicyName: row.StoragePolicyName,
		HitType:           row.HitType,
		Action:            row.Action,
		UserID:            row.UserID,
		Username:          row.Username,
		UserGroupID:       row.UserGroupID,
		UserGroupName:     row.UserGroupName,
		FileID:            row.FileID,
		FileName:          row.FileName,
		FileSize:          row.FileSize,
		ResourceType:      row.ResourceType,
		ResourceID:        row.ResourceID,
		Config:            config,
		CreatedAt:         row.CreatedAt,
	}
}

func defaultStoragePolicyHitFallbackConfig(extra map[string]interface{}) map[string]interface{} {
	config := map[string]interface{}{
		"source": "global_default",
	}
	for key, value := range extra {
		config[key] = value
	}
	return config
}

func formatStoragePolicyFileResourceID(id uint) string {
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("%d", id)
}
