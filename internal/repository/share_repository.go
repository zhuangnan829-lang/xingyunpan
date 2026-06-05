// 路径: internal/repository/share_repository.go
package repository

import (
	"context"
	"fmt"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// ShareRepository 分享仓储接口
type ShareRepository interface {
	// Create 创建分享和关联文件
	Create(ctx context.Context, share *model.Share, fileIDs []uint) error

	// GetByToken 根据 share_token 查询分享
	GetByToken(ctx context.Context, token string) (*model.Share, error)

	// GetByID 根据数字 ID 查询分享
	GetByID(ctx context.Context, shareID uint) (*model.Share, error)

	// GetByUserID 获取用户的所有分享
	GetByUserID(ctx context.Context, userID uint) ([]*model.Share, error)

	// Delete 删除分享（级联删除 share_files）
	Delete(ctx context.Context, shareID uint) error

	// IncrementDownloadCount 原子性增加下载计数
	IncrementDownloadCount(ctx context.Context, shareID uint) error

	// IncrementAccessCount 原子性增加访问计数
	IncrementAccessCount(ctx context.Context, shareID uint) error

	// GetShareFiles 获取分享的文件列表
	GetShareFiles(ctx context.Context, shareID uint) ([]*model.UserFile, error)
}

// shareRepository 分享仓储实现
type shareRepository struct {
	db *gorm.DB
}

// NewShareRepository 创建分享仓储实例
func NewShareRepository(db *gorm.DB) ShareRepository {
	return &shareRepository{db: db}
}

// Create 创建分享和关联文件
func (r *shareRepository) Create(ctx context.Context, share *model.Share, fileIDs []uint) error {
	if share == nil {
		return fmt.Errorf("分享对象不能为空")
	}
	if len(fileIDs) == 0 {
		return fmt.Errorf("文件列表不能为空")
	}

	// 使用事务确保原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 创建分享记录
		if err := tx.Create(share).Error; err != nil {
			return fmt.Errorf("创建分享失败: %w", err)
		}

		// 2. 批量创建分享文件关联
		shareFiles := make([]model.ShareFile, len(fileIDs))
		for i, fileID := range fileIDs {
			shareFiles[i] = model.ShareFile{
				ShareID: share.ID,
				FileID:  fileID,
			}
		}

		if err := tx.Create(&shareFiles).Error; err != nil {
			return fmt.Errorf("创建分享文件关联失败: %w", err)
		}

		return nil
	})
}

// GetByToken 根据 share_token 查询分享
func (r *shareRepository) GetByToken(ctx context.Context, token string) (*model.Share, error) {
	if token == "" {
		return nil, fmt.Errorf("分享令牌不能为空")
	}

	var share model.Share
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("share_token = ?", token).
		First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("分享不存在")
		}
		return nil, fmt.Errorf("查询分享失败: %w", err)
	}

	return &share, nil
}

// GetByID 根据数字 ID 查询分享
func (r *shareRepository) GetByID(ctx context.Context, shareID uint) (*model.Share, error) {
	if shareID == 0 {
		return nil, fmt.Errorf("分享 ID 不能为空")
	}

	var share model.Share
	if err := r.db.WithContext(ctx).
		Preload("User").
		First(&share, shareID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("分享不存在")
		}
		return nil, fmt.Errorf("查询分享失败: %w", err)
	}

	return &share, nil
}

// GetByUserID 获取用户的所有分享
func (r *shareRepository) GetByUserID(ctx context.Context, userID uint) ([]*model.Share, error) {
	if userID == 0 {
		return nil, fmt.Errorf("用户 ID 不能为空")
	}

	var shares []*model.Share
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("查询用户分享失败: %w", err)
	}

	return shares, nil
}

// Delete 删除分享（级联删除 share_files）
func (r *shareRepository) Delete(ctx context.Context, shareID uint) error {
	if shareID == 0 {
		return fmt.Errorf("分享 ID 不能为空")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("share_id = ?", shareID).Delete(&model.ShareFile{}).Error; err != nil {
			return fmt.Errorf("删除分享文件关联失败: %w", err)
		}
		if err := tx.Delete(&model.Share{}, shareID).Error; err != nil {
			return fmt.Errorf("删除分享失败: %w", err)
		}
		return nil
	})
}

// IncrementDownloadCount 原子性增加下载计数
func (r *shareRepository) IncrementDownloadCount(ctx context.Context, shareID uint) error {
	if shareID == 0 {
		return fmt.Errorf("分享 ID 不能为空")
	}

	// 使用原子性更新
	if err := r.db.WithContext(ctx).
		Model(&model.Share{}).
		Where("id = ?", shareID).
		UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("增加下载计数失败: %w", err)
	}

	return nil
}

// IncrementAccessCount 原子性增加访问计数
func (r *shareRepository) IncrementAccessCount(ctx context.Context, shareID uint) error {
	if shareID == 0 {
		return fmt.Errorf("分享 ID 不能为空")
	}

	// 使用原子性更新
	if err := r.db.WithContext(ctx).
		Model(&model.Share{}).
		Where("id = ?", shareID).
		UpdateColumn("access_count", gorm.Expr("access_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("增加访问计数失败: %w", err)
	}

	return nil
}

// GetShareFiles 获取分享的文件列表
func (r *shareRepository) GetShareFiles(ctx context.Context, shareID uint) ([]*model.UserFile, error) {
	if shareID == 0 {
		return nil, fmt.Errorf("分享 ID 不能为空")
	}

	var files []*model.UserFile
	if err := r.db.WithContext(ctx).
		Joins("JOIN share_files ON share_files.file_id = user_files.id").
		Where("share_files.share_id = ?", shareID).
		Where("user_files.deleted_at IS NULL").
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("查询分享文件失败: %w", err)
	}

	return files, nil
}
