package repository

import (
	"fmt"

	"gorm.io/gorm"
	"xingyunpan-v2/internal/model"
)

// UserGroupRepository handles persistence of user groups.
type UserGroupRepository interface {
	EnsureSchema() error
	List() ([]model.UserGroup, error)
	GetByID(id uint) (*model.UserGroup, error)
	GetByName(name string) (*model.UserGroup, error)
	Save(group *model.UserGroup) error
	Delete(id uint) error
	CountUsers(groupID uint) (int64, error)
	AssignUsersWithoutGroup(groupID uint) error
}

type userGroupRepository struct {
	db *gorm.DB
}

// NewUserGroupRepository creates a user group repository.
func NewUserGroupRepository(db *gorm.DB) UserGroupRepository {
	return &userGroupRepository{db: db}
}

// EnsureSchema keeps user groups and user-group relations in sync with models.
func (r *userGroupRepository) EnsureSchema() error {
	if r.db == nil {
		return fmt.Errorf("user groups database is not initialized")
	}

	// Avoid forcing a full AutoMigrate on every request for existing tables.
	// Some historical MySQL schemas have index drift on user_groups.name, and
	// GORM may try to drop a non-existent unique key during reconciliation.
	if !r.db.Migrator().HasTable(&model.UserGroup{}) {
		if err := r.db.AutoMigrate(&model.UserGroup{}); err != nil {
			return fmt.Errorf("migrate user groups failed: %w", err)
		}
	} else {
		requiredColumns := []string{"name", "description", "storage_policy_id", "max_capacity"}
		for _, column := range requiredColumns {
			if !r.db.Migrator().HasColumn(&model.UserGroup{}, column) {
				if err := r.db.AutoMigrate(&model.UserGroup{}); err != nil {
					return fmt.Errorf("migrate user groups failed: %w", err)
				}
				break
			}
		}
	}

	if !r.db.Migrator().HasTable(&model.User{}) {
		if err := r.db.AutoMigrate(&model.User{}); err != nil {
			return fmt.Errorf("migrate users for user groups failed: %w", err)
		}
	} else if !r.db.Migrator().HasColumn(&model.User{}, "user_group_id") {
		if err := r.db.Migrator().AddColumn(&model.User{}, "UserGroupID"); err != nil {
			return fmt.Errorf("add user_group_id column failed: %w", err)
		}
	}

	return nil
}

func (r *userGroupRepository) List() ([]model.UserGroup, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var items []model.UserGroup
	if err := r.db.Order("id asc").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("query user groups failed: %w", err)
	}
	return items, nil
}

func (r *userGroupRepository) GetByID(id uint) (*model.UserGroup, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var item model.UserGroup
	if err := r.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query user group failed: %w", err)
	}
	return &item, nil
}

func (r *userGroupRepository) GetByName(name string) (*model.UserGroup, error) {
	if err := r.EnsureSchema(); err != nil {
		return nil, err
	}

	var item model.UserGroup
	if err := r.db.Where("name = ?", name).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query user group by name failed: %w", err)
	}
	return &item, nil
}

func (r *userGroupRepository) Save(group *model.UserGroup) error {
	if group == nil {
		return fmt.Errorf("user group cannot be nil")
	}

	if err := r.EnsureSchema(); err != nil {
		return err
	}

	if group.ID == 0 {
		if err := r.db.Create(group).Error; err != nil {
			return fmt.Errorf("create user group failed: %w", err)
		}
		return nil
	}

	if err := r.db.Save(group).Error; err != nil {
		return fmt.Errorf("update user group failed: %w", err)
	}
	return nil
}

func (r *userGroupRepository) Delete(id uint) error {
	if err := r.EnsureSchema(); err != nil {
		return err
	}

	if err := r.db.Delete(&model.UserGroup{}, id).Error; err != nil {
		return fmt.Errorf("delete user group failed: %w", err)
	}
	return nil
}

func (r *userGroupRepository) CountUsers(groupID uint) (int64, error) {
	if err := r.EnsureSchema(); err != nil {
		return 0, err
	}

	var count int64
	if err := r.db.Model(&model.User{}).Where("user_group_id = ?", groupID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count user group members failed: %w", err)
	}
	return count, nil
}

func (r *userGroupRepository) AssignUsersWithoutGroup(groupID uint) error {
	if groupID == 0 {
		return nil
	}
	if err := r.EnsureSchema(); err != nil {
		return err
	}

	if err := r.db.Model(&model.User{}).Where("user_group_id = 0").Update("user_group_id", groupID).Error; err != nil {
		return fmt.Errorf("assign default user group failed: %w", err)
	}
	return nil
}
