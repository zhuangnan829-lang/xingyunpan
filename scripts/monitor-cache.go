// 路径: scripts/monitor-cache.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/olekukonko/tablewriter"
)

func main() {
	// 从环境变量读取 Redis 配置
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer client.Close()

	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("Redis 缓存监控")
	fmt.Println("========================================")
	fmt.Println()

	// 1. 显示缓存统计信息
	showCacheStats(ctx, client)

	// 2. 显示缓存键列表
	showCacheKeys(ctx, client)

	// 3. 显示缓存命中率
	showCacheHitRate(ctx, client)

	// 4. 显示内存使用情况
	showMemoryUsage(ctx, client)
}

func showCacheStats(ctx context.Context, client *redis.Client) {
	fmt.Println("1. 缓存统计信息")
	fmt.Println("----------------------------------------")

	info, err := client.Info(ctx, "stats").Result()
	if err != nil {
		log.Printf("获取统计信息失败: %v", err)
		return
	}

	fmt.Println(info)
	fmt.Println()
}

func showCacheKeys(ctx context.Context, client *redis.Client) {
	fmt.Println("2. 缓存键列表")
	fmt.Println("----------------------------------------")

	// 获取不同类型的缓存键
	patterns := []string{
		"search:*",
		"versions:*",
		"collaborators:*",
		"share_verify:*",
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"缓存类型", "键数量", "示例键", "TTL"})

	for _, pattern := range patterns {
		keys, err := client.Keys(ctx, pattern).Result()
		if err != nil {
			log.Printf("获取键失败: %v", err)
			continue
		}

		cacheType := pattern[:len(pattern)-2] // 移除 :*
		count := len(keys)

		exampleKey := "无"
		ttl := "N/A"
		if count > 0 {
			exampleKey = keys[0]
			ttlDuration, _ := client.TTL(ctx, exampleKey).Result()
			ttl = ttlDuration.String()
		}

		table.Append([]string{
			cacheType,
			fmt.Sprintf("%d", count),
			exampleKey,
			ttl,
		})
	}

	table.Render()
	fmt.Println()
}

func showCacheHitRate(ctx context.Context, client *redis.Client) {
	fmt.Println("3. 缓存命中率")
	fmt.Println("----------------------------------------")

	info, err := client.Info(ctx, "stats").Result()
	if err != nil {
		log.Printf("获取统计信息失败: %v", err)
		return
	}

	// 解析 keyspace_hits 和 keyspace_misses
	var hits, misses int64
	fmt.Sscanf(info, "keyspace_hits:%d\nkeyspace_misses:%d", &hits, &misses)

	total := hits + misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"指标", "值"})
	table.Append([]string{"缓存命中次数", fmt.Sprintf("%d", hits)})
	table.Append([]string{"缓存未命中次数", fmt.Sprintf("%d", misses)})
	table.Append([]string{"总请求次数", fmt.Sprintf("%d", total)})
	table.Append([]string{"缓存命中率", fmt.Sprintf("%.2f%%", hitRate)})
	table.Render()

	fmt.Println()

	// 评估缓存命中率
	if hitRate >= 80 {
		fmt.Println("✓ 缓存命中率良好 (≥ 80%)")
	} else if hitRate >= 60 {
		fmt.Println("⚠ 缓存命中率一般 (60-80%)，建议优化")
	} else {
		fmt.Println("✗ 缓存命中率较低 (< 60%)，需要优化")
	}

	fmt.Println()
}

func showMemoryUsage(ctx context.Context, client *redis.Client) {
	fmt.Println("4. 内存使用情况")
	fmt.Println("----------------------------------------")

	info, err := client.Info(ctx, "memory").Result()
	if err != nil {
		log.Printf("获取内存信息失败: %v", err)
		return
	}

	// 解析内存使用信息
	var usedMemory, maxMemory int64
	fmt.Sscanf(info, "used_memory:%d", &usedMemory)
	fmt.Sscanf(info, "maxmemory:%d", &maxMemory)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"指标", "值"})
	table.Append([]string{"已使用内存", formatBytes(usedMemory)})

	if maxMemory > 0 {
		table.Append([]string{"最大内存", formatBytes(maxMemory)})
		usage := float64(usedMemory) / float64(maxMemory) * 100
		table.Append([]string{"内存使用率", fmt.Sprintf("%.2f%%", usage)})
	} else {
		table.Append([]string{"最大内存", "未限制"})
	}

	table.Render()

	fmt.Println()
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
