package service

import (
	"errors"
	"fmt"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

var ErrQuotaExceeded = errors.New("quota exceeded: capacity not enough")

func reserveUserCapacity(tx *gorm.DB, userID uint, delta int64) error {
	if tx == nil {
		return fmt.Errorf("database is not initialized")
	}
	if userID == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}
	if delta <= 0 {
		return nil
	}

	result := tx.Model(&model.User{}).
		Where("id = ? AND (capacity = 0 OR used_size <= capacity - ?)", userID, delta).
		UpdateColumn("used_size", gorm.Expr("used_size + ?", delta))
	if result.Error != nil {
		return fmt.Errorf("update used capacity failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrQuotaExceeded
	}
	return nil
}
