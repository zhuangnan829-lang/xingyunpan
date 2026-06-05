// 路径: internal/config/redis.go
package config

import (
	"context"
	"fmt"
	"time"

	"xingyunpan-v2/pkg/logger"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RDB 全局 Redis 客户端
var RDB *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis() error {
	// 创建 Redis 客户端
	RDB = redis.NewClient(&redis.Options{
		Addr:         Config.Redis.GetAddr(),
		Password:     Config.Redis.Password,
		DB:           Config.Redis.DB,
		PoolSize:     Config.Redis.PoolSize,
		MinIdleConns: Config.Redis.MinIdleConns,
		MaxRetries:   Config.Redis.MaxRetries,
		DialTimeout:  Config.Redis.GetDialTimeout(),
		ReadTimeout:  Config.Redis.GetReadTimeout(),
		WriteTimeout: Config.Redis.GetWriteTimeout(),
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 连接测试失败: %w", err)
	}

	logger.Info("Redis 连接成功",
		zap.String("addr", Config.Redis.GetAddr()),
		zap.Int("db", Config.Redis.DB),
		zap.Int("pool_size", Config.Redis.PoolSize),
		zap.Int("min_idle_conns", Config.Redis.MinIdleConns),
	)

	return nil
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if RDB == nil {
		return nil
	}

	if err := RDB.Close(); err != nil {
		return fmt.Errorf("关闭 Redis 连接失败: %w", err)
	}

	logger.Info("Redis 连接已关闭")
	return nil
}

// GetRedis 获取 Redis 客户端（用于测试或特殊场景）
func GetRedis() *redis.Client {
	return RDB
}

// PingRedis 测试 Redis 连接
func PingRedis() error {
	if RDB == nil {
		return fmt.Errorf("Redis 未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 连接测试失败: %w", err)
	}

	return nil
}

// GetRedisStats 获取 Redis 连接池状态
func GetRedisStats() *redis.PoolStats {
	if RDB == nil {
		return nil
	}
	return RDB.PoolStats()
}
