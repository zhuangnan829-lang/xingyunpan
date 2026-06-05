// 路径: internal/repository/version_repository.go
package repository

import (
	"context"
	"fmt"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// VersionRepository 版本仓储接口
type VersionRepository interface {
	// Create 创建版本
	Create(ctx context.Context, version *model.FileVersion) error

	// GetByFileID 获取文件的所有版本
	GetByFileID(ctx context.Context, fileID uint) ([]*model.FileVersion, error)

	// GetByID 获取特定版本
	GetByID(ctx context.Context, versionID uint) (*model.FileVersion, error)

	// GetCurrentVersion 获取当前版本
	GetCurrentVersion(ctx context.Context, fileID uint) (*model.FileVersion, error)

	// SetCurrentVersion 标记版本为当前版本
	SetCurrentVersion(ctx context.Context, fileID uint, versionID uint) error

	// Delete 删除版本
	Delete(ctx context.Context, versionID uint) error

	// CountByFileID 获取文件版本数量
	CountByFileID(ctx context.Context, fileID uint) (int64, error)

	// GetOldestVersion 获取最旧的版本
	GetOldestVersion(ctx context.Context, fileID uint) (*model.FileVersion, error)
}

// versionRepository 版本仓储实现
type versionRepository struct {
	db *gorm.DB
}

// NewVersionRepository 创建版本仓储实例
func NewVersionRepository(db *gorm.DB) VersionRepository {
	return &versionRepository{db: db}
}

// Create 创建版本
func (r *versionRepository) Create(ctx context.Context, version *model.FileVersion) error {
	if version == nil {
		return fmt.Errorf("版本不能为空")
	}

	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return fmt.Errorf("创建版本失败: %w", err)
	}

	return nil
}

// GetByFileID 获取文件的所有版本
func (r *versionRepository) GetByFileID(ctx context.Context, fileID uint) ([]*model.FileVersion, error) {
	if fileID == 0 {
		return nil, fmt.Errorf("文件 ID 不能为空")
	}

	var versions []*model.FileVersion
	if err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("version_number DESC").
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("查询文件版本失败: %w", err)
	}

	return versions, nil
}

// GetByID 获取特定版本
func (r *versionRepository) GetByID(ctx context.Context, versionID uint) (*model.FileVersion, error) {
	if versionID == 0 {
		return nil, fmt.Errorf("版本 ID 不能为空")
	}

	var version model.FileVersion
	if err := r.db.WithContext(ctx).First(&version, versionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("版本不存在")
		}
		return nil, fmt.Errorf("查询版本失败: %w", err)
	}

	return &version, nil
}

// GetCurrentVersion 获取当前版本
func (r *versionRepository) GetCurrentVersion(ctx context.Context, fileID uint) (*model.FileVersion, error) {
	if fileID == 0 {
		return nil, fmt.Errorf("文件 ID 不能为空")
	}

	var version model.FileVersion
	if err := r.db.WithContext(ctx).
		Where("file_id = ? AND is_current = ?", fileID, true).
		First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("当前版本不存在")
		}
		return nil, fmt.Errorf("查询当前版本失败: %w", err)
	}

	return &version, nil
}

// SetCurrentVersion 标记版本为当前版本
func (r *versionRepository) SetCurrentVersion(ctx context.Context, fileID uint, versionID uint) error {
	if fileID == 0 || versionID == 0 {
		return fmt.Errorf("文件 ID 和版本 ID 不能为空")
	}

	// 使用事务确保原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 将该文件的所有版本标记为非当前版本
		if err := tx.Model(&model.FileVersion{}).
			Where("file_id = ?", fileID).
			Update("is_current", false).Error; err != nil {
			return fmt.Errorf("更新版本状态失败: %w", err)
		}

		// 2. 将指定版本标记为当前版本
		if err := tx.Model(&model.FileVersion{}).
			Where("id = ?", versionID).
			Update("is_current", true).Error; err != nil {
			return fmt.Errorf("设置当前版本失败: %w", err)
		}

		return nil
	})
}

// Delete 删除版本
func (r *versionRepository) Delete(ctx context.Context, versionID uint) error {
	if versionID == 0 {
		return fmt.Errorf("版本 ID 不能为空")
	}

	if err := r.db.WithContext(ctx).Delete(&model.FileVersion{}, versionID).Error; err != nil {
		return fmt.Errorf("删除版本失败: %w", err)
	}

	return nil
}

// CountByFileID 获取文件版本数量
func (r *versionRepository) CountByFileID(ctx context.Context, fileID uint) (int64, error) {
	if fileID == 0 {
		return 0, fmt.Errorf("文件 ID 不能为空")
	}

	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.FileVersion{}).
		Where("file_id = ?", fileID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("查询版本数量失败: %w", err)
	}

	return count, nil
}

// GetOldestVersion 获取最旧的版本
func (r *versionRepository) GetOldestVersion(ctx context.Context, fileID uint) (*model.FileVersion, error) {
	if fileID == 0 {
		return nil, fmt.Errorf("文件 ID 不能为空")
	}

	var version model.FileVersion
	if err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("version_number ASC").
		First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("版本不存在")
		}
		return nil, fmt.Errorf("查询最旧版本失败: %w", err)
	}

	return &version, nil
}
