// 路径: pkg/redis/lock.go
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Lock Redis 分布式锁
type Lock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

// NewLock 创建分布式锁实例
func NewLock(client *redis.Client, key string, ttl time.Duration) *Lock {
	return &Lock{
		client: client,
		key:    fmt.Sprintf("lock:%s", key),
		value:  fmt.Sprintf("%d", time.Now().UnixNano()),
		ttl:    ttl,
	}
}

// Acquire 获取锁
func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	return l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
}

// Release 释放锁
func (l *Lock) Release(ctx context.Context) error {
	// 使用 Lua 脚本确保只删除自己持有的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return l.client.Eval(ctx, script, []string{l.key}, l.value).Err()
}

// Extend 延长锁的过期时间
func (l *Lock) Extend(ctx context.Context) error {
	return l.client.Expire(ctx, l.key, l.ttl).Err()
}
