package repository

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetPreference(userID uint) (*model.UserPreference, error)
	ListAll() ([]model.User, error)
	ListAdminUsers(query *AdminUserListQuery) ([]AdminUserListRow, int64, error)
	GetDeletePreview(id uint) (*UserDeletePreview, error)
	ListByGroupID(groupID uint) ([]model.User, error)
	Update(user *model.User) error
	SavePreference(preference *model.UserPreference) error
	Delete(id uint) error
	UpdateUsedSize(id uint, size int64) error
	UpdateUsedCapacity(id uint, delta int64) error
}

type userRepository struct {
	db *gorm.DB
}

type AdminUserListQuery struct {
	Keyword       string
	Role          string
	Status        string
	UserGroupID   uint
	UserGroupName string
	Page          int
	PageSize      int
}

type AdminUserListRow struct {
	model.User
	UserGroupName string `gorm:"column:user_group_name"`
}

type UserDeletePreview struct {
	UserID                   uint  `json:"user_id"`
	FileCount                int64 `json:"file_count"`
	ShareCount               int64 `json:"share_count"`
	RecycleBinCount          int64 `json:"recycle_bin_count"`
	OfflineDownloadTaskCount int64 `json:"offline_download_task_count"`
	MultipartUploadCount     int64 `json:"multipart_upload_count"`
	DavAccountCount          int64 `json:"dav_account_count"`
	OAuthCredentialCount     int64 `json:"oauth_credential_count"`
	CollaborationOwnedCount  int64 `json:"collaboration_owned_count"`
	CollaborationSharedCount int64 `json:"collaboration_shared_count"`
	TrafficEventCount        int64 `json:"traffic_event_count"`
	UserPreferenceCount      int64 `json:"user_preference_count"`
	FileCustomPropertyCount  int64 `json:"file_custom_property_count"`
	StoragePolicyHitLogCount int64 `json:"storage_policy_hit_log_count"`
	UsedSize                 int64 `json:"used_size"`
	HasBlockingAssets        bool  `json:"has_blocking_assets"`
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	capacity := user.Capacity
	if err := r.db.Select("*").Create(user).Error; err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	if capacity == 0 {
		if err := r.db.Model(&model.User{}).Where("id = ?", user.ID).Update("capacity", int64(0)).Error; err != nil {
			return fmt.Errorf("set unlimited user capacity failed: %w", err)
		}
		user.Capacity = 0
	}
	return nil
}

func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetPreference(userID uint) (*model.UserPreference, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	var preference model.UserPreference
	if err := r.db.Where("user_id = ?", userID).First(&preference).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user preference not found")
		}
		return nil, fmt.Errorf("query user preference failed: %w", err)
	}
	return &preference, nil
}

func (r *userRepository) ListAll() ([]model.User, error) {
	var users []model.User
	if err := r.db.Order("id desc").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("query users failed: %w", err)
	}
	return users, nil
}

func (r *userRepository) ListAdminUsers(query *AdminUserListQuery) ([]AdminUserListRow, int64, error) {
	if query == nil {
		query = &AdminUserListQuery{}
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	base := r.db.Model(&model.User{}).
		Select("users.*, COALESCE(user_groups.name, '') AS user_group_name").
		Joins("LEFT JOIN user_groups ON user_groups.id = users.user_group_id")

	if query.Keyword != "" {
		like := "%" + strings.ToLower(query.Keyword) + "%"
		base = base.Where(
			"LOWER(users.username) LIKE ? OR LOWER(users.email) LIKE ? OR LOWER(COALESCE(user_groups.name, '')) LIKE ?",
			like,
			like,
			like,
		)
	}
	if query.Role == "admin" || query.Role == "user" {
		base = base.Where("users.role = ?", query.Role)
	}
	if query.Status == "enabled" {
		base = base.Where("users.enabled = ?", true)
	} else if query.Status == "disabled" {
		base = base.Where("users.enabled = ?", false)
	}
	if query.UserGroupID > 0 {
		base = base.Where("users.user_group_id = ?", query.UserGroupID)
	}
	if query.UserGroupName != "" {
		base = base.Where("user_groups.name = ?", query.UserGroupName)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count admin users failed: %w", err)
	}

	rows := make([]AdminUserListRow, 0, pageSize)
	if err := base.
		Order("users.id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("query admin users failed: %w", err)
	}
	return rows, total, nil
}

func (r *userRepository) GetDeletePreview(id uint) (*UserDeletePreview, error) {
	if id == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	var user model.User
	if err := r.db.Select("id", "used_size").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user failed: %w", err)
	}

	preview := &UserDeletePreview{UserID: user.ID, UsedSize: user.UsedSize}
	counts := []struct {
		table string
		where string
		dest  *int64
	}{
		{"user_files", "user_id = ?", &preview.FileCount},
		{"shares", "user_id = ?", &preview.ShareCount},
		{"recycle_bin", "user_id = ?", &preview.RecycleBinCount},
		{"offline_download_tasks", "user_id = ?", &preview.OfflineDownloadTaskCount},
		{"multipart_uploads", "user_id = ?", &preview.MultipartUploadCount},
		{"dav_accounts", "user_id = ?", &preview.DavAccountCount},
		{"oauth_credentials", "user_id = ?", &preview.OAuthCredentialCount},
		{"collaborations", "owner_id = ?", &preview.CollaborationOwnedCount},
		{"collaborations", "collaborator_id = ?", &preview.CollaborationSharedCount},
		{"traffic_events", "user_id = ?", &preview.TrafficEventCount},
		{"user_preferences", "user_id = ?", &preview.UserPreferenceCount},
		{"file_custom_property_values", "user_id = ?", &preview.FileCustomPropertyCount},
		{"storage_policy_hit_logs", "user_id = ?", &preview.StoragePolicyHitLogCount},
	}
	for _, item := range counts {
		if !r.db.Migrator().HasTable(item.table) {
			continue
		}
		query := r.db.Table(item.table).Where(item.where, id)
		if r.db.Migrator().HasColumn(item.table, "deleted_at") {
			query = query.Where("deleted_at IS NULL")
		}
		if err := query.Count(item.dest).Error; err != nil {
			return nil, fmt.Errorf("count %s failed: %w", item.table, err)
		}
	}

	preview.HasBlockingAssets = preview.FileCount > 0 ||
		preview.ShareCount > 0 ||
		preview.RecycleBinCount > 0 ||
		preview.OfflineDownloadTaskCount > 0 ||
		preview.MultipartUploadCount > 0 ||
		preview.DavAccountCount > 0 ||
		preview.OAuthCredentialCount > 0 ||
		preview.CollaborationOwnedCount > 0 ||
		preview.CollaborationSharedCount > 0 ||
		preview.FileCustomPropertyCount > 0 ||
		preview.StoragePolicyHitLogCount > 0 ||
		preview.UsedSize > 0

	return preview, nil
}

func (r *userRepository) ListByGroupID(groupID uint) ([]model.User, error) {
	var users []model.User
	if err := r.db.Where("user_group_id = ?", groupID).Order("id asc").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("query user group members failed: %w", err)
	}
	return users, nil
}

func (r *userRepository) Update(user *model.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}
	if user.ID == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}
	if err := r.db.Select("*").Save(user).Error; err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	return nil
}

func (r *userRepository) SavePreference(preference *model.UserPreference) error {
	if preference == nil {
		return fmt.Errorf("user preference cannot be nil")
	}
	if preference.UserID == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}

	var existing model.UserPreference
	err := r.db.Where("user_id = ?", preference.UserID).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		if err := r.db.Create(preference).Error; err != nil {
			return fmt.Errorf("save user preference failed: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("query user preference failed: %w", err)
	}

	updates := map[string]interface{}{
		"language":           preference.Language,
		"timezone":           preference.Timezone,
		"mode":               preference.Mode,
		"theme":              preference.Theme,
		"keep_versions":      preference.KeepVersions,
		"version_extensions": preference.VersionExtensions,
		"max_versions":       preference.MaxVersions,
		"view_sync":          preference.ViewSync,
		"expand_tree":        preference.ExpandTree,
		"folder_action":      preference.FolderAction,
		"home_visibility":    preference.HomeVisibility,
	}
	if err := r.db.Model(&existing).Updates(updates).Error; err != nil {
		return fmt.Errorf("save user preference failed: %w", err)
	}

	existing.Language = preference.Language
	existing.Timezone = preference.Timezone
	existing.Mode = preference.Mode
	existing.Theme = preference.Theme
	existing.KeepVersions = preference.KeepVersions
	existing.VersionExtensions = preference.VersionExtensions
	existing.MaxVersions = preference.MaxVersions
	existing.ViewSync = preference.ViewSync
	existing.ExpandTree = preference.ExpandTree
	existing.FolderAction = preference.FolderAction
	existing.HomeVisibility = preference.HomeVisibility
	*preference = existing
	return nil
}

func (r *userRepository) Delete(id uint) error {
	if id == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}
	if err := r.db.Delete(&model.User{}, id).Error; err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateUsedSize(id uint, size int64) error {
	if id == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}
	if err := r.db.Model(&model.User{}).Where("id = ?", id).Update("used_size", size).Error; err != nil {
		return fmt.Errorf("update user used size failed: %w", err)
	}
	return nil
}

func (r *userRepository) UpdateUsedCapacity(id uint, delta int64) error {
	if id == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}
	if delta <= 0 {
		return nil
	}
	result := r.db.Model(&model.User{}).
		Where("id = ? AND (capacity = 0 OR used_size <= capacity - ?)", id, delta).
		UpdateColumn("used_size", gorm.Expr("used_size + ?", delta))
	if result.Error != nil {
		return fmt.Errorf("update used capacity failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("quota exceeded: capacity not enough")
	}
	return nil
}
