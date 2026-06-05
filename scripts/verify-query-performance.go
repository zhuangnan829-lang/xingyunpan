package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type QueryTest struct {
	Name        string
	Query       string
	Description string
}

type QueryResult struct {
	Name        string
	Duration    time.Duration
	RowsScanned int
	UsesIndex   bool
	IndexName   string
}

func main() {
	// Get database connection info from environment or use defaults
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "xingyunpan")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "xingyunpan")

	if dbPassword == "" {
		log.Fatal("Please set DB_PASSWORD environment variable")
	}

	// Connect to database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("=== 查询性能验证报告 ===")
	fmt.Println()
	fmt.Printf("数据库: %s@%s:%s/%s\n", dbUser, dbHost, dbPort, dbName)
	fmt.Printf("测试时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Define test queries
	tests := []QueryTest{
		{
			Name:        "文件列表查询",
			Query:       "SELECT * FROM user_files WHERE user_id = 1 AND folder_id = 0 AND deleted_at IS NULL",
			Description: "查询用户的文件列表（最常用的查询）",
		},
		{
			Name:        "文件哈希查询",
			Query:       "SELECT * FROM physical_files WHERE file_hash = 'test_hash'",
			Description: "秒传功能：查找相同哈希的文件",
		},
		{
			Name:        "过期上传任务查询",
			Query:       "SELECT * FROM multipart_uploads WHERE status = 'pending' AND created_at < DATE_SUB(NOW(), INTERVAL 24 HOUR)",
			Description: "清理过期的分片上传任务",
		},
		{
			Name:        "用户上传任务查询",
			Query:       "SELECT * FROM multipart_uploads WHERE user_id = 1 AND status = 'pending'",
			Description: "查询用户的上传任务",
		},
		{
			Name:        "待处理删除任务查询",
			Query:       "SELECT * FROM file_deletions WHERE status = 'pending'",
			Description: "查找待处理的文件删除任务",
		},
		{
			Name:        "用户名查询",
			Query:       "SELECT * FROM users WHERE username = 'testuser'",
			Description: "登录时的用户名查找",
		},
		{
			Name:        "文件关联查询",
			Query:       "SELECT uf.*, pf.* FROM user_files uf LEFT JOIN physical_files pf ON uf.real_file_id = pf.id WHERE uf.user_id = 1 AND uf.folder_id = 0 AND uf.deleted_at IS NULL",
			Description: "查询文件及其物理文件信息",
		},
		{
			Name:        "零引用文件查询",
			Query:       "SELECT * FROM physical_files WHERE ref_count = 0",
			Description: "垃圾回收：查找引用计数为 0 的文件",
		},
	}

	// Run tests
	results := make([]QueryResult, 0, len(tests))
	for i, test := range tests {
		fmt.Printf("[%d/%d] 测试: %s\n", i+1, len(tests), test.Name)
		fmt.Printf("      描述: %s\n", test.Description)

		result := runQueryTest(db, test)
		results = append(results, result)

		fmt.Printf("      执行时间: %v\n", result.Duration)
		fmt.Printf("      扫描行数: %d\n", result.RowsScanned)
		fmt.Printf("      使用索引: %v", result.UsesIndex)
		if result.UsesIndex {
			fmt.Printf(" (%s)", result.IndexName)
		}
		fmt.Println()
		fmt.Println()
	}

	// Generate summary
	fmt.Println("=== 性能总结 ===")
	fmt.Println()

	totalDuration := time.Duration(0)
	indexUsageCount := 0
	for _, result := range results {
		totalDuration += result.Duration
		if result.UsesIndex {
			indexUsageCount++
		}
	}

	fmt.Printf("总测试数: %d\n", len(results))
	fmt.Printf("总执行时间: %v\n", totalDuration)
	fmt.Printf("平均执行时间: %v\n", totalDuration/time.Duration(len(results)))
	fmt.Printf("索引使用率: %.1f%% (%d/%d)\n",
		float64(indexUsageCount)/float64(len(results))*100,
		indexUsageCount, len(results))
	fmt.Println()

	// Performance evaluation
	fmt.Println("=== 性能评估 ===")
	fmt.Println()

	allFast := true
	for _, result := range results {
		// Consider queries under 10ms as fast
		if result.Duration > 10*time.Millisecond {
			allFast = false
			fmt.Printf("⚠️  %s 执行时间较长: %v\n", result.Name, result.Duration)
		}
	}

	if allFast {
		fmt.Println("✅ 所有查询执行时间都在 10ms 以内，性能优秀！")
	}

	if indexUsageCount == len(results) {
		fmt.Println("✅ 所有查询都使用了索引，优化成功！")
	} else {
		fmt.Printf("⚠️  有 %d 个查询未使用索引，可能需要进一步优化\n",
			len(results)-indexUsageCount)
	}

	fmt.Println()
	fmt.Println("=== 优化建议 ===")
	fmt.Println()

	if indexUsageCount < len(results) {
		fmt.Println("1. 检查未使用索引的查询，确认索引已创建")
		fmt.Println("2. 运行 ANALYZE TABLE 更新统计信息")
		fmt.Println("3. 使用 EXPLAIN 分析查询计划")
	}

	if !allFast {
		fmt.Println("1. 考虑添加更多索引或优化现有索引")
		fmt.Println("2. 检查数据量是否过大，考虑分页查询")
		fmt.Println("3. 优化查询语句，只查询需要的列")
	}

	if indexUsageCount == len(results) && allFast {
		fmt.Println("✅ 查询性能已达到最佳状态，无需进一步优化")
	}

	fmt.Println()
	fmt.Println("=== 验证完成 ===")
}

func runQueryTest(db *sql.DB, test QueryTest) QueryResult {
	result := QueryResult{
		Name: test.Name,
	}

	// First, use EXPLAIN to check if index is used
	explainQuery := "EXPLAIN " + test.Query
	var (
		id           sql.NullInt64
		selectType   sql.NullString
		table        sql.NullString
		partitions   sql.NullString
		queryType    sql.NullString
		possibleKeys sql.NullString
		key          sql.NullString
		keyLen       sql.NullString
		ref          sql.NullString
		rows         sql.NullInt64
		filtered     sql.NullFloat64
		extra        sql.NullString
	)

	err := db.QueryRow(explainQuery).Scan(
		&id, &selectType, &table, &partitions, &queryType,
		&possibleKeys, &key, &keyLen, &ref, &rows, &filtered, &extra,
	)

	if err == nil {
		result.RowsScanned = int(rows.Int64)
		if key.Valid && key.String != "" {
			result.UsesIndex = true
			result.IndexName = key.String
		}
	}

	// Measure query execution time
	start := time.Now()
	rows, err := db.Query(test.Query)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return result
	}
	defer rows.Close()

	// Count rows
	rowCount := 0
	for rows.Next() {
		rowCount++
	}
	result.Duration = time.Since(start)

	return result
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
