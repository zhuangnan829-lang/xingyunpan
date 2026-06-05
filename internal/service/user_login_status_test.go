package service

import (
	"path/filepath"
	"strings"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/crypto"
)

func TestDisabledUserLoginFails(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "disabled-login.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	password, err := crypto.HashPassword("secret1")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	disabledUser := model.User{
		Username: "disabled-user",
		Email:    "disabled@example.com",
		Password: password,
		Role:     "user",
		Enabled:  true,
	}
	if err := db.Create(&disabledUser).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Model(&model.User{}).Where("id = ?", disabledUser.ID).Update("enabled", false).Error; err != nil {
		t.Fatalf("disable user: %v", err)
	}

	userService := NewUserService(
		repository.NewUserRepository(db),
		nil,
		nil,
		"login-test-secret",
		24,
		168,
		0,
		nil,
		nil,
		nil,
		300,
		60,
	)

	_, err = userService.Login("disabled-user", "secret1")
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("Login error = %v, want disabled account error", err)
	}
}
