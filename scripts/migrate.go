// 路径: scripts/migrate.go
// 数据库迁移脚本
// 使用方法: go run scripts/migrate.go

package main

import (
	"fmt"
	"os"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/logger"
)

func main() {
	fmt.Println("=== 星云盘 V2 数据库迁移工具 ===")

	// 初始化日志
	if err := logger.Init(&logger.Config{
		Level:  "info",
		Format: "console",
		Output: "stdout",
	}); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	// 加载配置
	fmt.Println("正在加载配置...")
	if err := config.LoadDefault(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ 配置加载成功")

	// 初始化数据库连接
	fmt.Println("正在连接数据库...")
	if err := config.InitDatabase(); err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer config.CloseDatabase()
	fmt.Println("✓ 数据库连接成功")

	// 执行迁移
	fmt.Println("正在执行数据库迁移...")
	if err := config.AutoMigrate(
		&model.User{},
		&model.PhysicalFile{},
		&model.UserFile{},
		&model.MultipartUpload{},
		&model.FileDeletion{},
		// Phase 5 新增表
		&model.Share{},
		&model.ShareFile{},
		&model.RecycleBin{},
		&model.FileVersion{},
		&model.Collaboration{},
	); err != nil {
		fmt.Printf("数据库迁移失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ 数据库迁移成功")

	// 验证表是否创建
	fmt.Println("\n正在验证表结构...")
	tables := []interface{}{
		&model.User{},
		&model.PhysicalFile{},
		&model.UserFile{},
		&model.MultipartUpload{},
		&model.FileDeletion{},
		// Phase 5 新增表
		&model.Share{},
		&model.ShareFile{},
		&model.RecycleBin{},
		&model.FileVersion{},
		&model.Collaboration{},
	}

	for _, table := range tables {
		if config.DB.Migrator().HasTable(table) {
			fmt.Printf("✓ 表 %s 已创建\n", getTableName(table))
		} else {
			fmt.Printf("✗ 表 %s 创建失败\n", getTableName(table))
		}
	}

	fmt.Println("\n=== 迁移完成 ===")
}

// getTableName 获取模型的表名
func getTableName(m interface{}) string {
	switch m.(type) {
	case *model.User:
		return "users"
	case *model.PhysicalFile:
		return "physical_files"
	case *model.UserFile:
		return "user_files"
	case *model.MultipartUpload:
		return "multipart_uploads"
	case *model.FileDeletion:
		return "file_deletions"
	// Phase 5 新增表
	case *model.Share:
		return "shares"
	case *model.ShareFile:
		return "share_files"
	case *model.RecycleBin:
		return "recycle_bin"
	case *model.FileVersion:
		return "file_versions"
	case *model.Collaboration:
		return "collaborations"
	default:
		return "unknown"
	}
}
