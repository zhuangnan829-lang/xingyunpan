// 创建超级管理员账户
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User 用户模型（简化版）
type User struct {
	ID       uint   `gorm:"primarykey"`
	Username string `gorm:"uniqueIndex;size:50;not null"`
	Email    string `gorm:"uniqueIndex;size:100;not null"`
	Role     string `gorm:"size:20;default:'user';not null"`
}

func (User) TableName() string {
	return "users"
}

func main() {
	fmt.Println("========================================")
	fmt.Println("星云盘 - 创建超级管理员")
	fmt.Println("========================================")
	fmt.Println()

	// 数据库连接信息
	dsn := "xingyunpan:RMGdedeMoMnMc@tcp(117.24.15.9:3306)/xingyunpan?charset=utf8mb4&parseTime=True&loc=Local"

	fmt.Println("正在连接数据库...")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	fmt.Println("✅ 数据库连接成功")
	fmt.Println()

	// 查找目标用户
	adminEmail := "123456@xingyunnan.it.com"
	adminUsername := "Sincerity"

	var adminUser User
	result := db.Where("email = ? AND username = ?", adminEmail, adminUsername).First(&adminUser)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("❌ 错误: 用户不存在")
			fmt.Println()
			fmt.Println("请先通过以下方式注册用户:")
			fmt.Println("  用户名: Sincerity")
			fmt.Println("  邮箱: 123456@xingyunnan.it.com")
			fmt.Println("  密码: #A456718293a")
			fmt.Println()
			fmt.Println("注册后再运行此脚本。")
			return
		}
		log.Fatalf("❌ 查询用户失败: %v", result.Error)
	}

	fmt.Printf("找到用户: %s (%s)\n", adminUser.Username, adminUser.Email)
	fmt.Printf("当前角色: %s\n", adminUser.Role)
	fmt.Println()

	// 更新为超级管理员
	if adminUser.Role == "admin" {
		fmt.Println("✅ 该用户已经是超级管理员")
	} else {
		fmt.Println("正在设置为超级管理员...")
		result = db.Model(&adminUser).Update("role", "admin")
		if result.Error != nil {
			log.Fatalf("❌ 更新角色失败: %v", result.Error)
		}
		fmt.Println("✅ 已设置为超级管理员")
	}
	fmt.Println()

	// 将其他用户设置为普通用户
	fmt.Println("正在将其他用户设置为普通用户...")
	result = db.Model(&User{}).
		Where("email != ? OR username != ?", adminEmail, adminUsername).
		Update("role", "user")

	if result.Error != nil {
		log.Fatalf("❌ 更新其他用户失败: %v", result.Error)
	}

	fmt.Printf("✅ 已更新 %d 个用户为普通用户\n", result.RowsAffected)
	fmt.Println()

	// 显示所有用户
	fmt.Println("========================================")
	fmt.Println("当前用户列表:")
	fmt.Println("========================================")

	var users []User
	db.Order("role DESC, id ASC").Find(&users)

	fmt.Printf("%-5s %-20s %-30s %-10s\n", "ID", "用户名", "邮箱", "角色")
	fmt.Println("------------------------------------------------------------------------")
	for _, user := range users {
		roleDisplay := user.Role
		if user.Role == "admin" {
			roleDisplay = "🔑 " + user.Role
		}
		fmt.Printf("%-5d %-20s %-30s %-10s\n", user.ID, user.Username, user.Email, roleDisplay)
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("完成！")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("超级管理员信息:")
	fmt.Println("  用户名: Sincerity")
	fmt.Println("  邮箱: 123456@xingyunnan.it.com")
	fmt.Println("  角色: admin")
	fmt.Println()
}
