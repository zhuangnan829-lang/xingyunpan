package main

import (
	"fmt"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/internal/service"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := config.LoadDefault(); err != nil {
		panic(fmt.Errorf("load config failed: %w", err))
	}

	db, err := gorm.Open(mysql.Open(config.Config.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("open database failed: %w", err))
	}

	repo := repository.NewFileSystemSettingRepository(db)
	svc := service.NewFileSystemSettingService(repo, nil)

	payload, err := svc.Get()
	if err != nil {
		panic(fmt.Errorf("get file system settings failed: %w", err))
	}

	updated, err := svc.Update(payload)
	if err != nil {
		panic(fmt.Errorf("update file system settings failed: %w", err))
	}

	fmt.Printf("emoji_options seeded, length=%d\n", len(updated.EmojiOptions))
}
