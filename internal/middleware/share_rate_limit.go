// 路径: internal/middleware/share_rate_limit.go
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"xingyunpan-v2/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// SharePasswordRateLimiter 分享密码验证限流器
type SharePasswordRateLimiter struct {
	redis      *redis.Client
	maxRetries int           // 最大失败次数
	window     time.Duration // 时间窗口
}

// NewSharePasswordRateLimiter 创建分享密码验证限流器
// maxRetries: 最大失败次数（默认 5 次）
// window: 时间窗口（默认 15 分钟）
func NewSharePasswordRateLimiter(redis *redis.Client, maxRetries int, window time.Duration) *SharePasswordRateLimiter {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	if window <= 0 {
		window = 15 * time.Minute
	}

	return &SharePasswordRateLimiter{
		redis:      redis,
		maxRetries: maxRetries,
		window:     window,
	}
}

// getKey 生成 Redis key
// 格式: share_verify:{ip}:{shareID}
func (s *SharePasswordRateLimiter) getKey(ip string, shareID string) string {
	return fmt.Sprintf("share_verify:%s:%s", ip, shareID)
}

// CheckLimit 检查是否超过限流
// 返回: (是否允许, 剩余次数, 错误)
func (s *SharePasswordRateLimiter) CheckLimit(ctx context.Context, ip string, shareID string) (bool, int, error) {
	key := s.getKey(ip, shareID)

	// 获取当前失败次数
	count, err := s.redis.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, 0, fmt.Errorf("获取限流计数失败: %w", err)
	}

	// 如果超过最大次数，拒绝请求
	if count >= s.maxRetries {
		return false, 0, nil
	}

	// 允许请求，返回剩余次数
	remaining := s.maxRetries - count
	return true, remaining, nil
}

// RecordFailure 记录失败尝试
func (s *SharePasswordRateLimiter) RecordFailure(ctx context.Context, ip string, shareID string) error {
	key := s.getKey(ip, shareID)

	// 增加失败计数
	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, s.window)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("记录失败尝试失败: %w", err)
	}

	return nil
}

// ResetLimit 重置限流（密码验证成功后调用）
func (s *SharePasswordRateLimiter) ResetLimit(ctx context.Context, ip string, shareID string) error {
	key := s.getKey(ip, shareID)
	return s.redis.Del(ctx, key).Err()
}

// GetFailureCount 获取当前失败次数
func (s *SharePasswordRateLimiter) GetFailureCount(ctx context.Context, ip string, shareID string) (int, error) {
	key := s.getKey(ip, shareID)
	count, err := s.redis.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return 0, fmt.Errorf("获取失败次数失败: %w", err)
	}
	if err == redis.Nil {
		return 0, nil
	}
	return count, nil
}

// GetRemainingTime 获取限流剩余时间
func (s *SharePasswordRateLimiter) GetRemainingTime(ctx context.Context, ip string, shareID string) (time.Duration, error) {
	key := s.getKey(ip, shareID)
	ttl, err := s.redis.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取剩余时间失败: %w", err)
	}
	if ttl < 0 {
		return 0, nil
	}
	return ttl, nil
}

// SharePasswordRateLimitMiddleware 分享密码验证限流中间件
// 用于 POST /api/shares/:shareId/verify 端点
func SharePasswordRateLimitMiddleware() gin.HandlerFunc {
	// 创建限流器实例（5 次失败 / 15 分钟）
	limiter := NewSharePasswordRateLimiter(config.RDB, 5, 15*time.Minute)

	return func(c *gin.Context) {
		// 获取 IP 和 shareID
		ip := c.ClientIP()
		shareID := c.Param("shareId")

		if shareID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_REQUEST",
				"message": "分享 ID 不能为空",
			})
			c.Abort()
			return
		}

		// 检查限流
		ctx := context.Background()
		allowed, remaining, err := limiter.CheckLimit(ctx, ip, shareID)
		if err != nil {
			// 限流检查失败，记录错误但允许请求继续（降级策略）
			c.Set("rate_limit_error", err.Error())
			c.Next()
			return
		}

		if !allowed {
			// 获取剩余时间
			remainingTime, _ := limiter.GetRemainingTime(ctx, ip, shareID)
			retryAfter := int(remainingTime.Seconds())
			if retryAfter <= 0 {
				retryAfter = 60 // 默认 60 秒
			}

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "TOO_MANY_ATTEMPTS",
				"message": fmt.Sprintf("密码验证失败次数过多，请在 %d 秒后重试", retryAfter),
				"details": gin.H{
					"retry_after": retryAfter,
				},
			})
			c.Abort()
			return
		}

		// 将限流器和剩余次数存入上下文，供 Controller 使用
		c.Set("share_rate_limiter", limiter)
		c.Set("remaining_attempts", remaining)

		c.Next()
	}
}

// SharePasswordRateLimitMiddlewareWithConfig 自定义配置的分享密码验证限流中间件
func SharePasswordRateLimitMiddlewareWithConfig(maxRetries int, window time.Duration) gin.HandlerFunc {
	limiter := NewSharePasswordRateLimiter(config.RDB, maxRetries, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		shareID := c.Param("shareId")

		if shareID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "INVALID_REQUEST",
				"message": "分享 ID 不能为空",
			})
			c.Abort()
			return
		}

		ctx := context.Background()
		allowed, remaining, err := limiter.CheckLimit(ctx, ip, shareID)
		if err != nil {
			c.Set("rate_limit_error", err.Error())
			c.Next()
			return
		}

		if !allowed {
			remainingTime, _ := limiter.GetRemainingTime(ctx, ip, shareID)
			retryAfter := int(remainingTime.Seconds())
			if retryAfter <= 0 {
				retryAfter = 60
			}

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "TOO_MANY_ATTEMPTS",
				"message": fmt.Sprintf("密码验证失败次数过多，请在 %d 秒后重试", retryAfter),
				"details": gin.H{
					"retry_after": retryAfter,
				},
			})
			c.Abort()
			return
		}

		c.Set("share_rate_limiter", limiter)
		c.Set("remaining_attempts", remaining)

		c.Next()
	}
}
