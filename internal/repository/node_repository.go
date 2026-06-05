package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// NodeRepository handles persistence of compute nodes.
type NodeRepository interface {
	EnsureSchema() error
	List() ([]model.Node, error)
	GetByID(id uint) (*model.Node, error)
	Save(node *model.Node) error
	Delete(id uint) error
}

type nodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository creates a node repository.
func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepository{db: db}
}

// EnsureSchema keeps the nodes table schema in sync with the model.
func (r *nodeRepository) EnsureSchema() error {
	if r.db == nil {
		return fmt.Errorf("nodes database is not initialized")
	}

	if err := r.db.AutoMigrate(&model.Node{}); err != nil {
		return fmt.Errorf("migrate nodes failed: %w", err)
	}

	return nil
}

func (r *nodeRepository) List() ([]model.Node, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var items []model.Node
	if err := r.db.Order("is_built_in desc, id asc").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("query nodes failed: %w", err)
	}
	return items, nil
}

func (r *nodeRepository) GetByID(id uint) (*model.Node, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var item model.Node
	if err := r.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query node failed: %w", err)
	}
	return &item, nil
}

func (r *nodeRepository) Save(node *model.Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	if err := r.EnsureSchema(); err != nil {
		return err
	}

	if node.ID == 0 {
		if err := r.db.Create(node).Error; err != nil {
			return fmt.Errorf("create node failed: %w", err)
		}
		return nil
	}

	if err := r.db.Save(node).Error; err != nil {
		return fmt.Errorf("update node failed: %w", err)
	}

	return nil
}

func (r *nodeRepository) Delete(id uint) error {
	if err := r.EnsureSchema(); err != nil {
		return err
	}

	if err := r.db.Delete(&model.Node{}, id).Error; err != nil {
		return fmt.Errorf("delete node failed: %w", err)
	}

	return nil
}
