package repository

import (
	"context"
	"fmt"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

type DavAccountRepository interface {
	ListByUser(ctx context.Context, userID uint) ([]model.DavAccount, error)
	GetByIDForUser(ctx context.Context, userID uint, id uint) (*model.DavAccount, error)
	GetByToken(ctx context.Context, token string) (*model.DavAccount, error)
	Create(ctx context.Context, account *model.DavAccount) error
	Save(ctx context.Context, account *model.DavAccount) error
	Delete(ctx context.Context, userID uint, id uint) error
}

type davAccountRepository struct {
	db *gorm.DB
}

func NewDavAccountRepository(db *gorm.DB) DavAccountRepository {
	return &davAccountRepository{db: db}
}

func (r *davAccountRepository) ListByUser(ctx context.Context, userID uint) ([]model.DavAccount, error) {
	var accounts []model.DavAccount
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("查询 WebDAV 账号失败: %w", err)
	}
	return accounts, nil
}

func (r *davAccountRepository) GetByIDForUser(ctx context.Context, userID uint, id uint) (*model.DavAccount, error) {
	var account model.DavAccount
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("WebDAV 账号不存在")
		}
		return nil, fmt.Errorf("查询 WebDAV 账号失败: %w", err)
	}
	return &account, nil
}

func (r *davAccountRepository) GetByToken(ctx context.Context, token string) (*model.DavAccount, error) {
	var account model.DavAccount
	if err := r.db.WithContext(ctx).
		Where("account_token = ?", token).
		First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("WebDAV 账号不存在")
		}
		return nil, fmt.Errorf("查询 WebDAV 账号失败: %w", err)
	}
	return &account, nil
}

func (r *davAccountRepository) Create(ctx context.Context, account *model.DavAccount) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return fmt.Errorf("创建 WebDAV 账号失败: %w", err)
	}
	return nil
}

func (r *davAccountRepository) Save(ctx context.Context, account *model.DavAccount) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return fmt.Errorf("保存 WebDAV 账号失败: %w", err)
	}
	return nil
}

func (r *davAccountRepository) Delete(ctx context.Context, userID uint, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.DavAccount{})
	if result.Error != nil {
		return fmt.Errorf("删除 WebDAV 账号失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("WebDAV 账号不存在")
	}
	return nil
}
