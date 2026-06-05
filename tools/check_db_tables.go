package main

import (
	"fmt"
	"sort"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "xingyunpan:xingyunpan123@tcp(127.0.0.1:3306)/xingyunpan?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	var tables []string
	if err := db.Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		panic(err)
	}

	sort.Strings(tables)
	fmt.Printf("tables: %d\n", len(tables))
	for _, table := range tables {
		fmt.Println(table)
	}

	fmt.Println("----")

	names := []string{
		"user_files",
		"physical_files",
		"shares",
		"share_files",
		"file_versions",
		"multipart_uploads",
	}

	for _, name := range names {
		var count int64
		if err := db.Raw(
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?",
			name,
		).Scan(&count).Error; err != nil {
			panic(err)
		}
		fmt.Printf("%s=%d\n", name, count)
	}

	fmt.Println("---- details")
	type row struct {
		Name string
	}
	var details []row
	if err := db.Raw(`
		SELECT table_name AS name
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		  AND table_name IN ('user_files', 'physical_files', 'shares', 'share_files', 'file_versions', 'multipart_uploads', 'collaborations', 'recycle_bin', 'file_deletions')
		ORDER BY table_name
	`).Scan(&details).Error; err != nil {
		panic(err)
	}
	for _, item := range details {
		fmt.Println(item.Name)
	}
}
