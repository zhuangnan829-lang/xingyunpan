package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters for different IPs
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new IP-based rate limiter
// r: requests per second (e.g., 100 req/min = 100/60 = 1.67 req/sec)
// b: burst size (maximum number of requests allowed at once)
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// GetLimiter returns the rate limiter for the given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// CleanupOldLimiters removes limiters that haven't been used recently
// This prevents memory leaks from accumulating limiters for IPs that are no longer active
func (i *IPRateLimiter) CleanupOldLimiters() {
	i.mu.Lock()
	defer i.mu.Unlock()

	// In a production system, you would track last access time and remove old entries
	// For simplicity, we'll just clear the map periodically
	if len(i.ips) > 10000 {
		i.ips = make(map[string]*rate.Limiter)
	}
}

// UserRateLimiter manages rate limiters for authenticated users
type UserRateLimiter struct {
	users map[uint]*rate.Limiter
	mu    *sync.RWMutex
	r     rate.Limit
	b     int
}

// NewUserRateLimiter creates a new user-based rate limiter
func NewUserRateLimiter(r rate.Limit, b int) *UserRateLimiter {
	return &UserRateLimiter{
		users: make(map[uint]*rate.Limiter),
		mu:    &sync.RWMutex{},
		r:     r,
		b:     b,
	}
}

// GetLimiter returns the rate limiter for the given user ID
func (u *UserRateLimiter) GetLimiter(userID uint) *rate.Limiter {
	u.mu.Lock()
	defer u.mu.Unlock()

	limiter, exists := u.users[userID]
	if !exists {
		limiter = rate.NewLimiter(u.r, u.b)
		u.users[userID] = limiter
	}

	return limiter
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	IPLimiter   *IPRateLimiter
	UserLimiter *UserRateLimiter
}

// NewRateLimitConfig creates a new rate limit configuration
// IP limit: 100 requests per minute
// User limit: 1000 requests per minute
func NewRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		IPLimiter:   NewIPRateLimiter(rate.Limit(100.0/60.0), 10),    // 100 req/min, burst 10
		UserLimiter: NewUserRateLimiter(rate.Limit(1000.0/60.0), 50), // 1000 req/min, burst 50
	}
}

// RateLimitMiddleware creates a middleware that enforces rate limiting
func RateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc {
	// Start a cleanup goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			config.IPLimiter.CleanupOldLimiters()
		}
	}()

	return func(c *gin.Context) {
		// Check if user is authenticated
		userID, exists := c.Get("user_id")

		var limiter *rate.Limiter
		var limitType string

		if exists && userID != nil {
			// Use user-based rate limiter for authenticated users
			uid, ok := userID.(uint)
			if ok {
				limiter = config.UserLimiter.GetLimiter(uid)
				limitType = "user"
			} else {
				// Fallback to IP-based if user_id is not uint
				limiter = config.IPLimiter.GetLimiter(c.ClientIP())
				limitType = "ip"
			}
		} else {
			// Use IP-based rate limiter for unauthenticated requests
			limiter = config.IPLimiter.GetLimiter(c.ClientIP())
			limitType = "ip"
		}

		if !limiter.Allow() {
			// Rate limit exceeded
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Rate limit exceeded. Please try again later.",
			})

			// Log rate limit violation
			c.Set("rate_limit_exceeded", true)
			c.Set("rate_limit_type", limitType)

			c.Abort()
			return
		}

		c.Next()
	}
}
