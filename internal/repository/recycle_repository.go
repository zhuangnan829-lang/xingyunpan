// 路径: internal/repository/recycle_repository.go
package repository

import (
	"context"
	"fmt"
	"time"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

// RecycleRepository 回收站仓储接口
type RecycleRepository interface {
	// Create 创建回收站记录
	Create(ctx context.Context, item *model.RecycleBin) error

	// BatchCreate 批量创建回收站记录
	BatchCreate(ctx context.Context, items []*model.RecycleBin) error

	// GetByUserID 获取用户回收站列表（分页）
	GetByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.RecycleBin, int64, error)

	// GetByID 根据 ID 获取回收站项目
	GetByID(ctx context.Context, id uint) (*model.RecycleBin, error)

	// GetByIDs 批量获取回收站项目
	GetByIDs(ctx context.Context, ids []uint) ([]*model.RecycleBin, error)

	// Delete 删除回收站记录
	Delete(ctx context.Context, id uint) error

	// BatchDelete 批量删除回收站记录
	BatchDelete(ctx context.Context, ids []uint) error

	// DeleteAllByUserID 删除用户所有回收站记录
	DeleteAllByUserID(ctx context.Context, userID uint) error

	// GetExpiredItems 获取过期的回收站项目
	GetExpiredItems(ctx context.Context, limit int) ([]*model.RecycleBin, error)
}

// recycleRepository 回收站仓储实现
type recycleRepository struct {
	db *gorm.DB
}

// NewRecycleRepository 创建回收站仓储实例
func NewRecycleRepository(db *gorm.DB) RecycleRepository {
	return &recycleRepository{db: db}
}

// Create 创建回收站记录
func (r *recycleRepository) Create(ctx context.Context, item *model.RecycleBin) error {
	if item == nil {
		return fmt.Errorf("回收站项目不能为空")
	}

	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("创建回收站记录失败: %w", err)
	}

	return nil
}

// BatchCreate 批量创建回收站记录
func (r *recycleRepository) BatchCreate(ctx context.Context, items []*model.RecycleBin) error {
	if len(items) == 0 {
		return fmt.Errorf("回收站项目列表不能为空")
	}

	if err := r.db.WithContext(ctx).Create(&items).Error; err != nil {
		return fmt.Errorf("批量创建回收站记录失败: %w", err)
	}

	return nil
}

// GetByUserID 获取用户回收站列表（分页）
func (r *recycleRepository) GetByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.RecycleBin, int64, error) {
	if userID == 0 {
		return nil, 0, fmt.Errorf("用户 ID 不能为空")
	}

	var items []*model.RecycleBin
	var total int64

	// 查询总数
	if err := r.db.WithContext(ctx).
		Model(&model.RecycleBin{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询回收站总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("deleted_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("查询回收站列表失败: %w", err)
	}

	return items, total, nil
}

// GetByID 根据 ID 获取回收站项目
func (r *recycleRepository) GetByID(ctx context.Context, id uint) (*model.RecycleBin, error) {
	if id == 0 {
		return nil, fmt.Errorf("回收站项目 ID 不能为空")
	}

	var item model.RecycleBin
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("回收站项目不存在")
		}
		return nil, fmt.Errorf("查询回收站项目失败: %w", err)
	}

	return &item, nil
}

// GetByIDs 批量获取回收站项目
func (r *recycleRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.RecycleBin, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("回收站项目 ID 列表不能为空")
	}

	var items []*model.RecycleBin
	if err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("批量查询回收站项目失败: %w", err)
	}

	return items, nil
}

// Delete 删除回收站记录
func (r *recycleRepository) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return fmt.Errorf("回收站项目 ID 不能为空")
	}

	if err := r.db.WithContext(ctx).Delete(&model.RecycleBin{}, id).Error; err != nil {
		return fmt.Errorf("删除回收站记录失败: %w", err)
	}

	return nil
}

// BatchDelete 批量删除回收站记录
func (r *recycleRepository) BatchDelete(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return fmt.Errorf("回收站项目 ID 列表不能为空")
	}

	if err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&model.RecycleBin{}).Error; err != nil {
		return fmt.Errorf("批量删除回收站记录失败: %w", err)
	}

	return nil
}

// DeleteAllByUserID 删除用户所有回收站记录
func (r *recycleRepository) DeleteAllByUserID(ctx context.Context, userID uint) error {
	if userID == 0 {
		return fmt.Errorf("用户 ID 不能为空")
	}

	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.RecycleBin{}).Error; err != nil {
		return fmt.Errorf("删除用户所有回收站记录失败: %w", err)
	}

	return nil
}

// GetExpiredItems 获取过期的回收站项目
func (r *recycleRepository) GetExpiredItems(ctx context.Context, limit int) ([]*model.RecycleBin, error) {
	if limit <= 0 {
		limit = 100 // 默认限制
	}

	var items []*model.RecycleBin
	if err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, fmt.Errorf("查询过期回收站项目失败: %w", err)
	}

	return items, nil
}
