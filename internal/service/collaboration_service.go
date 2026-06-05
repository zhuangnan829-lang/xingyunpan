// 路径: internal/service/collaboration_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/cache"
)

// CollaborationService 协作服务接口
type CollaborationService interface {
	// AddCollaborator 添加协作者
	AddCollaborator(ctx context.Context, ownerID uint, fileID uint, username string, permission string) (*CollaboratorInfo, error)

	// GetCollaborators 获取协作者列表
	GetCollaborators(ctx context.Context, ownerID uint, fileID uint) ([]*CollaboratorInfo, error)

	// UpdateCollaboratorPermission 更新协作者权限
	UpdateCollaboratorPermission(ctx context.Context, ownerID uint, fileID uint, collaboratorID uint, permission string) error

	// RemoveCollaborator 移除协作者
	RemoveCollaborator(ctx context.Context, ownerID uint, fileID uint, collaboratorID uint) error

	// GetMyCollaborations 获取我的协作文件列表
	GetMyCollaborations(ctx context.Context, userID uint) ([]*CollaborationFileInfo, error)

	// CheckFilePermission 检查文件权限
	CheckFilePermission(ctx context.Context, userID uint, fileID uint) (*FilePermissionInfo, error)
}

// CollaboratorInfo 协作者信息
type CollaboratorInfo struct {
	UserID     uint      `json:"user_id"`
	Username   string    `json:"username"`
	Permission string    `json:"permission"`
	AddedAt    time.Time `json:"added_at"`
}

// CollaborationFileInfo 协作文件信息
type CollaborationFileInfo struct {
	FileID      uint      `json:"file_id"`
	FileName    string    `json:"file_name"`
	FileType    string    `json:"file_type"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	IsFolder    bool      `json:"is_folder"`
	OwnerName   string    `json:"owner_name"`
	Permission  string    `json:"permission"`
	SharedAt    time.Time `json:"shared_at"`
}

// FilePermissionInfo 文件权限信息
type FilePermissionInfo struct {
	CanView     bool `json:"can_view"`
	CanDownload bool `json:"can_download"`
	CanEdit     bool `json:"can_edit"`
	IsOwner     bool `json:"is_owner"`
}

// validPermissions 有效的权限值
var validPermissions = map[string]bool{
	"view":     true,
	"download": true,
	"edit":     true,
}

// collaborationService 协作服务实现
type collaborationService struct {
	collaborationRepo repository.CollaborationRepository
	userFileRepo      repository.UserFileRepository
	userRepo          repository.UserRepository
	cache             *cache.CacheService
}

// NewCollaborationService 创建协作服务实例
func NewCollaborationService(
	collaborationRepo repository.CollaborationRepository,
	userFileRepo repository.UserFileRepository,
	userRepo repository.UserRepository,
	cache *cache.CacheService,
) CollaborationService {
	return &collaborationService{
		collaborationRepo: collaborationRepo,
		userFileRepo:      userFileRepo,
		userRepo:          userRepo,
		cache:             cache,
	}
}

// AddCollaborator 添加协作者
func (s *collaborationService) AddCollaborator(ctx context.Context, ownerID uint, fileID uint, username string, permission string) (*CollaboratorInfo, error) {
	// 1. 验证文件存在且用户是所有者
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("文件不存在")
	}
	if file.UserID != ownerID {
		return nil, fmt.Errorf("无权限管理此文件的协作者")
	}

	// 2. 验证 permission 值
	if !validPermissions[permission] {
		return nil, fmt.Errorf("无效的权限值，必须为 view、download 或 edit 之一")
	}

	// 3. 根据 username 查找用户
	collaborator, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %s", username)
	}

	// 4. 不能将自己添加为协作者
	if collaborator.ID == ownerID {
		return nil, fmt.Errorf("不能将自己添加为协作者")
	}

	// 5. 检查唯一性约束（是否已经是协作者）
	isCollab, err := s.collaborationRepo.IsCollaborator(ctx, fileID, collaborator.ID)
	if err != nil {
		return nil, fmt.Errorf("检查协作者失败: %w", err)
	}
	if isCollab {
		return nil, fmt.Errorf("该用户已经是协作者")
	}

	// 6. 创建协作记录
	collaboration := &model.Collaboration{
		FileID:         fileID,
		OwnerID:        ownerID,
		CollaboratorID: collaborator.ID,
		Permission:     permission,
	}

	if err := s.collaborationRepo.Create(ctx, collaboration); err != nil {
		return nil, fmt.Errorf("添加协作者失败: %w", err)
	}

	// 7. 清除协作者列表缓存
	s.invalidateCollaboratorsCache(ctx, fileID)

	return &CollaboratorInfo{
		UserID:     collaborator.ID,
		Username:   collaborator.Username,
		Permission: permission,
		AddedAt:    collaboration.CreatedAt,
	}, nil
}

// GetCollaborators 获取协作者列表
func (s *collaborationService) GetCollaborators(ctx context.Context, ownerID uint, fileID uint) ([]*CollaboratorInfo, error) {
	// 1. 验证文件存在且用户是所有者
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("文件不存在")
	}
	if file.UserID != ownerID {
		return nil, fmt.Errorf("无权限查看此文件的协作者")
	}

	// 2. 尝试从缓存获取
	cacheKey := fmt.Sprintf("collaborators:%d", fileID)
	var cachedResult []*CollaboratorInfo
	if s.cache != nil {
		if err := s.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
			return cachedResult, nil
		}
	}

	// 3. 查询协作者列表
	collaborations, err := s.collaborationRepo.GetByFileID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("查询协作者失败: %w", err)
	}

	// 4. 构建响应
	result := make([]*CollaboratorInfo, 0, len(collaborations))
	for _, c := range collaborations {
		result = append(result, &CollaboratorInfo{
			UserID:     c.CollaboratorID,
			Username:   c.Collaborator.Username,
			Permission: c.Permission,
			AddedAt:    c.CreatedAt,
		})
	}

	// 5. 缓存结果（TTL=2分钟）
	if s.cache != nil {
		_ = s.cache.Set(ctx, cacheKey, result, 2*time.Minute)
	}

	return result, nil
}

// UpdateCollaboratorPermission 更新协作者权限
func (s *collaborationService) UpdateCollaboratorPermission(ctx context.Context, ownerID uint, fileID uint, collaboratorID uint, permission string) error {
	// 1. 验证文件存在且用户是所有者
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if file.UserID != ownerID {
		return fmt.Errorf("无权限管理此文件的协作者")
	}

	// 2. 验证 permission 值
	if !validPermissions[permission] {
		return fmt.Errorf("无效的权限值，必须为 view、download 或 edit 之一")
	}

	// 3. 更新权限
	if err := s.collaborationRepo.UpdatePermission(ctx, fileID, collaboratorID, permission); err != nil {
		return fmt.Errorf("更新协作者权限失败: %w", err)
	}

	// 4. 清除缓存
	s.invalidateCollaboratorsCache(ctx, fileID)

	return nil
}

// RemoveCollaborator 移除协作者
func (s *collaborationService) RemoveCollaborator(ctx context.Context, ownerID uint, fileID uint, collaboratorID uint) error {
	// 1. 验证文件存在且用户是所有者
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在")
	}
	if file.UserID != ownerID {
		return fmt.Errorf("无权限管理此文件的协作者")
	}

	// 2. 删除协作记录
	if err := s.collaborationRepo.Delete(ctx, fileID, collaboratorID); err != nil {
		return fmt.Errorf("移除协作者失败: %w", err)
	}

	// 3. 清除缓存
	s.invalidateCollaboratorsCache(ctx, fileID)

	return nil
}

// GetMyCollaborations 获取我的协作文件列表
func (s *collaborationService) GetMyCollaborations(ctx context.Context, userID uint) ([]*CollaborationFileInfo, error) {
	// 查询用户作为协作者的所有文件
	collaborations, err := s.collaborationRepo.GetByCollaboratorID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询协作文件失败: %w", err)
	}

	result := make([]*CollaborationFileInfo, 0, len(collaborations))
	for _, c := range collaborations {
		contentType := ""
		if c.UserFile.PhysicalFile != nil {
			contentType = c.UserFile.PhysicalFile.ContentType
		}

		result = append(result, &CollaborationFileInfo{
			FileID:      c.FileID,
			FileName:    c.UserFile.FileName,
			FileType:    contentType,
			ContentType: contentType,
			FileSize:    c.UserFile.FileSize,
			IsFolder:    c.UserFile.IsFolder,
			OwnerName:   c.Owner.Username,
			Permission:  c.Permission,
			SharedAt:    c.CreatedAt,
		})
	}

	return result, nil
}

// CheckFilePermission 检查文件权限
func (s *collaborationService) CheckFilePermission(ctx context.Context, userID uint, fileID uint) (*FilePermissionInfo, error) {
	// 1. 查询文件
	file, err := s.userFileRepo.GetByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("文件不存在")
	}

	// 2. 如果是所有者，返回全部权限
	if file.UserID == userID {
		return &FilePermissionInfo{
			CanView:     true,
			CanDownload: true,
			CanEdit:     true,
			IsOwner:     true,
		}, nil
	}

	// 3. 检查是否是协作者
	permission, err := s.collaborationRepo.GetPermission(ctx, fileID, userID)
	if err != nil {
		// 不是协作者，无权限
		return &FilePermissionInfo{
			CanView:     false,
			CanDownload: false,
			CanEdit:     false,
			IsOwner:     false,
		}, nil
	}

	// 4. 根据权限级别返回
	info := &FilePermissionInfo{IsOwner: false}
	switch permission {
	case "edit":
		info.CanView = true
		info.CanDownload = true
		info.CanEdit = true
	case "download":
		info.CanView = true
		info.CanDownload = true
		info.CanEdit = false
	case "view":
		info.CanView = true
		info.CanDownload = false
		info.CanEdit = false
	}

	return info, nil
}

// invalidateCollaboratorsCache 清除协作者列表缓存
func (s *collaborationService) invalidateCollaboratorsCache(ctx context.Context, fileID uint) {
	if s.cache != nil {
		cacheKey := fmt.Sprintf("collaborators:%d", fileID)
		_ = s.cache.Set(ctx, cacheKey, nil, 0)
	}
}
