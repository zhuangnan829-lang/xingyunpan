// 路径: internal/repository/collaboration_repository.go
package repository

import (
	"context"
	"fmt"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// CollaborationRepository 协作仓储接口
type CollaborationRepository interface {
	// Create 创建协作
	Create(ctx context.Context, collaboration *model.Collaboration) error

	// GetByFileID 获取文件的所有协作者
	GetByFileID(ctx context.Context, fileID uint) ([]*model.Collaboration, error)

	// GetByCollaboratorID 获取用户作为协作者的所有文件协作记录
	GetByCollaboratorID(ctx context.Context, collaboratorID uint) ([]*model.Collaboration, error)

	// GetByFileAndUser 获取特定协作关系
	GetByFileAndUser(ctx context.Context, fileID uint, collaboratorID uint) (*model.Collaboration, error)

	// UpdatePermission 更新协作权限
	UpdatePermission(ctx context.Context, fileID uint, collaboratorID uint, permission string) error

	// Delete 删除协作
	Delete(ctx context.Context, fileID uint, collaboratorID uint) error

	// IsCollaborator 检查用户是否是协作者
	IsCollaborator(ctx context.Context, fileID uint, userID uint) (bool, error)

	// GetPermission 获取用户对文件的权限
	GetPermission(ctx context.Context, fileID uint, userID uint) (string, error)
}

// collaborationRepository 协作仓储实现
type collaborationRepository struct {
	db *gorm.DB
}

// NewCollaborationRepository 创建协作仓储实例
func NewCollaborationRepository(db *gorm.DB) CollaborationRepository {
	return &collaborationRepository{db: db}
}

// Create 创建协作
func (r *collaborationRepository) Create(ctx context.Context, collaboration *model.Collaboration) error {
	if collaboration == nil {
		return fmt.Errorf("协作对象不能为空")
	}

	if err := r.db.WithContext(ctx).Create(collaboration).Error; err != nil {
		return fmt.Errorf("创建协作失败: %w", err)
	}

	return nil
}

// GetByFileID 获取文件的所有协作者
func (r *collaborationRepository) GetByFileID(ctx context.Context, fileID uint) ([]*model.Collaboration, error) {
	if fileID == 0 {
		return nil, fmt.Errorf("文件 ID 不能为空")
	}

	var collaborations []*model.Collaboration
	if err := r.db.WithContext(ctx).
		Preload("Collaborator").
		Where("file_id = ?", fileID).
		Order("created_at ASC").
		Find(&collaborations).Error; err != nil {
		return nil, fmt.Errorf("查询文件协作者失败: %w", err)
	}

	return collaborations, nil
}

// GetByCollaboratorID 获取用户作为协作者的所有文件协作记录
func (r *collaborationRepository) GetByCollaboratorID(ctx context.Context, collaboratorID uint) ([]*model.Collaboration, error) {
	if collaboratorID == 0 {
		return nil, fmt.Errorf("协作者 ID 不能为空")
	}

	var collaborations []*model.Collaboration
	if err := r.db.WithContext(ctx).
		Preload("UserFile.PhysicalFile").
		Preload("Owner").
		Where("collaborator_id = ?", collaboratorID).
		Order("created_at DESC").
		Find(&collaborations).Error; err != nil {
		return nil, fmt.Errorf("查询用户协作文件失败: %w", err)
	}

	return collaborations, nil
}

// GetByFileAndUser 获取特定协作关系
func (r *collaborationRepository) GetByFileAndUser(ctx context.Context, fileID uint, collaboratorID uint) (*model.Collaboration, error) {
	if fileID == 0 || collaboratorID == 0 {
		return nil, fmt.Errorf("文件 ID 和协作者 ID 不能为空")
	}

	var collaboration model.Collaboration
	if err := r.db.WithContext(ctx).
		Where("file_id = ? AND collaborator_id = ?", fileID, collaboratorID).
		First(&collaboration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("协作关系不存在")
		}
		return nil, fmt.Errorf("查询协作关系失败: %w", err)
	}

	return &collaboration, nil
}

// UpdatePermission 更新协作权限
func (r *collaborationRepository) UpdatePermission(ctx context.Context, fileID uint, collaboratorID uint, permission string) error {
	if fileID == 0 || collaboratorID == 0 {
		return fmt.Errorf("文件 ID 和协作者 ID 不能为空")
	}

	result := r.db.WithContext(ctx).
		Model(&model.Collaboration{}).
		Where("file_id = ? AND collaborator_id = ?", fileID, collaboratorID).
		Update("permission", permission)

	if result.Error != nil {
		return fmt.Errorf("更新协作权限失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("协作关系不存在")
	}

	return nil
}

// Delete 删除协作
func (r *collaborationRepository) Delete(ctx context.Context, fileID uint, collaboratorID uint) error {
	if fileID == 0 || collaboratorID == 0 {
		return fmt.Errorf("文件 ID 和协作者 ID 不能为空")
	}

	result := r.db.WithContext(ctx).
		Where("file_id = ? AND collaborator_id = ?", fileID, collaboratorID).
		Delete(&model.Collaboration{})

	if result.Error != nil {
		return fmt.Errorf("删除协作失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("协作关系不存在")
	}

	return nil
}

// IsCollaborator 检查用户是否是协作者
func (r *collaborationRepository) IsCollaborator(ctx context.Context, fileID uint, userID uint) (bool, error) {
	if fileID == 0 || userID == 0 {
		return false, fmt.Errorf("文件 ID 和用户 ID 不能为空")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Collaboration{}).
		Where("file_id = ? AND collaborator_id = ?", fileID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("检查协作者失败: %w", err)
	}

	return count > 0, nil
}

// GetPermission 获取用户对文件的权限
func (r *collaborationRepository) GetPermission(ctx context.Context, fileID uint, userID uint) (string, error) {
	if fileID == 0 || userID == 0 {
		return "", fmt.Errorf("文件 ID 和用户 ID 不能为空")
	}

	var collaboration model.Collaboration
	if err := r.db.WithContext(ctx).
		Where("file_id = ? AND collaborator_id = ?", fileID, userID).
		First(&collaboration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("协作关系不存在")
		}
		return "", fmt.Errorf("查询协作权限失败: %w", err)
	}

	return collaboration.Permission, nil
}
