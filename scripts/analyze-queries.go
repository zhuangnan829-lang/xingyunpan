// 路径: scripts/analyze-queries.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
)

type QueryAnalysis struct {
	Name        string
	Query       string
	ExpectedIdx string
}

func main() {
	// 从环境变量读取数据库配置
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:123456@tcp(localhost:3306)/xingyunpan?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("Phase 5 查询性能分析")
	fmt.Println("========================================")
	fmt.Println()

	// 定义要分析的查询
	queries := []QueryAnalysis{
		{
			Name:        "搜索 - 简单关键词",
			Query:       "SELECT * FROM user_files WHERE user_id = 1 AND deleted_at IS NULL AND file_name LIKE '%test%' ORDER BY modified_at DESC LIMIT 20",
			ExpectedIdx: "idx_file_name",
		},
		{
			Name:        "搜索 - 多条件过滤",
			Query:       "SELECT * FROM user_files WHERE user_id = 1 AND deleted_at IS NULL AND file_name LIKE '%doc%' AND file_type = 'pdf' AND file_size >= 1024 ORDER BY modified_at DESC LIMIT 20",
			ExpectedIdx: "idx_file_name",
		},
		{
			Name:        "版本历史 - 获取所有版本",
			Query:       "SELECT * FROM file_versions WHERE file_id = 'file_123' ORDER BY version_number DESC",
			ExpectedIdx: "idx_file_versions",
		},
		{
			Name:        "版本历史 - 获取当前版本",
			Query:       "SELECT * FROM file_versions WHERE file_id = 'file_123' AND is_current = true LIMIT 1",
			ExpectedIdx: "idx_current_version",
		},
		{
			Name:        "回收站 - 用户列表",
			Query:       "SELECT * FROM recycle_bin WHERE user_id = 1 ORDER BY deleted_at DESC LIMIT 20",
			ExpectedIdx: "idx_user_expires",
		},
		{
			Name:        "回收站 - 过期项目",
			Query:       fmt.Sprintf("SELECT * FROM recycle_bin WHERE expires_at < '%s' LIMIT 100", time.Now().Format("2006-01-02 15:04:05")),
			ExpectedIdx: "idx_user_expires",
		},
		{
			Name:        "分享 - Token 查询",
			Query:       "SELECT * FROM shares WHERE share_token = 'abc123' LIMIT 1",
			ExpectedIdx: "idx_share_token",
		},
		{
			Name:        "协作 - 文件协作者",
			Query:       "SELECT * FROM collaborations WHERE file_id = 'file_123'",
			ExpectedIdx: "idx_file_owner",
		},
	}

	// 分析每个查询
	allPassed := true
	for i, qa := range queries {
		fmt.Printf("%d. %s\n", i+1, qa.Name)
		fmt.Println("----------------------------------------")

		passed := analyzeQuery(db, qa)
		if !passed {
			allPassed = false
		}

		fmt.Println()
	}

	// 显示索引使用情况
	fmt.Println("========================================")
	fmt.Println("索引使用情况")
	fmt.Println("========================================")
	showIndexUsage(db)

	fmt.Println()
	if allPassed {
		fmt.Println("✓ 所有查询都正确使用了索引")
	} else {
		fmt.Println("✗ 某些查询未使用预期索引，需要优化")
		os.Exit(1)
	}
}

func analyzeQuery(db *sql.DB, qa QueryAnalysis) bool {
	// 执行 EXPLAIN
	explainQuery := "EXPLAIN " + qa.Query
	rows, err := db.Query(explainQuery)
	if err != nil {
		log.Printf("EXPLAIN 失败: %v", err)
		return false
	}
	defer rows.Close()

	// 解析 EXPLAIN 结果
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(columns)

	usesIndex := false
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("扫描结果失败: %v", err)
			continue
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val != nil {
				row[i] = fmt.Sprintf("%v", val)

				// 检查是否使用了预期索引
				if columns[i] == "key" && val != nil {
					keyStr := fmt.Sprintf("%v", val)
					if keyStr == qa.ExpectedIdx {
						usesIndex = true
					}
				}
			} else {
				row[i] = "NULL"
			}
		}
		table.Append(row)
	}

	table.Render()

	if usesIndex {
		fmt.Printf("✓ 使用了预期索引: %s\n", qa.ExpectedIdx)
		return true
	} else {
		fmt.Printf("✗ 未使用预期索引: %s\n", qa.ExpectedIdx)
		return false
	}
}

func showIndexUsage(db *sql.DB) {
	query := `
		SELECT 
			TABLE_NAME,
			INDEX_NAME,
			GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) as columns,
			CARDINALITY
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME IN ('user_files', 'shares', 'share_files', 'recycle_bin', 'file_versions', 'collaborations')
		GROUP BY TABLE_NAME, INDEX_NAME
		ORDER BY TABLE_NAME, INDEX_NAME
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询索引失败: %v", err)
		return
	}
	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"表名", "索引名", "列", "基数"})

	for rows.Next() {
		var tableName, indexName, columns string
		var cardinality sql.NullInt64

		if err := rows.Scan(&tableName, &indexName, &columns, &cardinality); err != nil {
			log.Printf("扫描结果失败: %v", err)
			continue
		}

		cardStr := "NULL"
		if cardinality.Valid {
			cardStr = fmt.Sprintf("%d", cardinality.Int64)
		}

		table.Append([]string{tableName, indexName, columns, cardStr})
	}

	table.Render()
}
