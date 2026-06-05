package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"

	"gorm.io/gorm"
)

type StoragePolicyValidationError struct {
	Message string
}

func (e *StoragePolicyValidationError) Error() string {
	return e.Message
}

func IsStoragePolicyValidationError(err error) bool {
	var validationErr *StoragePolicyValidationError
	return errors.As(err, &validationErr)
}

type storagePolicyRuntime struct {
	db     *gorm.DB
	userID uint
}

type StoragePolicyUploadTuning struct {
	ChunkSizeBytes     int
	ParallelChunkCount int
	HasPolicy          bool
}

type StoragePolicyBlobPathContext struct {
	FileName      string
	FileHash      string
	ParentID      *uint
	UserFileRepo  repository.UserFileRepository
	FallbackPath  string
	StorageExists func(string) bool
}

func newStoragePolicyRuntime(db *gorm.DB, userID uint) storagePolicyRuntime {
	return storagePolicyRuntime{db: db, userID: userID}
}

func (r storagePolicyRuntime) ValidateUpload(fileName string, fileSize int64) error {
	policy, err := r.loadPolicy()
	if err != nil || policy == nil {
		return err
	}

	if err := validateStoragePolicyFileSize(policy, fileSize); err != nil {
		return err
	}
	if err := validateStoragePolicyFileName(policy, fileName); err != nil {
		return err
	}
	return nil
}

func (r storagePolicyRuntime) ValidateFileName(fileName string) error {
	policy, err := r.loadPolicy()
	if err != nil || policy == nil {
		return err
	}
	return validateStoragePolicyFileName(policy, fileName)
}

func (r storagePolicyRuntime) UploadTuning() (*StoragePolicyUploadTuning, error) {
	policy, err := r.loadPolicy()
	if err != nil || policy == nil {
		return &StoragePolicyUploadTuning{}, err
	}

	return &StoragePolicyUploadTuning{
		ChunkSizeBytes:     int(storagePolicySizeBytes(policy.ChunkSize, policy.ChunkSizeUnit)),
		ParallelChunkCount: policy.ParallelChunkCount,
		HasPolicy:          true,
	}, nil
}

func (r storagePolicyRuntime) PreAllocateEnabled() (bool, error) {
	policy, err := r.loadPolicy()
	if err != nil || policy == nil {
		return false, err
	}
	return policy.PreAllocate, nil
}

func (r storagePolicyRuntime) RenderBlobStoragePath(ctx StoragePolicyBlobPathContext) (string, error) {
	fallback := cleanRelativeStoragePath(ctx.FallbackPath)
	if fallback == "" {
		fallback = cleanRelativeStoragePath(path.Join("files", strings.TrimSpace(ctx.FileHash)))
	}

	policy, err := r.loadPolicy()
	if err != nil || policy == nil {
		return fallback, err
	}

	dirPattern := strings.TrimSpace(policy.BlobPath)
	namePattern := strings.TrimSpace(policy.BlobNamePattern)
	if dirPattern == "" || namePattern == "" {
		return fallback, nil
	}

	folderPath, err := r.resolveUploadFolderPath(ctx.ParentID, ctx.UserFileRepo)
	if err != nil {
		return "", err
	}

	values := map[string]string{
		"uid":        fmt.Sprintf("%d", r.userID),
		"path":       folderPath,
		"originname": sanitizeStorageFileName(ctx.FileName),
		"randomkey8": randomStorageKey(8),
		"hash":       strings.TrimSpace(ctx.FileHash),
	}

	dir := cleanRelativeStoragePath(applyStoragePolicyTemplate(dirPattern, values))
	name := sanitizeStorageFileName(applyStoragePolicyTemplate(namePattern, values))
	if name == "" {
		name = strings.TrimSpace(ctx.FileHash)
	}
	candidate := cleanRelativeStoragePath(path.Join(dir, name))
	if candidate == "" {
		return fallback, nil
	}
	if len(candidate) > 480 {
		candidate = fallback
	}

	return ensureUniqueStoragePath(candidate, ctx.StorageExists), nil
}

func (r storagePolicyRuntime) CDNDownloadURL(storagePath string) (string, error) {
	policy, err := r.loadPolicy()
	if err != nil || policy == nil {
		return "", err
	}
	if !policy.EnableCDN {
		return "", nil
	}

	base := strings.TrimSpace(policy.DownloadCDN)
	if base == "" {
		return "", nil
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", nil
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", nil
	}

	relativePath := cleanRelativeStoragePath(storagePath)
	if relativePath == "" {
		return "", nil
	}

	return strings.TrimRight(base, "/") + "/" + escapeStoragePathForURL(relativePath), nil
}

func (r storagePolicyRuntime) resolveUploadFolderPath(parentID *uint, repo repository.UserFileRepository) (string, error) {
	if parentID == nil || *parentID == 0 || repo == nil {
		return "", nil
	}

	folders, err := repo.GetFolderPath(r.userID, *parentID)
	if err != nil {
		return "", fmt.Errorf("resolve upload folder path failed: %w", err)
	}

	segments := make([]string, 0, len(folders))
	for _, folder := range folders {
		if folder == nil {
			continue
		}
		segment := sanitizeStoragePathSegment(folder.FileName)
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	return strings.Join(segments, "/"), nil
}

func (r storagePolicyRuntime) loadPolicy() (*model.StoragePolicy, error) {
	if r.db == nil || r.userID == 0 {
		return nil, nil
	}

	var row struct {
		StoragePolicyID uint
	}
	if err := r.db.Table("users").
		Select("COALESCE(user_groups.storage_policy_id, 0) AS storage_policy_id").
		Joins("LEFT JOIN user_groups ON user_groups.id = users.user_group_id").
		Where("users.id = ?", r.userID).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("load user storage policy failed: %w", err)
	}
	if row.StoragePolicyID == 0 {
		return nil, nil
	}

	var policy model.StoragePolicy
	if err := r.db.First(&policy, row.StoragePolicyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("storage policy %d bound to user group does not exist", row.StoragePolicyID)
		}
		return nil, fmt.Errorf("load storage policy failed: %w", err)
	}
	return &policy, nil
}

func validateStoragePolicyFileSize(policy *model.StoragePolicy, fileSize int64) error {
	limit := storagePolicySizeBytes(policy.MaxFileSize, policy.MaxFileSizeUnit)
	if limit <= 0 || fileSize <= limit {
		return nil
	}
	return &StoragePolicyValidationError{Message: fmt.Sprintf("文件大小超过存储策略限制：最大允许 %d%s", policy.MaxFileSize, strings.ToUpper(policy.MaxFileSizeUnit))}
}

func validateStoragePolicyFileName(policy *model.StoragePolicy, fileName string) error {
	name := strings.TrimSpace(fileName)
	if name == "" {
		return &StoragePolicyValidationError{Message: "文件名不能为空"}
	}

	if err := validateStoragePolicyExtension(policy, name); err != nil {
		return err
	}
	if err := validateStoragePolicyNameRegex(policy, name); err != nil {
		return err
	}
	return nil
}

func validateStoragePolicyExtension(policy *model.StoragePolicy, fileName string) error {
	rules := parseStoragePolicyExtensions(policy.Extensions)
	if len(rules) == 0 {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	_, matched := rules[ext]
	mode := strings.ToLower(strings.TrimSpace(policy.ExtensionMode))
	switch mode {
	case "deny":
		if matched {
			return &StoragePolicyValidationError{Message: fmt.Sprintf("存储策略禁止上传 %s 类型文件", ext)}
		}
	default:
		if !matched {
			return &StoragePolicyValidationError{Message: fmt.Sprintf("文件扩展名不在存储策略允许范围内：%s", ext)}
		}
	}
	return nil
}

func validateStoragePolicyNameRegex(policy *model.StoragePolicy, fileName string) error {
	pattern := strings.TrimSpace(policy.NameRegex)
	if pattern == "" {
		return nil
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return &StoragePolicyValidationError{Message: fmt.Sprintf("存储策略文件名正则无效：%s", err.Error())}
	}

	matched := compiled.MatchString(fileName)
	mode := strings.ToLower(strings.TrimSpace(policy.NameRuleMode))
	switch mode {
	case "deny":
		if matched {
			return &StoragePolicyValidationError{Message: "文件名命中存储策略拒绝规则"}
		}
	default:
		if !matched {
			return &StoragePolicyValidationError{Message: "文件名不符合存储策略允许规则"}
		}
	}
	return nil
}

func parseStoragePolicyExtensions(raw string) map[string]struct{} {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	result := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		ext := strings.ToLower(strings.TrimSpace(part))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		result[ext] = struct{}{}
	}
	return result
}

func storagePolicySizeBytes(value int, unit string) int64 {
	if value <= 0 {
		return 0
	}
	size := int64(value)
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "KB":
		return size * 1024
	case "GB":
		return size * 1024 * 1024 * 1024
	default:
		return size * 1024 * 1024
	}
}

func applyStoragePolicyTemplate(pattern string, values map[string]string) string {
	result := pattern
	for key, value := range values {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}
	return result
}

func cleanRelativeStoragePath(value string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	normalized = strings.TrimPrefix(filepath.ToSlash(normalized), filepath.VolumeName(normalized))
	normalized = strings.TrimLeft(normalized, "/")
	if normalized == "" {
		return ""
	}
	cleaned := path.Clean("/" + normalized)
	cleaned = strings.TrimLeft(cleaned, "/")
	if cleaned == "." || cleaned == ".." {
		return ""
	}
	return cleaned
}

func sanitizeStorageFileName(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	value = path.Base(value)
	if value == "." || value == "/" {
		return ""
	}
	return sanitizeStoragePathSegment(value)
}

func sanitizeStoragePathSegment(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\x00", "")
	value = strings.ReplaceAll(value, "/", "_")
	value = strings.ReplaceAll(value, "\\", "_")
	value = strings.Trim(value, ". ")
	if value == "" || value == "." || value == ".." {
		return ""
	}
	return value
}

func ensureUniqueStoragePath(candidate string, exists func(string) bool) string {
	if exists == nil || !exists(candidate) {
		return candidate
	}

	dir := path.Dir(candidate)
	if dir == "." {
		dir = ""
	}
	base := path.Base(candidate)
	ext := path.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" {
		name = "blob"
	}

	for i := 0; i < 5; i++ {
		next := cleanRelativeStoragePath(path.Join(dir, fmt.Sprintf("%s-%s%s", name, randomStorageKey(8), ext)))
		if next != "" && !exists(next) {
			return next
		}
	}

	return cleanRelativeStoragePath(path.Join(dir, fmt.Sprintf("%s-%s%s", name, randomStorageKey(16), ext)))
}

func randomStorageKey(length int) string {
	if length <= 0 {
		return ""
	}
	buf := make([]byte, (length+1)/2)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("0", length)
	}
	return hex.EncodeToString(buf)[:length]
}

func escapeStoragePathForURL(storagePath string) string {
	parts := strings.Split(cleanRelativeStoragePath(storagePath), "/")
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		escaped = append(escaped, url.PathEscape(part))
	}
	return strings.Join(escaped, "/")
}
