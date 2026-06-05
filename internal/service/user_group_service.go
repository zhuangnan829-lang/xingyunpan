package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// UserGroupHighCapacityThresholdBytes defines the dashboard "high capacity" bucket.
// 0 means unlimited capacity and is reported separately as unlimited_groups.
const UserGroupHighCapacityThresholdBytes int64 = 200 * 1024 * 1024 * 1024

// UserGroupPayload is the API-facing user group shape.
type UserGroupPayload struct {
	ID                 uint   `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	StoragePolicyID    uint   `json:"storage_policy_id"`
	StoragePolicyName  string `json:"storage_policy_name"`
	UserCount          int64  `json:"user_count"`
	MaxCapacity        int64  `json:"max_capacity"`
	SyncMemberCapacity bool   `json:"sync_member_capacity"`
}

// UserGroupMemberPayload is the API-facing member shape for a user group.
type UserGroupMemberPayload struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	Capacity    int64  `json:"capacity"`
	UsedSize    int64  `json:"used_size"`
	UserGroupID uint   `json:"user_group_id"`
}

// UserGroupSummaryPayload is the backend-owned dashboard summary for user groups.
type UserGroupSummaryPayload struct {
	TotalGroups        int64  `json:"total_groups"`
	TotalUsers         int64  `json:"total_users"`
	DefaultGroup       string `json:"default_group"`
	HighCapacityGroups int64  `json:"high_capacity_groups"`
	UnlimitedGroups    int64  `json:"unlimited_groups"`
	HighCapacityBytes  int64  `json:"high_capacity_bytes"`
}

// UserGroupService provides admin CRUD for user groups.
type UserGroupService interface {
	List() ([]UserGroupPayload, error)
	Summary() (*UserGroupSummaryPayload, error)
	ListMembers(id uint) ([]UserGroupMemberPayload, error)
	Create(payload *UserGroupPayload) (*UserGroupPayload, error)
	Update(id uint, payload *UserGroupPayload) (*UserGroupPayload, error)
	Delete(id uint) error
}

type userGroupService struct {
	repo              repository.UserGroupRepository
	userRepo          repository.UserRepository
	storagePolicyRepo repository.StoragePolicyRepository
	siteSettingRepo   repository.SiteSettingRepository
}

// NewUserGroupService creates a user group service.
func NewUserGroupService(
	repo repository.UserGroupRepository,
	userRepo repository.UserRepository,
	storagePolicyRepo repository.StoragePolicyRepository,
	siteSettingRepo repository.SiteSettingRepository,
) UserGroupService {
	return &userGroupService{
		repo:              repo,
		userRepo:          userRepo,
		storagePolicyRepo: storagePolicyRepo,
		siteSettingRepo:   siteSettingRepo,
	}
}

func (s *userGroupService) List() ([]UserGroupPayload, error) {
	groups, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		if err := s.seedDefaults(); err != nil {
			return nil, err
		}
		groups, err = s.repo.List()
		if err != nil {
			return nil, err
		}
	}

	defaultGroupID, err := s.defaultGroupID()
	if err != nil {
		return nil, err
	}
	if err := s.repo.AssignUsersWithoutGroup(defaultGroupID); err != nil {
		return nil, err
	}

	return s.buildPayloads(groups)
}

func (s *userGroupService) Summary() (*UserGroupSummaryPayload, error) {
	groups, err := s.List()
	if err != nil {
		return nil, err
	}

	summary := &UserGroupSummaryPayload{
		TotalGroups:       int64(len(groups)),
		DefaultGroup:      "未配置",
		HighCapacityBytes: UserGroupHighCapacityThresholdBytes,
	}
	for _, group := range groups {
		summary.TotalUsers += group.UserCount
		if group.MaxCapacity == 0 {
			summary.UnlimitedGroups++
		} else if group.MaxCapacity >= UserGroupHighCapacityThresholdBytes {
			summary.HighCapacityGroups++
		}
	}

	if setting, err := s.siteSettingRepo.Get(); err != nil {
		return nil, err
	} else if setting != nil && strings.TrimSpace(setting.DefaultGroup) != "" {
		summary.DefaultGroup = strings.TrimSpace(setting.DefaultGroup)
	} else if defaultGroupID, err := s.defaultGroupID(); err != nil {
		return nil, err
	} else if defaultGroupID != 0 {
		for _, group := range groups {
			if group.ID == defaultGroupID {
				summary.DefaultGroup = group.Name
				break
			}
		}
	}

	return summary, nil
}

func (s *userGroupService) ListMembers(id uint) ([]UserGroupMemberPayload, error) {
	group, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("user group not found")
	}

	users, err := s.userRepo.ListByGroupID(id)
	if err != nil {
		return nil, err
	}

	result := make([]UserGroupMemberPayload, 0, len(users))
	for _, user := range users {
		result = append(result, UserGroupMemberPayload{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Role:        user.Role,
			Capacity:    user.Capacity,
			UsedSize:    user.UsedSize,
			UserGroupID: user.UserGroupID,
		})
	}

	return result, nil
}

func (s *userGroupService) Create(payload *UserGroupPayload) (*UserGroupPayload, error) {
	normalized, err := s.normalizePayload(0, payload)
	if err != nil {
		return nil, err
	}

	item := &model.UserGroup{
		Name:            normalized.Name,
		Description:     normalized.Description,
		StoragePolicyID: normalized.StoragePolicyID,
		MaxCapacity:     normalized.MaxCapacity,
	}
	if err := s.repo.Save(item); err != nil {
		return nil, err
	}

	return s.buildPayload(item)
}

func (s *userGroupService) Update(id uint, payload *UserGroupPayload) (*UserGroupPayload, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("user group not found")
	}

	normalized, err := s.normalizePayload(id, payload)
	if err != nil {
		return nil, err
	}

	oldName := strings.TrimSpace(existing.Name)
	newName := strings.TrimSpace(normalized.Name)
	existing.Name = normalized.Name
	existing.Description = normalized.Description
	existing.StoragePolicyID = normalized.StoragePolicyID
	existing.MaxCapacity = normalized.MaxCapacity
	if err := s.repo.Save(existing); err != nil {
		return nil, err
	}
	if oldName != newName {
		if err := s.syncDefaultGroupRename(oldName, newName); err != nil {
			return nil, err
		}
	}
	if normalized.SyncMemberCapacity {
		users, err := s.userRepo.ListByGroupID(id)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			user.Capacity = normalized.MaxCapacity
			if err := s.userRepo.Update(&user); err != nil {
				return nil, err
			}
		}
	}

	return s.buildPayload(existing)
}

func (s *userGroupService) syncDefaultGroupRename(oldName, newName string) error {
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	if oldName == "" || newName == "" {
		return nil
	}

	setting, err := s.siteSettingRepo.Get()
	if err != nil {
		return err
	}
	if setting == nil {
		return nil
	}
	if strings.TrimSpace(setting.DefaultGroup) != oldName {
		return nil
	}

	setting.DefaultGroup = newName
	return s.siteSettingRepo.Save(setting)
}

func (s *userGroupService) Delete(id uint) error {
	group, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("user group not found")
	}

	count, err := s.repo.CountUsers(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete a user group that still contains users")
	}

	items, err := s.repo.List()
	if err != nil {
		return err
	}
	if len(items) <= 1 {
		return fmt.Errorf("at least one user group must be retained")
	}

	setting, err := s.siteSettingRepo.Get()
	if err != nil {
		return err
	}
	if setting != nil && strings.TrimSpace(setting.DefaultGroup) == strings.TrimSpace(group.Name) {
		return fmt.Errorf("cannot delete the current default user group")
	}

	if err := s.ensureNoStoragePolicyGroupReference(group); err != nil {
		return err
	}

	return s.repo.Delete(id)
}

func (s *userGroupService) ensureNoStoragePolicyGroupReference(group *model.UserGroup) error {
	if group == nil || s.storagePolicyRepo == nil {
		return nil
	}

	groupName := strings.TrimSpace(group.Name)
	if groupName == "" {
		return nil
	}

	policies, err := s.storagePolicyRepo.List()
	if err != nil {
		return err
	}
	for _, policy := range policies {
		if strings.TrimSpace(policy.GroupsJSON) == "" {
			continue
		}
		var groups []string
		if err := json.Unmarshal([]byte(policy.GroupsJSON), &groups); err != nil {
			continue
		}
		for _, referencedGroup := range normalizeStoragePolicyGroups(groups) {
			if referencedGroup == groupName {
				return fmt.Errorf("cannot delete user group because storage policy %q still references it; migrate or remove the reference first", strings.TrimSpace(policy.Name))
			}
		}
	}
	return nil
}

func (s *userGroupService) buildPayloads(groups []model.UserGroup) ([]UserGroupPayload, error) {
	policies, err := s.storagePolicyRepo.List()
	if err != nil {
		return nil, err
	}
	policyNames := make(map[uint]string, len(policies))
	for _, policy := range policies {
		policyNames[policy.ID] = strings.TrimSpace(policy.Name)
	}

	result := make([]UserGroupPayload, 0, len(groups))
	for _, group := range groups {
		count, err := s.repo.CountUsers(group.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, UserGroupPayload{
			ID:                group.ID,
			Name:              group.Name,
			Description:       group.Description,
			StoragePolicyID:   group.StoragePolicyID,
			StoragePolicyName: policyNames[group.StoragePolicyID],
			UserCount:         count,
			MaxCapacity:       group.MaxCapacity,
		})
	}

	return result, nil
}

func (s *userGroupService) buildPayload(group *model.UserGroup) (*UserGroupPayload, error) {
	if group == nil {
		return nil, fmt.Errorf("user group not found")
	}

	policyName := ""
	policy, err := s.storagePolicyRepo.GetByID(group.StoragePolicyID)
	if err != nil {
		return nil, err
	}
	if policy != nil {
		policyName = strings.TrimSpace(policy.Name)
	}

	count, err := s.repo.CountUsers(group.ID)
	if err != nil {
		return nil, err
	}

	return &UserGroupPayload{
		ID:                group.ID,
		Name:              group.Name,
		Description:       group.Description,
		StoragePolicyID:   group.StoragePolicyID,
		StoragePolicyName: policyName,
		UserCount:         count,
		MaxCapacity:       group.MaxCapacity,
	}, nil
}

func (s *userGroupService) seedDefaults() error {
	policies, err := s.storagePolicyRepo.List()
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		return fmt.Errorf("at least one storage policy is required before creating user groups")
	}

	defaultPolicyID := policies[0].ID
	seeds := []model.UserGroup{
		{
			Name:            "Admin",
			Description:     "System administrators with full feature access.",
			StoragePolicyID: defaultPolicyID,
			MaxCapacity:     0,
		},
		{
			Name:            "User",
			Description:     "Default user group for newly registered accounts.",
			StoragePolicyID: defaultPolicyID,
			MaxCapacity:     50 * 1024 * 1024 * 1024,
		},
	}

	for _, seed := range seeds {
		existing, err := s.repo.GetByName(seed.Name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}
		item := seed
		if err := s.repo.Save(&item); err != nil {
			return err
		}
	}
	return nil
}

func (s *userGroupService) defaultGroupID() (uint, error) {
	setting, err := s.siteSettingRepo.Get()
	if err != nil {
		return 0, err
	}

	preferredName := "User"
	if setting != nil && strings.TrimSpace(setting.DefaultGroup) != "" {
		preferredName = strings.TrimSpace(setting.DefaultGroup)
	}

	group, err := s.repo.GetByName(preferredName)
	if err != nil {
		return 0, err
	}
	if group != nil {
		return group.ID, nil
	}

	items, err := s.repo.List()
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, nil
	}
	return items[0].ID, nil
}

func (s *userGroupService) normalizePayload(id uint, payload *UserGroupPayload) (*UserGroupPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("user group payload cannot be nil")
	}

	normalized := *payload
	normalized.Name = strings.TrimSpace(normalized.Name)
	normalized.Description = strings.TrimSpace(normalized.Description)

	if normalized.Name == "" {
		return nil, fmt.Errorf("user group name cannot be empty")
	}

	existing, err := s.repo.GetByName(normalized.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("user group name already exists")
	}

	policies, err := s.storagePolicyRepo.List()
	if err != nil {
		return nil, err
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("at least one storage policy is required")
	}

	if normalized.StoragePolicyID == 0 {
		normalized.StoragePolicyID = policies[0].ID
	}

	policy, err := s.storagePolicyRepo.GetByID(normalized.StoragePolicyID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, fmt.Errorf("storage policy not found")
	}

	normalized.StoragePolicyName = strings.TrimSpace(policy.Name)
	if normalized.MaxCapacity < 0 {
		normalized.MaxCapacity = 0
	}
	if normalized.Description == "" {
		normalized.Description = "Used for member quota, permission, and storage policy settings."
	}

	return &normalized, nil
}
