package main

import (
	"fmt"
	"log"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/crypto"
	"xingyunpan-v2/pkg/logger"
)

func main() {
	// 加载配置
	if err := config.LoadDefault(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logConfig := &logger.Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}
	if err := logger.Init(logConfig); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化数据库
	if err := config.InitDatabase(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer config.CloseDatabase()

	db := config.GetDB()

	// 1. 添加 role 字段（如果不存在）
	fmt.Println("步骤 1: 添加 role 字段到 users 表...")

	// 检查字段是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'role'").Scan(&count)

	if count == 0 {
		// 字段不存在，添加它
		if err := db.Exec("ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user' NOT NULL COMMENT '用户角色(admin/user)'").Error; err != nil {
			log.Fatalf("添加 role 字段失败: %v", err)
		}
		fmt.Println("✓ role 字段添加成功")
	} else {
		fmt.Println("✓ role 字段已存在")
	}

	// 2. 检查超级管理员是否存在
	fmt.Println("\n步骤 2: 检查超级管理员账户...")
	var admin model.User
	result := db.Where("email = ? AND username = ?", "123456@xingyunnan.it.com", "Sincerity").First(&admin)

	if result.Error != nil {
		// 管理员不存在，创建新账户
		fmt.Println("超级管理员不存在，正在创建...")

		hashedPassword, err := crypto.HashPassword("#A456718293a")
		if err != nil {
			log.Fatalf("密码加密失败: %v", err)
		}

		admin = model.User{
			Username: "Sincerity",
			Email:    "123456@xingyunnan.it.com",
			Password: hashedPassword,
			Role:     "admin",
			Capacity: 107374182400, // 100GB
			UsedSize: 0,
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Fatalf("创建超级管理员失败: %v", err)
		}
		fmt.Printf("✓ 超级管理员创建成功 (ID: %d)\n", admin.ID)
	} else {
		// 管理员已存在，更新为 admin 角色
		fmt.Printf("找到用户 (ID: %d)，正在设置为超级管理员...\n", admin.ID)
		if err := db.Model(&admin).Update("role", "admin").Error; err != nil {
			log.Fatalf("更新用户角色失败: %v", err)
		}
		fmt.Println("✓ 用户已设置为超级管理员")
	}

	// 3. 确保所有其他用户都是普通用户
	fmt.Println("\n步骤 3: 设置其他用户为普通用户...")
	result = db.Model(&model.User{}).
		Where("id != ?", admin.ID).
		Update("role", "user")

	if result.Error != nil {
		log.Fatalf("更新其他用户角色失败: %v", result.Error)
	}
	fmt.Printf("✓ 已将 %d 个用户设置为普通用户\n", result.RowsAffected)

	// 4. 显示所有用户及其角色
	fmt.Println("\n步骤 4: 当前用户列表:")
	fmt.Println("----------------------------------------")
	var users []model.User
	db.Find(&users)

	if len(users) == 0 {
		fmt.Println("数据库中没有用户")
	} else {
		for _, user := range users {
			roleLabel := "普通用户"
			if user.Role == "admin" {
				roleLabel = "超级管理员"
			}
			fmt.Printf("ID: %d | 用户名: %s | 邮箱: %s | 角色: %s\n",
				user.ID, user.Username, user.Email, roleLabel)
		}
	}
	fmt.Println("----------------------------------------")

	fmt.Println("\n✓ 所有操作完成!")
	fmt.Println("\n超级管理员登录信息:")
	fmt.Println("  用户名: Sincerity")
	fmt.Println("  邮箱: 123456@xingyunnan.it.com")
	fmt.Println("  密码: #A456718293a")
}
