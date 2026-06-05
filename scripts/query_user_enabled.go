package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
)

func main() {
	if err := config.LoadDefault(); err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	db, err := gorm.Open(mysql.Open(config.Config.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db failed: %v", err)
	}

	var user model.User
	if err := db.Where("email = ?", "3518974413@qq.com").First(&user).Error; err != nil {
		log.Fatalf("query user failed: %v", err)
	}

	fmt.Printf("id=%d username=%s email=%s role=%s enabled=%v user_group_id=%d\n", user.ID, user.Username, user.Email, user.Role, user.Enabled, user.UserGroupID)
}
