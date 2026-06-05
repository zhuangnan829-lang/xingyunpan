package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 从环境变量读取数据库连接信息
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "xingyunpan")
	dbPassword := getEnv("DB_PASSWORD", "xingyunpan123")
	dbName := getEnv("DB_NAME", "xingyunpan")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	fmt.Println("=== 数据库索引优化 ===")
	fmt.Println()
	fmt.Printf("数据库连接信息:\n")
	fmt.Printf("  Host: %s\n", dbHost)
	fmt.Printf("  Port: %s\n", dbPort)
	fmt.Printf("  User: %s\n", dbUser)
	fmt.Printf("  Database: %s\n", dbName)
	fmt.Println()

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	fmt.Println("[步骤 1/4] 测量优化前的查询性能...")
	measureQueryPerformance(db, "before")

	fmt.Println()
	fmt.Println("[步骤 2/4] 执行索引创建...")
	createIndexes(db)

	fmt.Println()
	fmt.Println("[步骤 3/4] 验证索引创建...")
	verifyIndexes(db)

	fmt.Println()
	fmt.Println("[步骤 4/4] 测量优化后的查询性能...")
	measureQueryPerformance(db, "after")

	fmt.Println()
	fmt.Println("=== 索引优化完成 ===")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func measureQueryPerformance(db *sql.DB, phase string) {
	queries := []string{
		"SELECT * FROM user_files WHERE user_id = 1 AND folder_id = 0 AND deleted_at IS NULL",
		"SELECT * FROM physical_files WHERE file_hash = 'test_hash'",
		"SELECT * FROM multipart_uploads WHERE status = 'pending' AND created_at < DATE_SUB(NOW(), INTERVAL 24 HOUR)",
		"SELECT * FROM multipart_uploads WHERE user_id = 1 AND status = 'pending'",
		"SELECT * FROM file_deletions WHERE status = 'pending'",
		"SELECT * FROM users WHERE username = 'testuser'",
	}

	fmt.Printf("\n查询性能测试 (%s optimization):\n", phase)
	for i, query := range queries {
		start := time.Now()
		rows, err := db.Query(query)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("  Query %d: ERROR - %v\n", i+1, err)
			continue
		}
		rows.Close()

		// 获取 EXPLAIN 结果
		explainQuery := "EXPLAIN " + query
		explainRow := db.QueryRow(explainQuery)
		var id, selectType, table, partitions, typ, possibleKeys, key, keyLen, ref, rowsCount, filtered, extra sql.NullString
		err = explainRow.Scan(&id, &selectType, &table, &partitions, &typ, &possibleKeys, &key, &keyLen, &ref, &rowsCount, &filtered, &extra)
		if err != nil {
			fmt.Printf("  Query %d: Duration=%v, EXPLAIN failed\n", i+1, duration)
		} else {
			fmt.Printf("  Query %d: Duration=%v, Type=%s, Key=%s, Rows=%s\n",
				i+1, duration, typ.String, key.String, rowsCount.String)
		}
	}
}

func createIndexes(db *sql.DB) {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_user_files_user_folder ON user_files(user_id, folder_id, deleted_at)",
		"CREATE INDEX IF NOT EXISTS idx_user_files_real_file ON user_files(real_file_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_files_parent ON user_files(parent_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_files_is_folder ON user_files(is_folder)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_physical_files_hash ON physical_files(file_hash)",
		"CREATE INDEX IF NOT EXISTS idx_physical_files_ref_count ON physical_files(ref_count)",
		"CREATE INDEX IF NOT EXISTS idx_multipart_status_created ON multipart_uploads(status, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_multipart_user_status ON multipart_uploads(user_id, status)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_multipart_upload_id ON multipart_uploads(upload_id)",
		"CREATE INDEX IF NOT EXISTS idx_file_deletions_status_created ON file_deletions(status, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_file_deletions_physical_file ON file_deletions(physical_file_id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username)",
	}

	for i, indexSQL := range indexes {
		_, err := db.Exec(indexSQL)
		if err != nil {
			fmt.Printf("  索引 %d 创建失败: %v\n", i+1, err)
		} else {
			fmt.Printf("  索引 %d 创建成功\n", i+1)
		}
	}

	// 更新表统计信息
	tables := []string{"users", "user_files", "physical_files", "multipart_uploads", "file_deletions"}
	fmt.Println("\n更新表统计信息...")
	for _, table := range tables {
		_, err := db.Exec("ANALYZE TABLE " + table)
		if err != nil {
			fmt.Printf("  %s: 失败 - %v\n", table, err)
		} else {
			fmt.Printf("  %s: 成功\n", table)
		}
	}
}

func verifyIndexes(db *sql.DB) {
	tables := []string{"user_files", "physical_files", "multipart_uploads", "file_deletions", "users"}

	for _, table := range tables {
		fmt.Printf("\n%s 表索引:\n", table)
		rows, err := db.Query("SHOW INDEX FROM " + table)
		if err != nil {
			fmt.Printf("  查询失败: %v\n", err)
			continue
		}

		var indexCount int
		for rows.Next() {
			var table, nonUnique, keyName, seqInIndex, columnName, collation, cardinality, subPart, packed, null, indexType, comment, indexComment, visible, expression sql.NullString
			err := rows.Scan(&table, &nonUnique, &keyName, &seqInIndex, &columnName, &collation, &cardinality, &subPart, &packed, &null, &indexType, &comment, &indexComment, &visible, &expression)
			if err == nil {
				indexCount++
				if indexCount <= 5 {
					fmt.Printf("  - %s (%s)\n", keyName.String, columnName.String)
				}
			}
		}
		rows.Close()
		fmt.Printf("  总计: %d 个索引\n", indexCount)
	}
}
