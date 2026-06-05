package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/cache"
	"xingyunpan-v2/pkg/crypto"
)

type AdminUserPayload struct {
	ID            uint      `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	Enabled       bool      `json:"enabled"`
	UserGroupID   uint      `json:"user_group_id"`
	UserGroupName string    `json:"user_group_name"`
	Capacity      int64     `json:"capacity"`
	UsedSize      int64     `json:"used_size"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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

type AdminUserListResult struct {
	Items    []AdminUserPayload `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type AdminUserUpsertPayload struct {
	Username            string `json:"username"`
	Email               string `json:"email"`
	Password            string `json:"password"`
	Role                string `json:"role"`
	Enabled             bool   `json:"enabled"`
	UserGroupID         uint   `json:"user_group_id"`
	Capacity            int64  `json:"capacity"`
	FollowGroupCapacity bool   `json:"follow_group_capacity"`
}

type AdminUserResetPasswordPayload struct {
	Password string `json:"password"`
}

type AdminUserBatchDeletePayload struct {
	IDs []uint `json:"ids"`
}

type AdminUserDeletePreviewPayload struct {
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

type AdminUserDeleteBlockedError struct {
	UserID  uint
	Preview *AdminUserDeletePreviewPayload
}

func (e *AdminUserDeleteBlockedError) Error() string {
	return fmt.Sprintf("user %d still has assets and cannot be deleted in safe mode", e.UserID)
}

type AdminUserBatchGroupPayload struct {
	IDs         []uint `json:"ids"`
	UserGroupID uint   `json:"user_group_id"`
}

type AdminUserBatchRolePayload struct {
	IDs  []uint `json:"ids"`
	Role string `json:"role"`
}

type AdminUserStatusPayload struct {
	Enabled bool `json:"enabled"`
}

type AdminUserBatchStatusPayload struct {
	IDs     []uint `json:"ids"`
	Enabled bool   `json:"enabled"`
}

type AdminUserService interface {
	List() ([]AdminUserPayload, error)
	ListUsers(query *AdminUserListQuery) (*AdminUserListResult, error)
	Create(payload *AdminUserUpsertPayload) (*AdminUserPayload, error)
	Update(id uint, payload *AdminUserUpsertPayload, operatorID uint) (*AdminUserPayload, error)
	GetDeletePreview(id uint) (*AdminUserDeletePreviewPayload, error)
	Delete(id uint, operatorID uint) error
	ResetPassword(id uint, payload *AdminUserResetPasswordPayload) error
	BatchDelete(payload *AdminUserBatchDeletePayload, operatorID uint) error
	BatchUpdateGroup(payload *AdminUserBatchGroupPayload, operatorID uint) ([]AdminUserPayload, error)
	BatchUpdateRole(payload *AdminUserBatchRolePayload, operatorID uint) ([]AdminUserPayload, error)
	UpdateStatus(id uint, payload *AdminUserStatusPayload, operatorID uint) (*AdminUserPayload, error)
	BatchUpdateStatus(payload *AdminUserBatchStatusPayload, operatorID uint) ([]AdminUserPayload, error)
}

type adminUserService struct {
	userRepo      repository.UserRepository
	userGroupRepo repository.UserGroupRepository
	cache         *cache.CacheService
}

func NewAdminUserService(
	userRepo repository.UserRepository,
	userGroupRepo repository.UserGroupRepository,
	cacheServices ...*cache.CacheService,
) AdminUserService {
	var cacheService *cache.CacheService
	if len(cacheServices) > 0 {
		cacheService = cacheServices[0]
	}
	return &adminUserService{
		userRepo:      userRepo,
		userGroupRepo: userGroupRepo,
		cache:         cacheService,
	}
}

func (s *adminUserService) List() ([]AdminUserPayload, error) {
	users, err := s.userRepo.ListAll()
	if err != nil {
		return nil, err
	}

	groups, err := s.userGroupRepo.List()
	if err != nil {
		return nil, err
	}

	groupNames := make(map[uint]string, len(groups))
	for _, group := range groups {
		groupNames[group.ID] = group.Name
	}

	result := make([]AdminUserPayload, 0, len(users))
	for _, user := range users {
		groupName := strings.TrimSpace(groupNames[user.UserGroupID])
		if groupName == "" {
			groupName = "未分组"
		}

		result = append(result, AdminUserPayload{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			Role:          user.Role,
			Enabled:       user.Enabled,
			UserGroupID:   user.UserGroupID,
			UserGroupName: groupName,
			Capacity:      user.Capacity,
			UsedSize:      user.UsedSize,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		})
	}

	return result, nil
}

func (s *adminUserService) ListUsers(query *AdminUserListQuery) (*AdminUserListResult, error) {
	normalized := normalizeAdminUserListQuery(query)
	rows, total, err := s.userRepo.ListAdminUsers(&repository.AdminUserListQuery{
		Keyword:       normalized.Keyword,
		Role:          normalized.Role,
		Status:        normalized.Status,
		UserGroupID:   normalized.UserGroupID,
		UserGroupName: normalized.UserGroupName,
		Page:          normalized.Page,
		PageSize:      normalized.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]AdminUserPayload, 0, len(rows))
	for _, row := range rows {
		groupName := strings.TrimSpace(row.UserGroupName)
		if groupName == "" {
			groupName = "未分组"
		}
		items = append(items, s.payloadFromUser(&row.User, groupName))
	}

	return &AdminUserListResult{
		Items:    items,
		Total:    total,
		Page:     normalized.Page,
		PageSize: normalized.PageSize,
	}, nil
}

func (s *adminUserService) Create(payload *AdminUserUpsertPayload) (*AdminUserPayload, error) {
	normalized, err := s.normalizePayload(0, payload, false)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := crypto.HashPassword(normalized.Password)
	if err != nil {
		return nil, err
	}

	item := &model.User{
		Username:    normalized.Username,
		Email:       normalized.Email,
		Password:    hashedPassword,
		Role:        normalized.Role,
		Enabled:     normalized.Enabled,
		UserGroupID: normalized.UserGroupID,
		Capacity:    normalized.Capacity,
		UsedSize:    0,
	}

	if err := s.userRepo.Create(item); err != nil {
		return nil, err
	}

	return s.buildPayload(item)
}

func (s *adminUserService) Update(id uint, payload *AdminUserUpsertPayload, operatorID uint) (*AdminUserPayload, error) {
	existing, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("user not found")
	}

	normalized, err := s.normalizePayload(id, payload, true)
	if err != nil {
		return nil, err
	}

	if operatorID == existing.ID && normalized.Role != "admin" {
		return nil, fmt.Errorf("cannot revoke your own admin role")
	}
	if operatorID == existing.ID && !normalized.Enabled {
		return nil, fmt.Errorf("cannot disable the currently logged-in administrator")
	}

	existing.Username = normalized.Username
	existing.Email = normalized.Email
	existing.Role = normalized.Role
	existing.Enabled = normalized.Enabled
	existing.UserGroupID = normalized.UserGroupID
	existing.Capacity = normalized.Capacity

	if normalized.Password != "" {
		hashedPassword, err := crypto.HashPassword(normalized.Password)
		if err != nil {
			return nil, err
		}
		existing.Password = hashedPassword
	}

	if err := s.userRepo.Update(existing); err != nil {
		return nil, err
	}
	s.invalidateUserCaches(existing.ID)

	return s.buildPayload(existing)
}

func (s *adminUserService) Delete(id uint, operatorID uint) error {
	if id == 0 {
		return fmt.Errorf("user id cannot be empty")
	}
	if id == operatorID {
		return fmt.Errorf("cannot delete the currently logged-in administrator")
	}
	if _, err := s.ensureUserDeletable(id); err != nil {
		return err
	}

	if err := s.userRepo.Delete(id); err != nil {
		return err
	}
	s.invalidateUserCaches(id)
	return nil
}

func (s *adminUserService) ResetPassword(id uint, payload *AdminUserResetPasswordPayload) error {
	if id == 0 {
		return fmt.Errorf("user id cannot be empty")
	}
	if payload == nil {
		return fmt.Errorf("reset password payload cannot be nil")
	}

	password := strings.TrimSpace(payload.Password)
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	existing, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("user not found")
	}

	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	existing.Password = hashedPassword

	if err := s.userRepo.Update(existing); err != nil {
		return err
	}
	s.invalidateUserCaches(existing.ID)
	return nil
}

func (s *adminUserService) BatchDelete(payload *AdminUserBatchDeletePayload, operatorID uint) error {
	if payload == nil {
		return fmt.Errorf("batch delete payload cannot be nil")
	}
	if len(payload.IDs) == 0 {
		return fmt.Errorf("please select at least one user")
	}

	seen := make(map[uint]struct{}, len(payload.IDs))
	targetIDs := make([]uint, 0, len(payload.IDs))
	for _, id := range payload.IDs {
		if id == 0 {
			continue
		}
		if id == operatorID {
			return fmt.Errorf("cannot batch delete the currently logged-in administrator")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if _, err := s.ensureUserDeletable(id); err != nil {
			return err
		}
		targetIDs = append(targetIDs, id)
	}

	for _, id := range targetIDs {
		if err := s.userRepo.Delete(id); err != nil {
			return err
		}
		s.invalidateUserCaches(id)
	}

	return nil
}

func (s *adminUserService) BatchUpdateGroup(payload *AdminUserBatchGroupPayload, operatorID uint) ([]AdminUserPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("batch group payload cannot be nil")
	}
	if len(payload.IDs) == 0 {
		return nil, fmt.Errorf("please select at least one user")
	}

	group, err := s.userGroupRepo.GetByID(payload.UserGroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("user group not found")
	}

	seen := make(map[uint]struct{}, len(payload.IDs))
	updated := make([]AdminUserPayload, 0, len(payload.IDs))
	for _, id := range payload.IDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		existing, err := s.userRepo.GetByID(id)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			continue
		}

		existing.UserGroupID = payload.UserGroupID
		existing.Capacity = group.MaxCapacity
		if err := s.userRepo.Update(existing); err != nil {
			return nil, err
		}
		s.invalidateUserCaches(existing.ID)

		item, err := s.buildPayload(existing)
		if err != nil {
			return nil, err
		}
		updated = append(updated, *item)
	}

	_ = operatorID
	return updated, nil
}

func (s *adminUserService) BatchUpdateRole(payload *AdminUserBatchRolePayload, operatorID uint) ([]AdminUserPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("batch role payload cannot be nil")
	}
	if len(payload.IDs) == 0 {
		return nil, fmt.Errorf("please select at least one user")
	}

	role := strings.TrimSpace(strings.ToLower(payload.Role))
	if role != "admin" && role != "user" {
		return nil, fmt.Errorf("invalid user role")
	}

	seen := make(map[uint]struct{}, len(payload.IDs))
	updated := make([]AdminUserPayload, 0, len(payload.IDs))
	for _, id := range payload.IDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		if id == operatorID && role != "admin" {
			return nil, fmt.Errorf("cannot revoke your own admin role")
		}

		existing, err := s.userRepo.GetByID(id)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			continue
		}

		existing.Role = role
		if err := s.userRepo.Update(existing); err != nil {
			return nil, err
		}
		s.invalidateUserCaches(existing.ID)

		item, err := s.buildPayload(existing)
		if err != nil {
			return nil, err
		}
		updated = append(updated, *item)
	}

	return updated, nil
}

func (s *adminUserService) UpdateStatus(id uint, payload *AdminUserStatusPayload, operatorID uint) (*AdminUserPayload, error) {
	if id == 0 {
		return nil, fmt.Errorf("user id cannot be empty")
	}
	if payload == nil {
		return nil, fmt.Errorf("status payload cannot be nil")
	}
	if id == operatorID && !payload.Enabled {
		return nil, fmt.Errorf("cannot disable the currently logged-in administrator")
	}

	existing, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("user not found")
	}

	existing.Enabled = payload.Enabled
	if err := s.userRepo.Update(existing); err != nil {
		return nil, err
	}
	s.invalidateUserCaches(existing.ID)

	return s.buildPayload(existing)
}

func (s *adminUserService) BatchUpdateStatus(payload *AdminUserBatchStatusPayload, operatorID uint) ([]AdminUserPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("batch status payload cannot be nil")
	}
	if len(payload.IDs) == 0 {
		return nil, fmt.Errorf("please select at least one user")
	}

	seen := make(map[uint]struct{}, len(payload.IDs))
	updated := make([]AdminUserPayload, 0, len(payload.IDs))
	for _, id := range payload.IDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		if id == operatorID && !payload.Enabled {
			return nil, fmt.Errorf("cannot disable the currently logged-in administrator")
		}

		existing, err := s.userRepo.GetByID(id)
		if err != nil {
			return nil, err
		}
		if existing == nil {
			continue
		}

		existing.Enabled = payload.Enabled
		if err := s.userRepo.Update(existing); err != nil {
			return nil, err
		}
		s.invalidateUserCaches(existing.ID)

		item, err := s.buildPayload(existing)
		if err != nil {
			return nil, err
		}
		updated = append(updated, *item)
	}

	return updated, nil
}

func (s *adminUserService) GetDeletePreview(id uint) (*AdminUserDeletePreviewPayload, error) {
	if id == 0 {
		return nil, fmt.Errorf("user id cannot be empty")
	}
	preview, err := s.userRepo.GetDeletePreview(id)
	if err != nil {
		return nil, err
	}
	return adminUserDeletePreviewFromRepo(preview), nil
}

func (s *adminUserService) ensureUserDeletable(id uint) (*AdminUserDeletePreviewPayload, error) {
	preview, err := s.GetDeletePreview(id)
	if err != nil {
		return nil, err
	}
	if preview.HasBlockingAssets {
		return preview, &AdminUserDeleteBlockedError{UserID: id, Preview: preview}
	}
	return preview, nil
}

func adminUserDeletePreviewFromRepo(preview *repository.UserDeletePreview) *AdminUserDeletePreviewPayload {
	if preview == nil {
		return nil
	}
	return &AdminUserDeletePreviewPayload{
		UserID:                   preview.UserID,
		FileCount:                preview.FileCount,
		ShareCount:               preview.ShareCount,
		RecycleBinCount:          preview.RecycleBinCount,
		OfflineDownloadTaskCount: preview.OfflineDownloadTaskCount,
		MultipartUploadCount:     preview.MultipartUploadCount,
		DavAccountCount:          preview.DavAccountCount,
		OAuthCredentialCount:     preview.OAuthCredentialCount,
		CollaborationOwnedCount:  preview.CollaborationOwnedCount,
		CollaborationSharedCount: preview.CollaborationSharedCount,
		TrafficEventCount:        preview.TrafficEventCount,
		UserPreferenceCount:      preview.UserPreferenceCount,
		FileCustomPropertyCount:  preview.FileCustomPropertyCount,
		StoragePolicyHitLogCount: preview.StoragePolicyHitLogCount,
		UsedSize:                 preview.UsedSize,
		HasBlockingAssets:        preview.HasBlockingAssets,
	}
}

func (s *adminUserService) invalidateUserCaches(userID uint) {
	if s.cache == nil || userID == 0 {
		return
	}
	_ = s.cache.InvalidateUserProfileAndSession(context.Background(), userID)
}

func normalizeAdminUserListQuery(query *AdminUserListQuery) AdminUserListQuery {
	if query == nil {
		query = &AdminUserListQuery{}
	}

	normalized := *query
	normalized.Keyword = strings.TrimSpace(normalized.Keyword)
	normalized.Role = strings.TrimSpace(strings.ToLower(normalized.Role))
	normalized.Status = strings.TrimSpace(strings.ToLower(normalized.Status))
	normalized.UserGroupName = strings.TrimSpace(normalized.UserGroupName)

	if normalized.Role != "admin" && normalized.Role != "user" {
		normalized.Role = "all"
	}
	if normalized.Status != "enabled" && normalized.Status != "disabled" {
		normalized.Status = "all"
	}
	if normalized.Page <= 0 {
		normalized.Page = 1
	}
	if normalized.PageSize <= 0 {
		normalized.PageSize = 10
	}
	if normalized.PageSize > 100 {
		normalized.PageSize = 100
	}
	return normalized
}

func (s *adminUserService) buildPayload(user *model.User) (*AdminUserPayload, error) {
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	groupName := "未分组"
	group, err := s.userGroupRepo.GetByID(user.UserGroupID)
	if err != nil {
		return nil, err
	}
	if group != nil && strings.TrimSpace(group.Name) != "" {
		groupName = strings.TrimSpace(group.Name)
	}

	payload := s.payloadFromUser(user, groupName)
	return &payload, nil
}

func (s *adminUserService) payloadFromUser(user *model.User, groupName string) AdminUserPayload {
	return AdminUserPayload{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Role:          user.Role,
		Enabled:       user.Enabled,
		UserGroupID:   user.UserGroupID,
		UserGroupName: groupName,
		Capacity:      user.Capacity,
		UsedSize:      user.UsedSize,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

func (s *adminUserService) normalizePayload(id uint, payload *AdminUserUpsertPayload, isUpdate bool) (*AdminUserUpsertPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("user payload cannot be nil")
	}

	normalized := *payload
	normalized.Username = strings.TrimSpace(normalized.Username)
	normalized.Email = strings.TrimSpace(strings.ToLower(normalized.Email))
	normalized.Password = strings.TrimSpace(normalized.Password)
	normalized.Role = strings.TrimSpace(strings.ToLower(normalized.Role))

	if normalized.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if normalized.Email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if !isUpdate && normalized.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if normalized.Password != "" && len(normalized.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}
	if normalized.Role == "" {
		normalized.Role = "user"
	}
	if normalized.Role != "admin" && normalized.Role != "user" {
		return nil, fmt.Errorf("invalid user role")
	}

	allUsers, err := s.userRepo.ListAll()
	if err != nil {
		return nil, err
	}
	for _, user := range allUsers {
		if user.ID == id {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(user.Username), normalized.Username) {
			return nil, fmt.Errorf("username already exists")
		}
		if strings.EqualFold(strings.TrimSpace(user.Email), normalized.Email) {
			return nil, fmt.Errorf("email already exists")
		}
	}

	group, err := s.userGroupRepo.GetByID(normalized.UserGroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("user group not found")
	}

	if normalized.Capacity < 0 {
		normalized.Capacity = 0
	}
	if normalized.FollowGroupCapacity {
		normalized.Capacity = group.MaxCapacity
	}

	if !isUpdate && !payload.Enabled {
		normalized.Enabled = false
	}

	return &normalized, nil
}
