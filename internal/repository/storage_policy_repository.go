package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// StoragePolicyRepository handles persistence of storage policies.
type StoragePolicyRepository interface {
	EnsureSchema() error
	List() ([]model.StoragePolicy, error)
	GetByID(id uint) (*model.StoragePolicy, error)
	Save(policy *model.StoragePolicy) error
	Delete(id uint) error
}

type storagePolicyRepository struct {
	db *gorm.DB
}

// NewStoragePolicyRepository creates a storage policy repository.
func NewStoragePolicyRepository(db *gorm.DB) StoragePolicyRepository {
	return &storagePolicyRepository{db: db}
}

// EnsureSchema creates the storage policies table if it is missing.
func (r *storagePolicyRepository) EnsureSchema() error {
	if r.db == nil {
		return fmt.Errorf("storage policies database is not initialized")
	}

	if r.db.Migrator().HasTable(&model.StoragePolicy{}) {
		return nil
	}

	if err := r.db.AutoMigrate(&model.StoragePolicy{}); err != nil {
		return fmt.Errorf("migrate storage policies failed: %w", err)
	}

	return nil
}

func (r *storagePolicyRepository) List() ([]model.StoragePolicy, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var items []model.StoragePolicy
	if err := r.db.Order("id asc").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("query storage policies failed: %w", err)
	}
	return items, nil
}

func (r *storagePolicyRepository) GetByID(id uint) (*model.StoragePolicy, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var item model.StoragePolicy
	if err := r.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query storage policy failed: %w", err)
	}
	return &item, nil
}

func (r *storagePolicyRepository) Save(policy *model.StoragePolicy) error {
	if policy == nil {
		return fmt.Errorf("storage policy cannot be nil")
	}

	if err := r.EnsureSchema(); err != nil {
		return err
	}

	if policy.ID == 0 {
		if err := r.db.Create(policy).Error; err != nil {
			return fmt.Errorf("create storage policy failed: %w", err)
		}
		return nil
	}

	if err := r.db.Save(policy).Error; err != nil {
		return fmt.Errorf("update storage policy failed: %w", err)
	}
	return nil
}

func (r *storagePolicyRepository) Delete(id uint) error {
	if err := r.EnsureSchema(); err != nil {
		return err
	}

	if err := r.db.Delete(&model.StoragePolicy{}, id).Error; err != nil {
		return fmt.Errorf("delete storage policy failed: %w", err)
	}
	return nil
}
