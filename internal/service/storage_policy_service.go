package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"

	"gorm.io/gorm"
)

// StoragePolicyPayload is the API-facing storage policy shape.
type StoragePolicyPayload struct {
	ID                 uint     `json:"id"`
	Name               string   `json:"name"`
	Type               string   `json:"type"`
	Groups             []string `json:"groups"`
	EffectiveUserCount int64    `json:"effective_user_count"`
	EffectiveFileCount int64    `json:"effective_file_count"`
	BlobPath           string   `json:"blob_path"`
	BlobNamePattern    string   `json:"blob_name_pattern"`
	MaxFileSize        int      `json:"max_file_size"`
	MaxFileSizeUnit    string   `json:"max_file_size_unit"`
	ExtensionMode      string   `json:"extension_mode"`
	Extensions         string   `json:"extensions"`
	NameRuleMode       string   `json:"name_rule_mode"`
	NameRegex          string   `json:"name_regex"`
	ChunkSize          int      `json:"chunk_size"`
	ChunkSizeUnit      string   `json:"chunk_size_unit"`
	PreAllocate        bool     `json:"pre_allocate"`
	ParallelChunkCount int      `json:"parallel_chunk_count"`
	EnableCDN          bool     `json:"enable_cdn"`
	DownloadCDN        string   `json:"download_cdn"`
	EnableEncryption   bool     `json:"enable_encryption"`
	EncryptionKeyID    string   `json:"encryption_key_id"`
}

type StoragePolicyPreviewPayload struct {
	Policy             StoragePolicyPayload         `json:"policy"`
	Groups             []StoragePolicyGroupCoverage `json:"groups"`
	UserCount          int64                        `json:"user_count"`
	ExistingFileCount  int64                        `json:"existing_file_count"`
	NewUploadConfig    StoragePolicyUploadPreview   `json:"new_upload_config"`
	HistoricalBlobNote string                       `json:"historical_blob_note"`
}

type StoragePolicyGroupCoverage struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	UserCount int64  `json:"user_count"`
	FileCount int64  `json:"file_count"`
}

type StoragePolicyUploadPreview struct {
	MaxFileSize        int    `json:"max_file_size"`
	MaxFileSizeUnit    string `json:"max_file_size_unit"`
	ExtensionMode      string `json:"extension_mode"`
	Extensions         string `json:"extensions"`
	NameRuleMode       string `json:"name_rule_mode"`
	NameRegex          string `json:"name_regex"`
	ChunkSize          int    `json:"chunk_size"`
	ChunkSizeUnit      string `json:"chunk_size_unit"`
	ParallelChunkCount int    `json:"parallel_chunk_count"`
	PreAllocate        bool   `json:"pre_allocate"`
	EnableCDN          bool   `json:"enable_cdn"`
	DownloadCDN        string `json:"download_cdn"`
	EnableEncryption   bool   `json:"enable_encryption"`
	EncryptionKeyID    string `json:"encryption_key_id"`
}

type StoragePolicyActor struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type StoragePolicyAuditPayload struct {
	ID              uint                         `json:"id"`
	StoragePolicyID uint                         `json:"storage_policy_id"`
	Action          string                       `json:"action"`
	OperatorID      uint                         `json:"operator_id"`
	OperatorName    string                       `json:"operator_name"`
	SourceAuditID   uint                         `json:"source_audit_id"`
	Before          *StoragePolicyPayload        `json:"before,omitempty"`
	After           *StoragePolicyPayload        `json:"after,omitempty"`
	Groups          []StoragePolicyGroupCoverage `json:"groups"`
	UserCount       int64                        `json:"user_count"`
	CreatedAt       time.Time                    `json:"created_at"`
}

type StoragePolicyGroupMigrationPayload struct {
	SourcePolicyID uint                         `json:"source_policy_id"`
	TargetPolicyID uint                         `json:"target_policy_id"`
	Groups         []StoragePolicyGroupCoverage `json:"groups"`
	UserCount      int64                        `json:"user_count"`
}

// StoragePolicyService provides admin CRUD for storage policies.
type StoragePolicyService interface {
	List() ([]StoragePolicyPayload, error)
	Get(id uint) (*StoragePolicyPayload, error)
	Preview(id uint) (*StoragePolicyPreviewPayload, error)
	History(id uint, limit int) ([]StoragePolicyAuditPayload, error)
	RecentHits(id uint, limit int) ([]StoragePolicyHitLogPayload, error)
	AuditDetail(id uint, auditID uint) (*StoragePolicyAuditPayload, error)
	Rollback(id uint, auditID uint, actor StoragePolicyActor) (*StoragePolicyPayload, error)
	MigrateGroups(id uint, targetPolicyID uint, groupIDs []uint, actor StoragePolicyActor) (*StoragePolicyGroupMigrationPayload, error)
	RepairLegacyDefaults(actor StoragePolicyActor) ([]StoragePolicyAuditPayload, error)
	Copy(id uint, actor StoragePolicyActor) (*StoragePolicyPayload, error)
	Export(id uint) (*StoragePolicyPayload, error)
	Import(payload *StoragePolicyPayload, overwriteID uint, actor StoragePolicyActor) (*StoragePolicyPayload, error)
	Create(payload *StoragePolicyPayload, actor StoragePolicyActor) (*StoragePolicyPayload, error)
	Update(id uint, payload *StoragePolicyPayload, actor StoragePolicyActor) (*StoragePolicyPayload, error)
	Delete(id uint, actor StoragePolicyActor) error
}

type storagePolicyService struct {
	repo repository.StoragePolicyRepository
	db   *gorm.DB
}

var storagePolicyLogTableMu sync.Mutex

// NewStoragePolicyService creates a storage policy service.
func NewStoragePolicyService(repo repository.StoragePolicyRepository, db ...*gorm.DB) StoragePolicyService {
	var database *gorm.DB
	if len(db) > 0 {
		database = db[0]
	}
	return &storagePolicyService{repo: repo, db: database}
}

func (s *storagePolicyService) List() ([]StoragePolicyPayload, error) {
	items, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		seed := defaultStoragePolicyPayload()
		created, err := s.Create(&seed, StoragePolicyActor{Name: "system"})
		if err != nil {
			return nil, err
		}
		return []StoragePolicyPayload{*created}, nil
	}

	result := make([]StoragePolicyPayload, 0, len(items))
	for _, item := range items {
		result = append(result, s.withEffectiveCoverage(toStoragePolicyPayload(&item)))
	}
	return result, nil
}

func (s *storagePolicyService) Get(id uint) (*StoragePolicyPayload, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("storage policy not found")
	}
	payload := s.withEffectiveCoverage(toStoragePolicyPayload(item))
	return &payload, nil
}

func (s *storagePolicyService) Preview(id uint) (*StoragePolicyPreviewPayload, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("storage policy not found")
	}

	policy := s.withEffectiveCoverage(toStoragePolicyPayload(item))
	groups, userCount, fileCount, err := s.effectiveGroupCoverage(id)
	if err != nil {
		return nil, err
	}

	return &StoragePolicyPreviewPayload{
		Policy:            policy,
		Groups:            groups,
		UserCount:         userCount,
		ExistingFileCount: fileCount,
		NewUploadConfig: StoragePolicyUploadPreview{
			MaxFileSize:        policy.MaxFileSize,
			MaxFileSizeUnit:    policy.MaxFileSizeUnit,
			ExtensionMode:      policy.ExtensionMode,
			Extensions:         policy.Extensions,
			NameRuleMode:       policy.NameRuleMode,
			NameRegex:          policy.NameRegex,
			ChunkSize:          policy.ChunkSize,
			ChunkSizeUnit:      policy.ChunkSizeUnit,
			ParallelChunkCount: policy.ParallelChunkCount,
			PreAllocate:        policy.PreAllocate,
			EnableCDN:          policy.EnableCDN,
			DownloadCDN:        policy.DownloadCDN,
			EnableEncryption:   policy.EnableEncryption,
			EncryptionKeyID:    policy.EncryptionKeyID,
		},
		HistoricalBlobNote: "Existing physical blob metadata is kept unchanged; new uploads use this policy immediately.",
	}, nil
}

func (s *storagePolicyService) History(id uint, limit int) ([]StoragePolicyAuditPayload, error) {
	if s.db == nil {
		return []StoragePolicyAuditPayload{}, nil
	}
	if err := s.ensureStoragePolicyLogTables(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	var rows []model.StoragePolicyAudit
	if err := s.db.Where("storage_policy_id = ?", id).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load storage policy audits failed: %w", err)
	}

	result := make([]StoragePolicyAuditPayload, 0, len(rows))
	for _, row := range rows {
		payload, err := toStoragePolicyAuditPayload(&row)
		if err != nil {
			return nil, err
		}
		result = append(result, payload)
	}
	return result, nil
}

func (s *storagePolicyService) RecentHits(id uint, limit int) ([]StoragePolicyHitLogPayload, error) {
	if s.db == nil {
		return []StoragePolicyHitLogPayload{}, nil
	}
	if err := s.ensureStoragePolicyLogTables(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var rows []model.StoragePolicyHitLog
	if err := s.db.Where("storage_policy_id = ?", id).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load storage policy hit logs failed: %w", err)
	}

	result := make([]StoragePolicyHitLogPayload, 0, len(rows))
	for _, row := range rows {
		result = append(result, toStoragePolicyHitLogPayload(&row))
	}
	return result, nil
}

func (s *storagePolicyService) AuditDetail(id uint, auditID uint) (*StoragePolicyAuditPayload, error) {
	if s.db == nil {
		return nil, fmt.Errorf("storage policy audit storage is unavailable")
	}
	if err := s.ensureStoragePolicyLogTables(); err != nil {
		return nil, err
	}

	var row model.StoragePolicyAudit
	if err := s.db.Where("id = ? AND storage_policy_id = ?", auditID, id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("storage policy audit not found")
		}
		return nil, fmt.Errorf("load storage policy audit failed: %w", err)
	}

	payload, err := toStoragePolicyAuditPayload(&row)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

func (s *storagePolicyService) Rollback(id uint, auditID uint, actor StoragePolicyActor) (*StoragePolicyPayload, error) {
	audit, err := s.AuditDetail(id, auditID)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(audit.Action, "delete") || audit.After == nil {
		return nil, fmt.Errorf("deleted storage policy versions cannot be rolled back")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("storage policy not found")
	}
	before := s.withEffectiveCoverage(toStoragePolicyPayload(existing))

	target := *audit.After
	target.ID = id
	normalized, err := normalizeStoragePolicyPayload(&target)
	if err != nil {
		return nil, err
	}

	item := toStoragePolicyModel(normalized, existing)
	item.ID = id
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}

	result := s.withEffectiveCoverage(toStoragePolicyPayload(item))
	if err := s.recordAudit(id, "rollback", actor, &before, &result, auditID); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *storagePolicyService) MigrateGroups(id uint, targetPolicyID uint, groupIDs []uint, actor StoragePolicyActor) (*StoragePolicyGroupMigrationPayload, error) {
	if s.db == nil {
		return nil, fmt.Errorf("storage policy database is unavailable")
	}
	if id == 0 || targetPolicyID == 0 {
		return nil, fmt.Errorf("source and target storage policies are required")
	}
	if id == targetPolicyID {
		return nil, fmt.Errorf("target storage policy must be different from source policy")
	}

	source, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, fmt.Errorf("source storage policy not found")
	}
	target, err := s.repo.GetByID(targetPolicyID)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, fmt.Errorf("target storage policy not found")
	}

	groups, userCount, _, err := s.groupCoverageForMigration(id, groupIDs)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("no user groups are bound to this storage policy")
	}

	ids := make([]uint, 0, len(groups))
	for _, group := range groups {
		ids = append(ids, group.ID)
	}

	if err := s.db.Model(&model.UserGroup{}).
		Where("storage_policy_id = ? AND id IN ?", id, ids).
		Update("storage_policy_id", targetPolicyID).Error; err != nil {
		return nil, fmt.Errorf("migrate user groups failed: %w", err)
	}

	before := s.withEffectiveCoverage(toStoragePolicyPayload(source))
	after := s.withEffectiveCoverage(toStoragePolicyPayload(target))
	if err := s.recordAuditWithCoverage(id, "migrate_groups", actor, &before, &after, groups, userCount, 0); err != nil {
		return nil, err
	}

	return &StoragePolicyGroupMigrationPayload{
		SourcePolicyID: id,
		TargetPolicyID: targetPolicyID,
		Groups:         groups,
		UserCount:      userCount,
	}, nil
}

func (s *storagePolicyService) RepairLegacyDefaults(actor StoragePolicyActor) ([]StoragePolicyAuditPayload, error) {
	items, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []StoragePolicyAuditPayload{}, nil
	}

	repaired := make([]StoragePolicyAuditPayload, 0)
	for _, item := range items {
		before := s.withEffectiveCoverage(toStoragePolicyPayload(&item))
		after := before
		changed := repairLegacyStoragePolicyPayload(&after)
		if !changed {
			continue
		}

		normalized, err := normalizeStoragePolicyPayload(&after)
		if err != nil {
			return nil, fmt.Errorf("normalize repaired storage policy %d failed: %w", item.ID, err)
		}
		modelItem := toStoragePolicyModel(normalized, &item)
		modelItem.ID = item.ID
		if err := s.repo.Save(modelItem); err != nil {
			return nil, err
		}
		result := s.withEffectiveCoverage(toStoragePolicyPayload(modelItem))
		if err := s.recordAudit(item.ID, "repair_legacy", actor, &before, &result, 0); err != nil {
			return nil, err
		}
		audits, err := s.History(item.ID, 1)
		if err != nil {
			return nil, err
		}
		if len(audits) > 0 {
			repaired = append(repaired, audits[0])
		}
	}
	return repaired, nil
}

func (s *storagePolicyService) Copy(id uint, actor StoragePolicyActor) (*StoragePolicyPayload, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("storage policy not found")
	}

	source := s.withEffectiveCoverage(toStoragePolicyPayload(existing))
	copied := source
	copied.ID = 0
	copied.EffectiveUserCount = 0
	copied.EffectiveFileCount = 0
	copied.Name = s.uniqueStoragePolicyName(fmt.Sprintf("%s 副本", strings.TrimSpace(source.Name)), 0)
	if err := validateStoragePolicyImportPayload(&copied); err != nil {
		return nil, err
	}

	item := toStoragePolicyModel(&copied, nil)
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	result := s.withEffectiveCoverage(toStoragePolicyPayload(item))
	if err := s.recordAudit(item.ID, "copy", actor, nil, &result, 0); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *storagePolicyService) Export(id uint) (*StoragePolicyPayload, error) {
	payload, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	exported := *payload
	return &exported, nil
}

func (s *storagePolicyService) Import(payload *StoragePolicyPayload, overwriteID uint, actor StoragePolicyActor) (*StoragePolicyPayload, error) {
	normalized, err := normalizeStoragePolicyPayload(payload)
	if err != nil {
		return nil, err
	}
	if err := validateStoragePolicyImportPayload(normalized); err != nil {
		return nil, err
	}
	if err := s.validateStoragePolicyImportGroupConflicts(normalized.Groups, overwriteID); err != nil {
		return nil, err
	}

	if overwriteID > 0 {
		existing, err := s.repo.GetByID(overwriteID)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			return nil, fmt.Errorf("storage policy not found")
		}
		before := s.withEffectiveCoverage(toStoragePolicyPayload(existing))
		normalized.ID = overwriteID
		item := toStoragePolicyModel(normalized, existing)
		item.ID = overwriteID
		if err := s.repo.Save(item); err != nil {
			return nil, err
		}
		result := s.withEffectiveCoverage(toStoragePolicyPayload(item))
		if err := s.recordAudit(overwriteID, "import_overwrite", actor, &before, &result, 0); err != nil {
			return nil, err
		}
		return &result, nil
	}

	normalized.ID = 0
	normalized.Name = s.uniqueStoragePolicyName(normalized.Name, 0)
	item := toStoragePolicyModel(normalized, nil)
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	result := s.withEffectiveCoverage(toStoragePolicyPayload(item))
	if err := s.recordAudit(item.ID, "import", actor, nil, &result, 0); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *storagePolicyService) Create(payload *StoragePolicyPayload, actor StoragePolicyActor) (*StoragePolicyPayload, error) {
	normalized, err := normalizeStoragePolicyPayload(payload)
	if err != nil {
		return nil, err
	}

	item := toStoragePolicyModel(normalized, nil)
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	result := s.withEffectiveCoverage(toStoragePolicyPayload(item))
	if err := s.recordAudit(item.ID, "create", actor, nil, &result, 0); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *storagePolicyService) Update(id uint, payload *StoragePolicyPayload, actor StoragePolicyActor) (*StoragePolicyPayload, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("storage policy not found")
	}
	before := s.withEffectiveCoverage(toStoragePolicyPayload(existing))

	normalized, err := normalizeStoragePolicyPayload(payload)
	if err != nil {
		return nil, err
	}

	item := toStoragePolicyModel(normalized, existing)
	item.ID = id
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}
	result := s.withEffectiveCoverage(toStoragePolicyPayload(item))
	if err := s.recordAudit(id, "update", actor, &before, &result, 0); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *storagePolicyService) Delete(id uint, actor StoragePolicyActor) error {
	items, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(items) <= 1 {
		return fmt.Errorf("at least one storage policy must be retained")
	}

	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("storage policy not found")
	}

	groups, userCount, _, err := s.effectiveGroupCoverage(id)
	if err != nil {
		return err
	}
	if len(groups) > 0 || userCount > 0 {
		return fmt.Errorf("storage policy is still bound to user groups; migrate user groups before deleting")
	}

	before := s.withEffectiveCoverage(toStoragePolicyPayload(existing))
	if err := s.recordAudit(id, "delete", actor, &before, nil, 0); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func defaultStoragePolicyPayload() StoragePolicyPayload {
	return StoragePolicyPayload{
		Name:               "默认存储策略",
		Type:               "local",
		Groups:             []string{"Admin", "User"},
		BlobPath:           "/cloudreve/data/uploads/{uid}/{path}",
		BlobNamePattern:    "{uid}_{randomkey8}_{originname}",
		MaxFileSize:        0,
		MaxFileSizeUnit:    "MB",
		ExtensionMode:      "allow",
		Extensions:         "",
		NameRuleMode:       "allow",
		NameRegex:          "",
		ChunkSize:          25,
		ChunkSizeUnit:      "MB",
		PreAllocate:        true,
		ParallelChunkCount: 1,
		EnableCDN:          false,
		DownloadCDN:        "",
		EnableEncryption:   false,
		EncryptionKeyID:    "",
	}
}

func normalizeStoragePolicyPayload(payload *StoragePolicyPayload) (*StoragePolicyPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("storage policy payload cannot be nil")
	}

	normalized := *payload
	normalized.Name = strings.TrimSpace(normalized.Name)
	normalized.Type = strings.ToLower(strings.TrimSpace(normalized.Type))
	normalized.BlobPath = strings.TrimSpace(normalized.BlobPath)
	normalized.BlobNamePattern = strings.TrimSpace(normalized.BlobNamePattern)
	normalized.MaxFileSizeUnit = strings.ToUpper(strings.TrimSpace(normalized.MaxFileSizeUnit))
	normalized.ExtensionMode = strings.ToLower(strings.TrimSpace(normalized.ExtensionMode))
	normalized.Extensions = strings.TrimSpace(normalized.Extensions)
	normalized.NameRuleMode = strings.ToLower(strings.TrimSpace(normalized.NameRuleMode))
	normalized.NameRegex = strings.TrimSpace(normalized.NameRegex)
	normalized.ChunkSizeUnit = strings.ToUpper(strings.TrimSpace(normalized.ChunkSizeUnit))
	normalized.DownloadCDN = strings.TrimSpace(normalized.DownloadCDN)
	normalized.EncryptionKeyID = strings.TrimSpace(normalized.EncryptionKeyID)

	if normalized.Name == "" {
		return nil, fmt.Errorf("storage policy name cannot be empty")
	}
	if normalized.BlobPath == "" {
		return nil, fmt.Errorf("blob path cannot be empty")
	}
	if normalized.BlobNamePattern == "" {
		return nil, fmt.Errorf("blob name pattern cannot be empty")
	}

	switch normalized.Type {
	case "local", "remote", "oss", "onedrive", "cos", "s3", "obs", "balance":
	default:
		return nil, fmt.Errorf("unsupported storage policy type")
	}

	switch normalized.MaxFileSizeUnit {
	case "KB", "MB", "GB":
	default:
		return nil, fmt.Errorf("unsupported max file size unit")
	}

	switch normalized.ExtensionMode {
	case "allow", "deny":
	default:
		return nil, fmt.Errorf("unsupported extension mode")
	}

	switch normalized.NameRuleMode {
	case "allow", "deny":
	default:
		return nil, fmt.Errorf("unsupported file name rule mode")
	}

	switch normalized.ChunkSizeUnit {
	case "KB", "MB", "GB":
	default:
		return nil, fmt.Errorf("unsupported chunk size unit")
	}

	normalized.MaxFileSize = clampInt(normalized.MaxFileSize, 0, 1048576)
	normalized.ChunkSize = clampInt(normalized.ChunkSize, 0, 1048576)
	normalized.ParallelChunkCount = clampInt(normalized.ParallelChunkCount, 1, 64)
	normalized.Groups = normalizeStoragePolicyGroups(normalized.Groups)
	if len(normalized.Groups) == 0 {
		normalized.Groups = []string{"Admin", "User"}
	}
	return &normalized, nil
}

func validateStoragePolicyImportPayload(payload *StoragePolicyPayload) error {
	if payload == nil {
		return fmt.Errorf("storage policy payload cannot be nil")
	}
	if err := validateStoragePolicyTemplate("blob path", payload.BlobPath, true); err != nil {
		return err
	}
	if err := validateStoragePolicyTemplate("blob name", payload.BlobNamePattern, false); err != nil {
		return err
	}
	if strings.TrimSpace(payload.NameRegex) != "" {
		if _, err := regexp.Compile(payload.NameRegex); err != nil {
			return fmt.Errorf("invalid file name regex: %w", err)
		}
	}
	if payload.EnableCDN {
		base := strings.TrimSpace(payload.DownloadCDN)
		if base == "" {
			return fmt.Errorf("download CDN must be set when CDN download is enabled")
		}
		parsed, err := url.Parse(base)
		if err != nil || parsed == nil || parsed.Host == "" {
			return fmt.Errorf("download CDN must be a valid URL")
		}
		scheme := strings.ToLower(parsed.Scheme)
		if scheme != "http" && scheme != "https" {
			return fmt.Errorf("download CDN only supports http or https")
		}
	}
	if payload.EnableEncryption {
		keyID := strings.TrimSpace(payload.EncryptionKeyID)
		if len(keyID) > 255 || strings.Contains(keyID, "\x00") || strings.ContainsAny(keyID, `/\`) {
			return fmt.Errorf("encryption key id contains illegal characters")
		}
	}
	return nil
}

func validateStoragePolicyTemplate(label string, pattern string, allowPath bool) error {
	value := strings.TrimSpace(pattern)
	if value == "" {
		return fmt.Errorf("%s cannot be empty", label)
	}
	if strings.Contains(value, "\x00") {
		return fmt.Errorf("%s contains illegal null byte", label)
	}

	allowedVariables := map[string]string{
		"uid":        "1",
		"path":       "folder/sub",
		"originname": "origin.txt",
		"randomkey8": "abc12345",
		"hash":       strings.Repeat("a", 64),
	}
	variablePattern := regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)
	for _, match := range variablePattern.FindAllStringSubmatch(value, -1) {
		if _, ok := allowedVariables[match[1]]; !ok {
			return fmt.Errorf("%s contains unsupported variable {%s}", label, match[1])
		}
	}

	rendered := value
	for key, sample := range allowedVariables {
		rendered = strings.ReplaceAll(rendered, "{"+key+"}", sample)
	}
	if strings.Contains(rendered, "{") || strings.Contains(rendered, "}") {
		return fmt.Errorf("%s contains malformed variable syntax", label)
	}

	if !allowPath {
		if strings.ContainsAny(rendered, `/\`) {
			return fmt.Errorf("%s cannot contain path separators", label)
		}
		if sanitizeStorageFileName(rendered) == "" {
			return fmt.Errorf("%s cannot resolve to an empty file name", label)
		}
		return nil
	}

	if strings.Contains(value, ":") {
		return fmt.Errorf("%s cannot contain drive letters or URL schemes", label)
	}
	for _, segment := range strings.FieldsFunc(rendered, func(r rune) bool {
		return r == '/' || r == '\\'
	}) {
		segment = strings.TrimSpace(segment)
		if segment == "." || segment == ".." {
			return fmt.Errorf("%s cannot contain relative path traversal", label)
		}
	}
	if cleanRelativeStoragePath(rendered) == "" {
		return fmt.Errorf("%s cannot resolve to an empty path", label)
	}
	return nil
}

func (s *storagePolicyService) validateStoragePolicyImportGroupConflicts(groups []string, overwriteID uint) error {
	if s.db == nil || len(groups) == 0 {
		return nil
	}

	var rows []StoragePolicyGroupCoverage
	if err := s.db.Table("user_groups").
		Select("user_groups.id AS id, user_groups.name AS name, COUNT(users.id) AS user_count").
		Joins("LEFT JOIN users ON users.user_group_id = user_groups.id").
		Where("user_groups.name IN ? AND user_groups.storage_policy_id <> ? AND user_groups.storage_policy_id <> 0", groups, overwriteID).
		Group("user_groups.id, user_groups.name").
		Order("user_groups.id asc").
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("check storage policy user group conflicts failed: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, fmt.Sprintf("%s(%d users)", row.Name, row.UserCount))
	}
	return fmt.Errorf("user group binding conflict: %s", strings.Join(names, ", "))
}

func (s *storagePolicyService) uniqueStoragePolicyName(base string, ignoreID uint) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "导入存储策略"
	}
	items, err := s.repo.List()
	if err != nil {
		return base
	}
	used := map[string]struct{}{}
	for _, item := range items {
		if ignoreID > 0 && item.ID == ignoreID {
			continue
		}
		used[item.Name] = struct{}{}
	}
	if _, ok := used[base]; !ok {
		return base
	}

	suffixes := make([]int, 0)
	prefix := base + " "
	for name := range used {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		var index int
		if _, err := fmt.Sscanf(strings.TrimPrefix(name, prefix), "%d", &index); err == nil && index > 0 {
			suffixes = append(suffixes, index)
		}
	}
	sort.Ints(suffixes)
	next := 2
	for _, index := range suffixes {
		if index == next {
			next++
		}
	}
	return fmt.Sprintf("%s %d", base, next)
}

func repairLegacyStoragePolicyPayload(payload *StoragePolicyPayload) bool {
	if payload == nil {
		return false
	}

	changed := false
	defaults := defaultStoragePolicyPayload()

	name := strings.TrimSpace(payload.Name)
	if isLegacyDefaultStoragePolicyName(name) {
		payload.Name = defaults.Name
		changed = true
	}

	if repairedType, ok := repairLegacyStoragePolicyType(payload.Type); ok {
		if payload.Type != repairedType {
			payload.Type = repairedType
			changed = true
		}
	}

	if strings.TrimSpace(payload.BlobPath) == "" || isMojibakeText(payload.BlobPath) {
		payload.BlobPath = defaults.BlobPath
		changed = true
	}
	if strings.TrimSpace(payload.BlobNamePattern) == "" || isMojibakeText(payload.BlobNamePattern) {
		payload.BlobNamePattern = defaults.BlobNamePattern
		changed = true
	}

	if !isStoragePolicySizeUnit(payload.MaxFileSizeUnit) {
		payload.MaxFileSizeUnit = defaults.MaxFileSizeUnit
		changed = true
	}
	if !isStoragePolicyMode(payload.ExtensionMode) {
		payload.ExtensionMode = defaults.ExtensionMode
		changed = true
	}
	if !isStoragePolicyMode(payload.NameRuleMode) {
		payload.NameRuleMode = defaults.NameRuleMode
		changed = true
	}
	if !isStoragePolicySizeUnit(payload.ChunkSizeUnit) {
		payload.ChunkSizeUnit = defaults.ChunkSizeUnit
		changed = true
	}
	if payload.ChunkSize <= 0 {
		payload.ChunkSize = defaults.ChunkSize
		changed = true
	}
	if payload.ParallelChunkCount <= 0 {
		payload.ParallelChunkCount = defaults.ParallelChunkCount
		changed = true
	}
	if len(normalizeStoragePolicyGroups(payload.Groups)) == 0 {
		payload.Groups = defaults.Groups
		changed = true
	}

	return changed
}

func isLegacyDefaultStoragePolicyName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return true
	}
	if name == "默认存储策略" {
		return false
	}
	switch name {
	case "Default storage policy", "default storage policy", "榛樿瀛樺偍绛栫暐":
		return true
	}
	return isMojibakeText(name) && (strings.Contains(name, "瀛") || strings.Contains(name, "偍") || strings.Contains(name, "绛") || strings.Contains(name, "榛"))
}

func repairLegacyStoragePolicyType(value string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "local", "remote", "oss", "onedrive", "cos", "s3", "obs", "balance":
		return normalized, normalized != value
	case "", "default", "local server cluster":
		return "local", true
	}

	displayTypes := map[string]string{
		"本地存储":       "local",
		"本地策略":       "local",
		"本地服务器集群":    "local",
		"从机存储":       "remote",
		"从机存储节点":     "remote",
		"阿里云 OSS":    "oss",
		"腾讯云 COS":    "cos",
		"MinIO 对象存储": "s3",
		"S3 / MinIO": "s3",
		"华为云 OBS":    "obs",
		"负载均衡":       "balance",
	}
	if repaired, ok := displayTypes[strings.TrimSpace(value)]; ok {
		return repaired, true
	}
	if isMojibakeText(value) {
		return "local", true
	}
	return normalized, false
}

func isStoragePolicySizeUnit(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "KB", "MB", "GB":
		return true
	default:
		return false
	}
}

func isStoragePolicyMode(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "allow", "deny":
		return true
	default:
		return false
	}
}

func isMojibakeText(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	markers := []string{"�", "榛", "绛", "瀛", "偍", "鎬", "鎴", "鎵", "涓", "鑺", "鏈", "杩", "闆", "傜", "嬫", "嗗", "冩"}
	for _, marker := range markers {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}

func normalizeStoragePolicyGroups(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func toStoragePolicyPayload(item *model.StoragePolicy) StoragePolicyPayload {
	if item == nil {
		return defaultStoragePolicyPayload()
	}

	var groups []string
	if strings.TrimSpace(item.GroupsJSON) != "" {
		_ = json.Unmarshal([]byte(item.GroupsJSON), &groups)
	}

	return StoragePolicyPayload{
		ID:                 item.ID,
		Name:               item.Name,
		Type:               item.Type,
		Groups:             normalizeStoragePolicyGroups(groups),
		BlobPath:           item.BlobPath,
		BlobNamePattern:    item.BlobNamePattern,
		MaxFileSize:        item.MaxFileSize,
		MaxFileSizeUnit:    item.MaxFileSizeUnit,
		ExtensionMode:      item.ExtensionMode,
		Extensions:         item.Extensions,
		NameRuleMode:       item.NameRuleMode,
		NameRegex:          item.NameRegex,
		ChunkSize:          item.ChunkSize,
		ChunkSizeUnit:      item.ChunkSizeUnit,
		PreAllocate:        item.PreAllocate,
		ParallelChunkCount: item.ParallelChunkCount,
		EnableCDN:          item.EnableCDN,
		DownloadCDN:        item.DownloadCDN,
		EnableEncryption:   item.EnableEncryption,
		EncryptionKeyID:    item.EncryptionKeyID,
	}
}

func toStoragePolicyModel(payload *StoragePolicyPayload, target *model.StoragePolicy) *model.StoragePolicy {
	item := &model.StoragePolicy{}
	if target != nil {
		*item = *target
	}

	groupsJSON, _ := json.Marshal(normalizeStoragePolicyGroups(payload.Groups))
	item.Name = payload.Name
	item.Type = payload.Type
	item.GroupsJSON = string(groupsJSON)
	item.BlobPath = payload.BlobPath
	item.BlobNamePattern = payload.BlobNamePattern
	item.MaxFileSize = payload.MaxFileSize
	item.MaxFileSizeUnit = payload.MaxFileSizeUnit
	item.ExtensionMode = payload.ExtensionMode
	item.Extensions = payload.Extensions
	item.NameRuleMode = payload.NameRuleMode
	item.NameRegex = payload.NameRegex
	item.ChunkSize = payload.ChunkSize
	item.ChunkSizeUnit = payload.ChunkSizeUnit
	item.PreAllocate = payload.PreAllocate
	item.ParallelChunkCount = payload.ParallelChunkCount
	item.EnableCDN = payload.EnableCDN
	item.DownloadCDN = payload.DownloadCDN
	item.EnableEncryption = payload.EnableEncryption
	item.EncryptionKeyID = payload.EncryptionKeyID
	return item
}

func (s *storagePolicyService) withEffectiveCoverage(payload StoragePolicyPayload) StoragePolicyPayload {
	groups, userCount, fileCount, err := s.effectiveGroupCoverage(payload.ID)
	if err != nil || len(groups) == 0 {
		return payload
	}

	payload.Groups = make([]string, 0, len(groups))
	for _, group := range groups {
		payload.Groups = append(payload.Groups, group.Name)
	}
	payload.EffectiveUserCount = userCount
	payload.EffectiveFileCount = fileCount
	return payload
}

func (s *storagePolicyService) effectiveGroupCoverage(policyID uint) ([]StoragePolicyGroupCoverage, int64, int64, error) {
	if s.db == nil || policyID == 0 {
		return nil, 0, 0, nil
	}

	var rows []StoragePolicyGroupCoverage
	if err := s.db.Table("user_groups").
		Select(`
			user_groups.id AS id,
			user_groups.name AS name,
			COUNT(DISTINCT users.id) AS user_count,
			COUNT(user_files.id) AS file_count
		`).
		Joins("LEFT JOIN users ON users.user_group_id = user_groups.id").
		Joins("LEFT JOIN user_files ON user_files.user_id = users.id AND user_files.is_folder = FALSE AND user_files.deleted_at IS NULL").
		Where("user_groups.storage_policy_id = ?", policyID).
		Group("user_groups.id, user_groups.name").
		Order("user_groups.id asc").
		Scan(&rows).Error; err != nil {
		return nil, 0, 0, fmt.Errorf("load storage policy coverage failed: %w", err)
	}

	var userCount int64
	var fileCount int64
	for _, row := range rows {
		userCount += row.UserCount
		fileCount += row.FileCount
	}
	return rows, userCount, fileCount, nil
}

func (s *storagePolicyService) groupCoverageForMigration(policyID uint, groupIDs []uint) ([]StoragePolicyGroupCoverage, int64, int64, error) {
	if s.db == nil || policyID == 0 {
		return nil, 0, 0, nil
	}

	query := s.db.Table("user_groups").
		Select(`
			user_groups.id AS id,
			user_groups.name AS name,
			COUNT(DISTINCT users.id) AS user_count,
			COUNT(user_files.id) AS file_count
		`).
		Joins("LEFT JOIN users ON users.user_group_id = user_groups.id").
		Joins("LEFT JOIN user_files ON user_files.user_id = users.id AND user_files.is_folder = FALSE AND user_files.deleted_at IS NULL").
		Where("user_groups.storage_policy_id = ?", policyID)

	if len(groupIDs) > 0 {
		query = query.Where("user_groups.id IN ?", groupIDs)
	}

	var rows []StoragePolicyGroupCoverage
	if err := query.Group("user_groups.id, user_groups.name").
		Order("user_groups.id asc").
		Scan(&rows).Error; err != nil {
		return nil, 0, 0, fmt.Errorf("load storage policy migration coverage failed: %w", err)
	}

	var userCount int64
	var fileCount int64
	for _, row := range rows {
		userCount += row.UserCount
		fileCount += row.FileCount
	}
	return rows, userCount, fileCount, nil
}

func (s *storagePolicyService) recordAudit(policyID uint, action string, actor StoragePolicyActor, before *StoragePolicyPayload, after *StoragePolicyPayload, sourceAuditID uint) error {
	if s.db == nil || policyID == 0 {
		return nil
	}

	groups, userCount, _, err := s.effectiveGroupCoverage(policyID)
	if err != nil {
		return err
	}
	return s.recordAuditWithCoverage(policyID, action, actor, before, after, groups, userCount, sourceAuditID)
}

func (s *storagePolicyService) recordAuditWithCoverage(policyID uint, action string, actor StoragePolicyActor, before *StoragePolicyPayload, after *StoragePolicyPayload, groups []StoragePolicyGroupCoverage, userCount int64, sourceAuditID uint) error {
	if s.db == nil || policyID == 0 {
		return nil
	}
	if err := s.ensureStoragePolicyLogTables(); err != nil {
		return err
	}
	beforeJSON, err := marshalStoragePolicyAuditSnapshot(before)
	if err != nil {
		return err
	}
	afterJSON, err := marshalStoragePolicyAuditSnapshot(after)
	if err != nil {
		return err
	}
	groupsJSON, err := json.Marshal(groups)
	if err != nil {
		return err
	}

	operatorName := strings.TrimSpace(actor.Name)
	if operatorName == "" && actor.ID > 0 {
		operatorName = fmt.Sprintf("user:%d", actor.ID)
	}
	if operatorName == "" {
		operatorName = "system"
	}

	return s.db.Create(&model.StoragePolicyAudit{
		StoragePolicyID: policyID,
		Action:          action,
		OperatorID:      actor.ID,
		OperatorName:    operatorName,
		SourceAuditID:   sourceAuditID,
		BeforeJSON:      beforeJSON,
		AfterJSON:       afterJSON,
		GroupsJSON:      string(groupsJSON),
		UserCount:       userCount,
	}).Error
}

func (s *storagePolicyService) ensureStoragePolicyLogTables() error {
	if s.db == nil {
		return nil
	}
	storagePolicyLogTableMu.Lock()
	defer storagePolicyLogTableMu.Unlock()

	if !s.db.Migrator().HasTable(&model.StoragePolicyAudit{}) {
		if err := s.db.AutoMigrate(&model.StoragePolicyAudit{}); err != nil {
			return fmt.Errorf("migrate storage policy audit table failed: %w", err)
		}
	}
	if !s.db.Migrator().HasTable(&model.StoragePolicyHitLog{}) {
		if err := s.db.AutoMigrate(&model.StoragePolicyHitLog{}); err != nil {
			return fmt.Errorf("migrate storage policy hit log table failed: %w", err)
		}
	}
	return nil
}

func marshalStoragePolicyAuditSnapshot(payload *StoragePolicyPayload) (string, error) {
	if payload == nil {
		return "", nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func toStoragePolicyAuditPayload(row *model.StoragePolicyAudit) (StoragePolicyAuditPayload, error) {
	payload := StoragePolicyAuditPayload{
		ID:              row.ID,
		StoragePolicyID: row.StoragePolicyID,
		Action:          row.Action,
		OperatorID:      row.OperatorID,
		OperatorName:    row.OperatorName,
		SourceAuditID:   row.SourceAuditID,
		UserCount:       row.UserCount,
		CreatedAt:       row.CreatedAt,
		Groups:          []StoragePolicyGroupCoverage{},
	}

	if strings.TrimSpace(row.BeforeJSON) != "" {
		var before StoragePolicyPayload
		if err := json.Unmarshal([]byte(row.BeforeJSON), &before); err != nil {
			return payload, fmt.Errorf("decode storage policy audit before failed: %w", err)
		}
		payload.Before = &before
	}
	if strings.TrimSpace(row.AfterJSON) != "" {
		var after StoragePolicyPayload
		if err := json.Unmarshal([]byte(row.AfterJSON), &after); err != nil {
			return payload, fmt.Errorf("decode storage policy audit after failed: %w", err)
		}
		payload.After = &after
	}
	if strings.TrimSpace(row.GroupsJSON) != "" {
		if err := json.Unmarshal([]byte(row.GroupsJSON), &payload.Groups); err != nil {
			return payload, fmt.Errorf("decode storage policy audit groups failed: %w", err)
		}
	}
	return payload, nil
}
