package main

import (
	"fmt"
	"log"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := config.LoadDefault(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	db, err := gorm.Open(mysql.Open(config.Config.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	var users []model.User
	if err := db.Order("id asc").Find(&users).Error; err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}

	if len(users) == 0 {
		fmt.Println("未查询到用户")
		return
	}

	fmt.Printf("共查询到 %d 个用户\n", len(users))
	for _, user := range users {
		hashSummary := "无"
		if user.Password != "" {
			hashSummary = fmt.Sprintf("已存储 bcrypt 哈希（长度 %d）", len(user.Password))
		}

		fmt.Printf(
			"ID=%d | 用户名=%s | 邮箱=%s | 角色=%s | 创建时间=%s | 密码=%s\n",
			user.ID,
			user.Username,
			user.Email,
			user.Role,
			user.CreatedAt.Format("2006-01-02 15:04:05"),
			hashSummary,
		)
	}
}
