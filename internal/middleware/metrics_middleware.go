package middleware

import (
	"strconv"
	"time"

	"xingyunpan-v2/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware 记录 HTTP 请求指标
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// 更新请求计数器
		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		// 更新请求延迟直方图
		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)

		// 记录 API 请求分布
		metrics.RecordApiRequest(path)
	}
}
