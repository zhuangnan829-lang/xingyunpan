package middleware

import (
	"time"

	"xingyunpan-v2/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	RequestIDKey = "request_id"
	UserIDKey    = "user_id"
)

// LoggingMiddleware 结构化日志中间件
// 为每个请求生成 request_id，记录请求信息和响应信息
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成 request_id
		requestID := uuid.New().String()
		c.Set(RequestIDKey, requestID)

		// 记录请求开始时间
		start := time.Now()

		// 记录请求开始
		logger.Info("Request started",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// 处理请求
		c.Next()

		// 计算请求耗时
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		// 获取 user_id（如果已认证）
		var userID uint
		if uid, exists := c.Get(UserIDKey); exists {
			if id, ok := uid.(uint); ok {
				userID = id
			}
		}

		// 构建请求上下文
		reqCtx := &logger.RequestContext{
			RequestID: requestID,
			UserID:    userID,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Status:    status,
			Duration:  duration,
		}

		// 检查是否触发限流
		if rateLimitExceeded, exists := c.Get("rate_limit_exceeded"); exists && rateLimitExceeded.(bool) {
			limitType := "unknown"
			if lt, exists := c.Get("rate_limit_type"); exists {
				limitType = lt.(string)
			}
			logger.WarnWithContext(reqCtx, "Rate limit exceeded",
				zap.String("limit_type", limitType),
				zap.String("client_ip", c.ClientIP()),
			)
			return
		}

		// 根据状态码选择日志级别
		if status >= 500 {
			// 服务器错误：记录错误日志
			logger.ErrorWithContext(reqCtx, "Request completed with server error",
				zap.String("error", c.Errors.String()),
			)
		} else if status >= 400 {
			// 客户端错误：记录警告日志
			logger.WarnWithContext(reqCtx, "Request completed with client error")
		} else {
			// 成功：记录信息日志
			logger.InfoWithContext(reqCtx, "Request completed")
		}
	}
}

// GetRequestID 从上下文中获取 request_id
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUserID 从上下文中获取 user_id
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(UserIDKey); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}
