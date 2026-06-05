// 路径: scripts/verify-indexes.go
// 验证数据库索引是否正确创建
package main

import (
	"fmt"
	"log"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
)

func main() {
	// 加载配置
	if err := config.LoadDefault(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := config.InitDatabase(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer config.CloseDatabase()

	// 执行迁移
	fmt.Println("执行数据库迁移...")
	if err := config.AutoMigrate(
		&model.User{},
		&model.PhysicalFile{},
		&model.UserFile{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	fmt.Println("✅ 数据库迁移成功")

	// 验证索引
	fmt.Println("\n验证 user_files 表索引...")
	
	db := config.GetDB()
	
	// 查询索引信息
	var indexes []struct {
		Table      string `gorm:"column:TABLE_NAME"`
		NonUnique  int    `gorm:"column:NON_UNIQUE"`
		KeyName    string `gorm:"column:INDEX_NAME"`
		SeqInIndex int    `gorm:"column:SEQ_IN_INDEX"`
		ColumnName string `gorm:"column:COLUMN_NAME"`
	}
	
	result := db.Raw(`
		SELECT 
			TABLE_NAME,
			NON_UNIQUE,
			INDEX_NAME as KeyName,
			SEQ_IN_INDEX as SeqInIndex,
			COLUMN_NAME as ColumnName
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = 'user_files'
		ORDER BY INDEX_NAME, SEQ_IN_INDEX
	`).Scan(&indexes)
	
	if result.Error != nil {
		log.Fatalf("查询索引失败: %v", result.Error)
	}

	// 打印索引信息
	fmt.Println("\n当前索引:")
	fmt.Println("----------------------------------------")
	currentIndex := ""
	for _, idx := range indexes {
		if idx.KeyName != currentIndex {
			if currentIndex != "" {
				fmt.Println()
			}
			currentIndex = idx.KeyName
			fmt.Printf("索引名: %s\n", idx.KeyName)
		}
		fmt.Printf("  - 列 %d: %s\n", idx.SeqInIndex, idx.ColumnName)
	}
	fmt.Println("----------------------------------------")

	// 验证关键索引
	fmt.Println("\n验证关键索引:")
	
	requiredIndexes := map[string][]string{
		"idx_user_parent": {"user_id", "parent_id"},
		"idx_user_folder": {"is_folder"},
	}
	
	indexMap := make(map[string][]string)
	for _, idx := range indexes {
		indexMap[idx.KeyName] = append(indexMap[idx.KeyName], idx.ColumnName)
	}
	
	allOK := true
	for indexName, expectedColumns := range requiredIndexes {
		actualColumns, exists := indexMap[indexName]
		if !exists {
			fmt.Printf("❌ 索引 %s 不存在\n", indexName)
			allOK = false
			continue
		}
		
		if len(actualColumns) != len(expectedColumns) {
			fmt.Printf("❌ 索引 %s 列数不匹配: 期望 %d, 实际 %d\n", 
				indexName, len(expectedColumns), len(actualColumns))
			allOK = false
			continue
		}
		
		match := true
		for i, expected := range expectedColumns {
			if actualColumns[i] != expected {
				match = false
				break
			}
		}
		
		if match {
			fmt.Printf("✅ 索引 %s 正确: %v\n", indexName, actualColumns)
		} else {
			fmt.Printf("❌ 索引 %s 列顺序不匹配: 期望 %v, 实际 %v\n", 
				indexName, expectedColumns, actualColumns)
			allOK = false
		}
	}
	
	fmt.Println()
	if allOK {
		fmt.Println("========================================")
		fmt.Println("✅ 所有索引验证通过!")
		fmt.Println("========================================")
	} else {
		fmt.Println("========================================")
		fmt.Println("❌ 索引验证失败,请检查上述错误")
		fmt.Println("========================================")
		log.Fatal("索引验证失败")
	}
}
