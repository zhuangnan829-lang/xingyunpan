package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/storage"

	"gorm.io/gorm"
)

type AdminBlobListQuery struct {
	Page            int
	PageSize        int
	OwnerID         uint
	Kind            string
	StoragePolicyID uint
	Keyword         string
	MinSize         *int64
	MaxSize         *int64
	RefCountMin     *int
	RefCountMax     *int
	Encrypted       *bool
	CreatedFrom     *time.Time
	CreatedTo       *time.Time
	SortBy          string
	SortOrder       string
}

type AdminBlobLinkedFilePayload struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Extension string    `json:"extension"`
	SizeBytes int64     `json:"size_bytes"`
	OwnerID   uint      `json:"owner_id"`
	OwnerName string    `json:"owner_name"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminBlobLinkedVersionPayload struct {
	ID            uint      `json:"id"`
	FileID        uint      `json:"file_id"`
	FileName      string    `json:"file_name"`
	VersionNumber int       `json:"version_number"`
	SizeBytes     int64     `json:"size_bytes"`
	OwnerID       uint      `json:"owner_id"`
	OwnerName     string    `json:"owner_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type AdminBlobLinkedCachePayload struct {
	Source string `json:"source"`
	Path   string `json:"path"`
}

type AdminBlobReferenceSourcePayload struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AdminBlobPayload struct {
	ID                    uint                              `json:"id"`
	Kind                  string                            `json:"kind"`
	KindLabel             string                            `json:"kind_label"`
	Source                string                            `json:"source"`
	Hash                  string                            `json:"hash"`
	ContentType           string                            `json:"content_type"`
	SizeBytes             int64                             `json:"size_bytes"`
	ReferenceCount        int                               `json:"reference_count"`
	StoredReferenceCount  int                               `json:"stored_reference_count"`
	CreatedAt             time.Time                         `json:"created_at"`
	UpdatedAt             time.Time                         `json:"updated_at"`
	LastAccessedAt        *time.Time                        `json:"last_accessed_at"`
	CreatorID             uint                              `json:"creator_id"`
	CreatorName           string                            `json:"creator_name"`
	CreatorBadge          string                            `json:"creator_badge"`
	StoragePolicyID       uint                              `json:"storage_policy_id"`
	StoragePolicyName     string                            `json:"storage_policy_name"`
	StoragePolicySubtitle string                            `json:"storage_policy_subtitle"`
	StorageType           string                            `json:"storage_type"`
	Encrypted             bool                              `json:"encrypted"`
	EncryptionStatus      *MasterKeyStatusPayload           `json:"encryption_status,omitempty"`
	Locked                bool                              `json:"locked"`
	LockedReason          string                            `json:"locked_reason"`
	LockedAt              *time.Time                        `json:"locked_at"`
	LockedBy              uint                              `json:"locked_by"`
	UploadSessionID       string                            `json:"upload_session_id,omitempty"`
	LinkedFiles           []AdminBlobLinkedFilePayload      `json:"linked_files"`
	LinkedVersions        []AdminBlobLinkedVersionPayload   `json:"linked_versions"`
	LinkedCaches          []AdminBlobLinkedCachePayload     `json:"linked_caches"`
	ReferenceSources      []AdminBlobReferenceSourcePayload `json:"reference_sources"`
	CanDelete             bool                              `json:"can_delete"`
	DeleteBlockedReasons  []string                          `json:"delete_blocked_reasons"`
	MissingOnStorage      bool                              `json:"missing_on_storage"`
	HealthStatus          string                            `json:"health_status"`
}

type AdminBlobListPayload struct {
	List           []AdminBlobPayload `json:"list"`
	Total          int64              `json:"total"`
	Page           int                `json:"page"`
	PageSize       int                `json:"page_size"`
	TotalSize      int64              `json:"total_size"`
	ReferenceTotal int64              `json:"reference_total"`
	EncryptedCount int64              `json:"encrypted_count"`
	OrphanCount    int64              `json:"orphan_count"`
}

type AdminBlobBatchDeleteResultItem struct {
	ID     uint   `json:"id"`
	Reason string `json:"reason,omitempty"`
}

type AdminBlobBatchDeletePayload struct {
	Deleted []AdminBlobBatchDeleteResultItem `json:"deleted"`
	Skipped []AdminBlobBatchDeleteResultItem `json:"skipped"`
	Failed  []AdminBlobBatchDeleteResultItem `json:"failed"`
}

type AdminBlobScanIssuePayload struct {
	BlobID uint   `json:"blob_id,omitempty"`
	Path   string `json:"path,omitempty"`
	Hash   string `json:"hash,omitempty"`
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type AdminBlobScanTaskPayload struct {
	ID                   uint                        `json:"id"`
	Status               string                      `json:"status"`
	Progress             int                         `json:"progress"`
	TotalPhysicalFiles   int64                       `json:"total_physical_files"`
	ScannedPhysicalFiles int64                       `json:"scanned_physical_files"`
	StorageFileCount     int64                       `json:"storage_file_count"`
	OrphanCount          int64                       `json:"orphan_count"`
	MissingOnStorage     int64                       `json:"missing_on_storage"`
	RefCountMismatch     int64                       `json:"ref_count_mismatch"`
	DuplicateHash        int64                       `json:"duplicate_hash"`
	ZeroSize             int64                       `json:"zero_size"`
	InvalidPath          int64                       `json:"invalid_path"`
	ExtraStorageFiles    int64                       `json:"extra_storage_files"`
	StartedAt            time.Time                   `json:"started_at"`
	FinishedAt           *time.Time                  `json:"finished_at"`
	LastError            string                      `json:"last_error,omitempty"`
	Issues               []AdminBlobScanIssuePayload `json:"issues"`
}

type AdminBlobService interface {
	List(query *AdminBlobListQuery) (*AdminBlobListPayload, error)
	Get(id uint) (*AdminBlobPayload, error)
	Download(id uint) (io.ReadCloser, string, string, error)
	Delete(id uint) error
	BatchDelete(ids []uint) (*AdminBlobBatchDeletePayload, error)
	Lock(id, adminID uint, reason string) (*AdminBlobPayload, error)
	Unlock(id, adminID uint) (*AdminBlobPayload, error)
	Scan() (*AdminBlobScanTaskPayload, error)
	LatestScan() (*AdminBlobScanTaskPayload, error)
}

type adminBlobService struct {
	db         *gorm.DB
	storage    storage.Storage
	masterKeys MasterKeyResolver
}

type adminBlobStorageRoot interface {
	BasePath() string
}

func NewAdminBlobService(db *gorm.DB, blobStorage storage.Storage) AdminBlobService {
	return &adminBlobService{db: db, storage: blobStorage, masterKeys: NewFileSystemMasterKeyResolver(db, repository.NewFileSystemSettingRepository(db))}
}

type adminBlobRow struct {
	ID                 uint
	FileHash           string
	FileSize           int64
	StoragePath        string
	RefCount           int
	StorageType        string
	ContentType        string
	Encrypted          bool
	Locked             bool
	LockedReason       string
	LockedAt           *time.Time
	LockedBy           uint
	CreatedAt          time.Time
	UpdatedAt          time.Time
	UserRefCount       int
	VersionRefCount    int
	CacheRefCount      int
	TaskRefCount       int
	RealReferenceCount int
	Kind               string
	CreatorID          uint
	CreatorName        string
	StoragePolicyID    uint
	StoragePolicyName  string
}

func (s *adminBlobService) List(query *AdminBlobListQuery) (*AdminBlobListPayload, error) {
	normalized := normalizeAdminBlobListQuery(query)
	baseSQL, args := s.buildBlobBaseSQL(normalized)

	var total int64
	if err := s.db.Raw("SELECT COUNT(*) FROM ("+baseSQL+") AS blob_base", args...).Scan(&total).Error; err != nil {
		return nil, err
	}

	var summary struct {
		TotalSize      int64
		ReferenceTotal int64
		EncryptedCount int64
		OrphanCount    int64
	}
	if err := s.db.Raw(`SELECT COALESCE(SUM(file_size),0) AS total_size,
		COALESCE(SUM(real_reference_count),0) AS reference_total,
		COALESCE(SUM(CASE WHEN encrypted THEN 1 ELSE 0 END),0) AS encrypted_count,
		COALESCE(SUM(CASE WHEN kind = 'orphan' THEN 1 ELSE 0 END),0) AS orphan_count
		FROM (`+baseSQL+`) AS blob_base`, args...).Scan(&summary).Error; err != nil {
		return nil, err
	}

	orderClause := adminBlobOrderClause(normalized.SortBy, normalized.SortOrder)
	limitArgs := append(append([]interface{}{}, args...), normalized.PageSize, (normalized.Page-1)*normalized.PageSize)
	var rows []adminBlobRow
	if err := s.db.Raw("SELECT * FROM ("+baseSQL+") AS blob_base "+orderClause+" LIMIT ? OFFSET ?", limitArgs...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	list := make([]AdminBlobPayload, 0, len(rows))
	for _, row := range rows {
		item, err := s.payloadFromRow(row, false)
		if err != nil {
			return nil, err
		}
		list = append(list, *item)
	}

	return &AdminBlobListPayload{
		List:           list,
		Total:          total,
		Page:           normalized.Page,
		PageSize:       normalized.PageSize,
		TotalSize:      summary.TotalSize,
		ReferenceTotal: summary.ReferenceTotal,
		EncryptedCount: summary.EncryptedCount,
		OrphanCount:    summary.OrphanCount,
	}, nil
}

func (s *adminBlobService) Get(id uint) (*AdminBlobPayload, error) {
	if id == 0 {
		return nil, fmt.Errorf("blob id cannot be empty")
	}
	query := normalizeAdminBlobListQuery(&AdminBlobListQuery{Page: 1, PageSize: 1})
	baseSQL, args := s.buildBlobBaseSQL(query)
	args = append(args, id)
	var row adminBlobRow
	if err := s.db.Raw("SELECT * FROM ("+baseSQL+") AS blob_base WHERE id = ?", args...).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, fmt.Errorf("blob not found")
	}
	return s.payloadFromRow(row, true)
}

func (s *adminBlobService) Download(id uint) (io.ReadCloser, string, string, error) {
	detail, err := s.Get(id)
	if err != nil {
		return nil, "", "", err
	}
	if detail.MissingOnStorage {
		return nil, "", "", fmt.Errorf("blob storage file does not exist")
	}
	var physical model.PhysicalFile
	if err := s.db.First(&physical, id).Error; err != nil {
		return nil, "", "", err
	}
	reader, err := readPhysicalBlob(s.storage, &physical, s.masterKeys)
	if err != nil {
		return nil, "", "", err
	}
	fileName := s.downloadFileName(detail)
	contentType := strings.TrimSpace(physical.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(fileName))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return reader, fileName, contentType, nil
}

func (s *adminBlobService) Delete(id uint) error {
	result, err := s.deleteOne(id)
	if err != nil {
		return err
	}
	if result.Reason != "" {
		return errors.New(result.Reason)
	}
	return nil
}

func (s *adminBlobService) BatchDelete(ids []uint) (*AdminBlobBatchDeletePayload, error) {
	if len(ids) == 0 {
		return &AdminBlobBatchDeletePayload{}, nil
	}
	limit := s.maxBatchActionSize()
	if len(ids) > limit {
		return nil, fmt.Errorf("max_batch_action_size exceeded: limit %d", limit)
	}
	result := &AdminBlobBatchDeletePayload{}
	for _, id := range uniqueAdminBlobUintIDs(ids) {
		item, err := s.deleteOne(id)
		if err != nil {
			result.Failed = append(result.Failed, AdminBlobBatchDeleteResultItem{ID: id, Reason: err.Error()})
			continue
		}
		if item.Reason != "" {
			result.Skipped = append(result.Skipped, item)
		} else {
			result.Deleted = append(result.Deleted, item)
		}
	}
	return result, nil
}

func (s *adminBlobService) Lock(id, adminID uint, reason string) (*AdminBlobPayload, error) {
	if id == 0 {
		return nil, fmt.Errorf("blob id cannot be empty")
	}
	now := time.Now()
	if err := s.db.Model(&model.PhysicalFile{}).Where("id = ?", id).Updates(map[string]interface{}{
		"locked":        true,
		"locked_reason": strings.TrimSpace(reason),
		"locked_at":     &now,
		"locked_by":     adminID,
	}).Error; err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *adminBlobService) Unlock(id, adminID uint) (*AdminBlobPayload, error) {
	if id == 0 {
		return nil, fmt.Errorf("blob id cannot be empty")
	}
	if err := s.db.Model(&model.PhysicalFile{}).Where("id = ?", id).Updates(map[string]interface{}{
		"locked":        false,
		"locked_reason": "",
		"locked_at":     nil,
		"locked_by":     adminID,
	}).Error; err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *adminBlobService) Scan() (*AdminBlobScanTaskPayload, error) {
	now := time.Now()
	task := model.BlobScanTask{Status: model.BlobScanTaskStatusRunning, Progress: 1, StartedAt: now}
	if err := s.db.Create(&task).Error; err != nil {
		return nil, err
	}
	payload, scanErr := s.runScan(&task)
	if scanErr != nil {
		finished := time.Now()
		_ = s.db.Model(&task).Updates(map[string]interface{}{"status": model.BlobScanTaskStatusFailed, "progress": 100, "finished_at": &finished, "last_error": scanErr.Error()}).Error
		payload, _ = s.LatestScan()
		if payload != nil {
			return payload, scanErr
		}
		return nil, scanErr
	}
	return payload, nil
}

func (s *adminBlobService) LatestScan() (*AdminBlobScanTaskPayload, error) {
	var task model.BlobScanTask
	if err := s.db.Order("created_at desc, id desc").First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("blob scan task not found")
		}
		return nil, err
	}
	return blobScanTaskPayload(&task), nil
}

func (s *adminBlobService) buildBlobBaseSQL(q *AdminBlobListQuery) (string, []interface{}) {
	args := []interface{}{}
	where := []string{"pf.deleted_at IS NULL"}
	if q.OwnerID > 0 {
		where = append(where, "COALESCE(uf.owner_id, fv.owner_id, 0) = ?")
		args = append(args, q.OwnerID)
	}
	if q.StoragePolicyID > 0 {
		where = append(where, "COALESCE(uf.storage_policy_id, fv.storage_policy_id, 0) = ?")
		args = append(args, q.StoragePolicyID)
	}
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		where = append(where, `(pf.storage_path LIKE ? OR pf.file_hash LIKE ? OR pf.content_type LIKE ? OR EXISTS (
			SELECT 1 FROM user_files kf WHERE kf.physical_file_id = pf.id AND kf.deleted_at IS NULL AND kf.file_name LIKE ?
		))`)
		args = append(args, like, like, like, like)
	}
	if q.MinSize != nil {
		where = append(where, "pf.file_size >= ?")
		args = append(args, *q.MinSize)
	}
	if q.MaxSize != nil {
		where = append(where, "pf.file_size <= ?")
		args = append(args, *q.MaxSize)
	}
	if q.Encrypted != nil {
		where = append(where, "pf.encrypted = ?")
		args = append(args, *q.Encrypted)
	}
	if q.CreatedFrom != nil {
		where = append(where, "pf.created_at >= ?")
		args = append(args, *q.CreatedFrom)
	}
	if q.CreatedTo != nil {
		where = append(where, "pf.created_at <= ?")
		args = append(args, *q.CreatedTo)
	}

	taskRefExpr := s.adminBlobTaskRefSQL()
	kindExpr := adminBlobKindSQL(taskRefExpr)
	refExpr := "COALESCE(uf.user_ref_count,0) + COALESCE(fv.version_ref_count,0) + CASE WHEN LOWER(pf.storage_path) LIKE '%thumb%' OR LOWER(pf.storage_path) LIKE '%thumbnail%' OR LOWER(pf.storage_path) LIKE '%cache%' THEN 1 ELSE 0 END + " + taskRefExpr
	if q.Kind != "" && q.Kind != "all" {
		where = append(where, kindExpr+" = ?")
		args = append(args, q.Kind)
	}
	if q.RefCountMin != nil {
		where = append(where, refExpr+" >= ?")
		args = append(args, *q.RefCountMin)
	}
	if q.RefCountMax != nil {
		where = append(where, refExpr+" <= ?")
		args = append(args, *q.RefCountMax)
	}

	sql := `SELECT pf.id, pf.file_hash, pf.file_size, pf.storage_path, pf.ref_count, pf.storage_type,
		pf.content_type, pf.encrypted, pf.locked, pf.locked_reason, pf.locked_at, pf.locked_by, pf.created_at, pf.updated_at,
		COALESCE(uf.user_ref_count,0) AS user_ref_count,
		COALESCE(fv.version_ref_count,0) AS version_ref_count,
		CASE WHEN LOWER(pf.storage_path) LIKE '%thumb%' OR LOWER(pf.storage_path) LIKE '%thumbnail%' OR LOWER(pf.storage_path) LIKE '%cache%' THEN 1 ELSE 0 END AS cache_ref_count,
		` + taskRefExpr + ` AS task_ref_count,
		` + refExpr + ` AS real_reference_count,
		` + kindExpr + ` AS kind,
		COALESCE(uf.owner_id, fv.owner_id, 0) AS creator_id,
		COALESCE(uf.owner_name, fv.owner_name, '') AS creator_name,
		COALESCE(uf.storage_policy_id, fv.storage_policy_id, 0) AS storage_policy_id,
		COALESCE(uf.storage_policy_name, fv.storage_policy_name, '') AS storage_policy_name
	FROM physical_files pf
	LEFT JOIN (
		SELECT user_files.physical_file_id, COUNT(*) AS user_ref_count, MIN(users.id) AS owner_id,
			MIN(users.username) AS owner_name, MIN(COALESCE(storage_policies.id,0)) AS storage_policy_id,
			MIN(COALESCE(storage_policies.name,'')) AS storage_policy_name
		FROM user_files
		LEFT JOIN users ON users.id = user_files.user_id
		LEFT JOIN user_groups ON user_groups.id = users.user_group_id
		LEFT JOIN storage_policies ON storage_policies.id = user_groups.storage_policy_id
		WHERE user_files.physical_file_id IS NOT NULL AND user_files.deleted_at IS NULL
		GROUP BY user_files.physical_file_id
	) uf ON uf.physical_file_id = pf.id
	LEFT JOIN (
		SELECT file_versions.physical_file_id, COUNT(*) AS version_ref_count, MIN(users.id) AS owner_id,
			MIN(users.username) AS owner_name, MIN(COALESCE(storage_policies.id,0)) AS storage_policy_id,
			MIN(COALESCE(storage_policies.name,'')) AS storage_policy_name
		FROM file_versions
		LEFT JOIN users ON users.id = file_versions.uploader_id
		LEFT JOIN user_groups ON user_groups.id = users.user_group_id
		LEFT JOIN storage_policies ON storage_policies.id = user_groups.storage_policy_id
		WHERE file_versions.deleted_at IS NULL
		GROUP BY file_versions.physical_file_id
	) fv ON fv.physical_file_id = pf.id
	WHERE ` + strings.Join(where, " AND ")
	return sql, args
}

func (s *adminBlobService) adminBlobTaskRefSQL() string {
	if s.db != nil && s.db.Dialector != nil && s.db.Dialector.Name() == "sqlite" {
		return "CASE WHEN EXISTS (SELECT 1 FROM queue_jobs qj WHERE qj.deleted_at IS NULL AND (qj.resource_id = CAST(pf.id AS TEXT) OR (pf.storage_path <> '' AND qj.payload LIKE ('%' || pf.storage_path || '%')))) THEN 1 ELSE 0 END"
	}
	return "CASE WHEN EXISTS (SELECT 1 FROM queue_jobs qj WHERE qj.deleted_at IS NULL AND (qj.resource_id = CAST(pf.id AS CHAR) OR (pf.storage_path <> '' AND qj.payload LIKE CONCAT('%', pf.storage_path, '%')))) THEN 1 ELSE 0 END"
}

func adminBlobKindSQL(taskRefExpr string) string {
	cacheExpr := "LOWER(pf.storage_path) LIKE '%thumb%' OR LOWER(pf.storage_path) LIKE '%thumbnail%' OR LOWER(pf.storage_path) LIKE '%cache%'"
	return "CASE WHEN COALESCE(uf.user_ref_count,0) > 0 THEN 'file' WHEN COALESCE(fv.version_ref_count,0) > 0 THEN 'version' WHEN " + cacheExpr + " THEN 'thumbnail' WHEN " + taskRefExpr + " > 0 THEN 'file' ELSE 'orphan' END"
}

func (s *adminBlobService) payloadFromRow(row adminBlobRow, includeDetail bool) (*AdminBlobPayload, error) {
	item := &AdminBlobPayload{
		ID:                    row.ID,
		Kind:                  row.Kind,
		KindLabel:             adminBlobKindLabel(row.Kind),
		Source:                strings.TrimSpace(row.StoragePath),
		Hash:                  row.FileHash,
		ContentType:           row.ContentType,
		SizeBytes:             row.FileSize,
		ReferenceCount:        row.RealReferenceCount,
		StoredReferenceCount:  row.RefCount,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		CreatorID:             row.CreatorID,
		CreatorName:           fallbackString(row.CreatorName, "system"),
		CreatorBadge:          creatorBadge(row.CreatorName),
		StoragePolicyID:       row.StoragePolicyID,
		StoragePolicyName:     fallbackString(row.StoragePolicyName, "unknown policy"),
		StoragePolicySubtitle: row.StorageType,
		StorageType:           row.StorageType,
		Encrypted:             row.Encrypted,
		Locked:                row.Locked,
		LockedReason:          row.LockedReason,
		LockedAt:              row.LockedAt,
		LockedBy:              row.LockedBy,
		LinkedFiles:           []AdminBlobLinkedFilePayload{},
		LinkedVersions:        []AdminBlobLinkedVersionPayload{},
		LinkedCaches:          []AdminBlobLinkedCachePayload{},
		ReferenceSources:      []AdminBlobReferenceSourcePayload{},
		HealthStatus:          "ok",
	}
	if row.CacheRefCount > 0 {
		item.LinkedCaches = append(item.LinkedCaches, AdminBlobLinkedCachePayload{Source: "storage_path", Path: row.StoragePath})
		item.ReferenceSources = append(item.ReferenceSources, AdminBlobReferenceSourcePayload{Type: "cache", ID: row.StoragePath, Name: "thumbnail/cache path"})
	}
	if row.TaskRefCount > 0 {
		item.ReferenceSources = append(item.ReferenceSources, AdminBlobReferenceSourcePayload{Type: "queue_job", ID: row.StoragePath, Name: "queue task reference"})
	}
	if includeDetail {
		if err := s.populateReferences(item); err != nil {
			return nil, err
		}
	}
	if item.Encrypted && s.masterKeys != nil {
		item.EncryptionStatus = s.masterKeys.Status()
	}
	item.MissingOnStorage = strings.TrimSpace(row.StoragePath) == "" || s.storage == nil || !s.storage.Exists(row.StoragePath)
	item.DeleteBlockedReasons = s.deleteBlockedReasons(item)
	item.CanDelete = len(item.DeleteBlockedReasons) == 0
	if item.MissingOnStorage {
		item.HealthStatus = "missing"
	} else if row.RefCount != row.RealReferenceCount {
		item.HealthStatus = "ref_mismatch"
	}
	return item, nil
}

func (s *adminBlobService) populateReferences(item *AdminBlobPayload) error {
	var files []struct {
		ID        uint
		FileName  string
		FileSize  int64
		UserID    uint
		Username  string
		CreatedAt time.Time
	}
	if err := s.db.Table("user_files").
		Select("user_files.id, user_files.file_name, user_files.file_size, user_files.user_id, users.username, user_files.created_at").
		Joins("LEFT JOIN users ON users.id = user_files.user_id").
		Where("user_files.physical_file_id = ? AND user_files.deleted_at IS NULL", item.ID).
		Order("user_files.created_at asc").
		Scan(&files).Error; err != nil {
		return err
	}
	for _, file := range files {
		item.LinkedFiles = append(item.LinkedFiles, AdminBlobLinkedFilePayload{ID: file.ID, Name: file.FileName, Extension: extensionBadge(file.FileName), SizeBytes: file.FileSize, OwnerID: file.UserID, OwnerName: file.Username, CreatedAt: file.CreatedAt})
		item.ReferenceSources = append(item.ReferenceSources, AdminBlobReferenceSourcePayload{Type: "user_file", ID: strconv.FormatUint(uint64(file.ID), 10), Name: file.FileName})
	}

	var versions []struct {
		ID            uint
		FileID        uint
		FileName      string
		VersionNumber int
		FileSize      int64
		UploaderID    uint
		Username      string
		CreatedAt     time.Time
	}
	if err := s.db.Table("file_versions").
		Select("file_versions.id, file_versions.file_id, user_files.file_name, file_versions.version_number, file_versions.file_size, file_versions.uploader_id, users.username, file_versions.created_at").
		Joins("LEFT JOIN user_files ON user_files.id = file_versions.file_id").
		Joins("LEFT JOIN users ON users.id = file_versions.uploader_id").
		Where("file_versions.physical_file_id = ? AND file_versions.deleted_at IS NULL", item.ID).
		Order("file_versions.created_at asc").
		Scan(&versions).Error; err != nil {
		return err
	}
	for _, version := range versions {
		item.LinkedVersions = append(item.LinkedVersions, AdminBlobLinkedVersionPayload{ID: version.ID, FileID: version.FileID, FileName: version.FileName, VersionNumber: version.VersionNumber, SizeBytes: version.FileSize, OwnerID: version.UploaderID, OwnerName: version.Username, CreatedAt: version.CreatedAt})
		item.ReferenceSources = append(item.ReferenceSources, AdminBlobReferenceSourcePayload{Type: "file_version", ID: strconv.FormatUint(uint64(version.ID), 10), Name: fmt.Sprintf("%s v%d", version.FileName, version.VersionNumber)})
	}
	return nil
}

func (s *adminBlobService) deleteOne(id uint) (AdminBlobBatchDeleteResultItem, error) {
	item, err := s.Get(id)
	if err != nil {
		return AdminBlobBatchDeleteResultItem{ID: id}, err
	}
	if reasons := s.deleteBlockedReasons(item); len(reasons) > 0 {
		return AdminBlobBatchDeleteResultItem{ID: id, Reason: strings.Join(reasons, "; ")}, nil
	}
	if strings.TrimSpace(item.Source) != "" && s.storage != nil && s.storage.Exists(item.Source) {
		if err := s.storage.Delete(item.Source); err != nil {
			return AdminBlobBatchDeleteResultItem{ID: id}, err
		}
	}
	if err := s.db.Unscoped().Delete(&model.PhysicalFile{}, id).Error; err != nil {
		return AdminBlobBatchDeleteResultItem{ID: id}, err
	}
	return AdminBlobBatchDeleteResultItem{ID: id}, nil
}

func (s *adminBlobService) deleteBlockedReasons(item *AdminBlobPayload) []string {
	var reasons []string
	if item == nil {
		return []string{"blob not found"}
	}
	if item.Locked {
		reasons = append(reasons, "blob is locked")
	}
	if item.Kind != "orphan" {
		reasons = append(reasons, "only orphan blobs can be deleted")
	}
	if item.ReferenceCount > 0 || len(item.LinkedFiles) > 0 || len(item.LinkedVersions) > 0 {
		reasons = append(reasons, "blob is still referenced")
	}
	return reasons
}

func (s *adminBlobService) runScan(task *model.BlobScanTask) (*AdminBlobScanTaskPayload, error) {
	var physicalFiles []model.PhysicalFile
	if err := s.db.Find(&physicalFiles).Error; err != nil {
		return nil, err
	}

	storagePaths, walkErr := s.walkStorageFiles()
	dbPaths := map[string]struct{}{}
	hashCounts := map[string]int{}
	issues := []AdminBlobScanIssuePayload{}
	var orphanCount, missingCount, mismatchCount, duplicateRows, zeroSize, invalidPath int64

	for _, physical := range physicalFiles {
		dbPaths[filepath.ToSlash(strings.TrimSpace(physical.StoragePath))] = struct{}{}
		if strings.TrimSpace(physical.FileHash) != "" {
			hashCounts[physical.FileHash]++
		}
		realRefs, err := s.realReferenceCount(physical.ID, physical.StoragePath)
		if err != nil {
			return nil, err
		}
		if realRefs == 0 {
			orphanCount++
			issues = append(issues, AdminBlobScanIssuePayload{BlobID: physical.ID, Path: physical.StoragePath, Type: "orphan", Reason: "no DB or cache reference"})
		}
		if strings.TrimSpace(physical.StoragePath) == "" || s.storage == nil || !s.storage.Exists(physical.StoragePath) {
			missingCount++
			issues = append(issues, AdminBlobScanIssuePayload{BlobID: physical.ID, Path: physical.StoragePath, Type: "missing_on_storage", Reason: "DB record exists but storage file is absent"})
		}
		if physical.RefCount != realRefs {
			mismatchCount++
			issues = append(issues, AdminBlobScanIssuePayload{BlobID: physical.ID, Path: physical.StoragePath, Type: "ref_count_mismatch", Reason: fmt.Sprintf("stored=%d real=%d", physical.RefCount, realRefs)})
		}
		if physical.FileSize == 0 {
			zeroSize++
			issues = append(issues, AdminBlobScanIssuePayload{BlobID: physical.ID, Path: physical.StoragePath, Type: "zero_size", Reason: "physical_files.file_size is zero"})
		}
		if invalidStoragePath(physical.StoragePath) {
			invalidPath++
			issues = append(issues, AdminBlobScanIssuePayload{BlobID: physical.ID, Path: physical.StoragePath, Type: "invalid_path", Reason: "storage path is empty, absolute, or escapes storage root"})
		}
	}
	for hash, count := range hashCounts {
		if count > 1 {
			duplicateRows += int64(count)
			issues = append(issues, AdminBlobScanIssuePayload{Hash: hash, Type: "duplicate_hash", Reason: fmt.Sprintf("%d physical file rows share this hash", count)})
		}
	}
	var extraStorageFiles int64
	for path := range storagePaths {
		if _, ok := dbPaths[path]; !ok {
			extraStorageFiles++
			issues = append(issues, AdminBlobScanIssuePayload{Path: path, Type: "orphan_storage_file", Reason: "storage file exists without physical_files row"})
		}
	}
	if walkErr != nil {
		issues = append(issues, AdminBlobScanIssuePayload{Type: "storage_walk_error", Reason: walkErr.Error()})
	}

	resultJSON, _ := json.Marshal(issues)
	finished := time.Now()
	updates := map[string]interface{}{
		"status":                 model.BlobScanTaskStatusCompleted,
		"progress":               100,
		"total_physical_files":   int64(len(physicalFiles)),
		"scanned_physical_files": int64(len(physicalFiles)),
		"storage_file_count":     int64(len(storagePaths)),
		"orphan_count":           orphanCount,
		"missing_on_storage":     missingCount,
		"ref_count_mismatch":     mismatchCount,
		"duplicate_hash":         duplicateRows,
		"zero_size":              zeroSize,
		"invalid_path":           invalidPath,
		"extra_storage_files":    extraStorageFiles,
		"finished_at":            &finished,
		"result_json":            string(resultJSON),
	}
	if err := s.db.Model(task).Updates(updates).Error; err != nil {
		return nil, err
	}
	for key, value := range updates {
		_ = key
		_ = value
	}
	return s.LatestScan()
}

func (s *adminBlobService) realReferenceCount(id uint, storagePath string) (int, error) {
	var userCount, versionCount, taskCount int64
	if err := s.db.Model(&model.UserFile{}).Where("physical_file_id = ?", id).Count(&userCount).Error; err != nil {
		return 0, err
	}
	if err := s.db.Model(&model.FileVersion{}).Where("physical_file_id = ?", id).Count(&versionCount).Error; err != nil {
		return 0, err
	}
	taskQuery := s.db.Model(&model.QueueJob{}).Where("resource_id = ?", strconv.FormatUint(uint64(id), 10))
	if strings.TrimSpace(storagePath) != "" {
		taskQuery = taskQuery.Or("payload LIKE ?", "%"+storagePath+"%")
	}
	if err := taskQuery.Count(&taskCount).Error; err != nil {
		return 0, err
	}
	cacheCount := int64(0)
	lower := strings.ToLower(storagePath)
	if strings.Contains(lower, "thumb") || strings.Contains(lower, "thumbnail") || strings.Contains(lower, "cache") {
		cacheCount = 1
	}
	if taskCount > 0 {
		taskCount = 1
	}
	return int(userCount + versionCount + cacheCount + taskCount), nil
}

func (s *adminBlobService) walkStorageFiles() (map[string]struct{}, error) {
	result := map[string]struct{}{}
	provider, ok := s.storage.(adminBlobStorageRoot)
	if !ok || strings.TrimSpace(provider.BasePath()) == "" {
		return result, nil
	}
	root := provider.BasePath()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		result[filepath.ToSlash(rel)] = struct{}{}
		return nil
	})
	return result, err
}

func blobScanTaskPayload(task *model.BlobScanTask) *AdminBlobScanTaskPayload {
	var issues []AdminBlobScanIssuePayload
	_ = json.Unmarshal([]byte(task.ResultJSON), &issues)
	return &AdminBlobScanTaskPayload{
		ID:                   task.ID,
		Status:               task.Status,
		Progress:             task.Progress,
		TotalPhysicalFiles:   task.TotalPhysicalFiles,
		ScannedPhysicalFiles: task.ScannedPhysicalFiles,
		StorageFileCount:     task.StorageFileCount,
		OrphanCount:          task.OrphanCount,
		MissingOnStorage:     task.MissingOnStorage,
		RefCountMismatch:     task.RefCountMismatch,
		DuplicateHash:        task.DuplicateHash,
		ZeroSize:             task.ZeroSize,
		InvalidPath:          task.InvalidPath,
		ExtraStorageFiles:    task.ExtraStorageFiles,
		StartedAt:            task.StartedAt,
		FinishedAt:           task.FinishedAt,
		LastError:            task.LastError,
		Issues:               issues,
	}
}

func normalizeAdminBlobListQuery(query *AdminBlobListQuery) *AdminBlobListQuery {
	q := &AdminBlobListQuery{Page: 1, PageSize: 20, Kind: "all", SortBy: "id", SortOrder: "desc"}
	if query != nil {
		*q = *query
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 200 {
		q.PageSize = 200
	}
	q.Kind = strings.ToLower(strings.TrimSpace(q.Kind))
	if q.Kind == "" {
		q.Kind = "all"
	}
	q.Keyword = strings.TrimSpace(q.Keyword)
	q.SortBy = strings.ToLower(strings.TrimSpace(q.SortBy))
	q.SortOrder = strings.ToLower(strings.TrimSpace(q.SortOrder))
	if q.SortOrder != "asc" {
		q.SortOrder = "desc"
	}
	return q
}

func adminBlobOrderClause(sortBy, sortOrder string) string {
	columns := map[string]string{
		"id":              "id",
		"size":            "file_size",
		"reference_count": "real_reference_count",
		"created_at":      "created_at",
	}
	column := columns[sortBy]
	if column == "" {
		column = "id"
	}
	if sortOrder != "asc" {
		sortOrder = "desc"
	}
	return "ORDER BY " + column + " " + sortOrder
}

func (s *adminBlobService) maxBatchActionSize() int {
	var setting model.FileSystemSetting
	if err := s.db.Order("id asc").First(&setting).Error; err == nil && setting.MaxBatchActionSize > 0 {
		return setting.MaxBatchActionSize
	}
	return defaultMaxBatchActionSize
}

const defaultMaxBatchActionSize = 3000

func uniqueAdminBlobUintIDs(ids []uint) []uint {
	seen := map[uint]struct{}{}
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func adminBlobKindLabel(kind string) string {
	switch kind {
	case "thumbnail":
		return "Thumbnail Cache"
	case "version":
		return "Version Blob"
	case "live-photo":
		return "Live Photo"
	case "orphan":
		return "Orphan Blob"
	default:
		return "File Blob"
	}
}

func extensionBadge(fileName string) string {
	ext := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(filepath.Ext(fileName))), ".")
	if ext == "" {
		return "B"
	}
	if len(ext) > 3 {
		return ext[:3]
	}
	return ext
}

func creatorBadge(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "S"
	}
	for _, r := range name {
		return strings.ToUpper(string(r))
	}
	return "S"
}

func fallbackString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func invalidStoragePath(path string) bool {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || filepath.IsAbs(trimmed) {
		return true
	}
	cleaned := filepath.Clean(trimmed)
	return cleaned == "." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) || cleaned == ".."
}

func (s *adminBlobService) downloadFileName(detail *AdminBlobPayload) string {
	if detail != nil {
		if len(detail.LinkedFiles) > 0 && strings.TrimSpace(detail.LinkedFiles[0].Name) != "" {
			return detail.LinkedFiles[0].Name
		}
		if strings.TrimSpace(detail.Hash) != "" {
			return fmt.Sprintf("blob-%d-%s.bin", detail.ID, detail.Hash)
		}
		return fmt.Sprintf("blob-%d.bin", detail.ID)
	}
	return "blob.bin"
}

func sortScanIssues(issues []AdminBlobScanIssuePayload) {
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Type == issues[j].Type {
			return issues[i].BlobID < issues[j].BlobID
		}
		return issues[i].Type < issues[j].Type
	})
}
