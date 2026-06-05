// 路径: internal/controller/health_controller.go
package controller

import (
	"net/http"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthController 健康检查控制器
type HealthController struct{}

// NewHealthController 创建健康检查控制器
func NewHealthController() *HealthController {
	return &HealthController{}
}

// Check 健康检查接口
// @Summary 健康检查
// @Description 检查服务、数据库和 Redis 的健康状态
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "服务正常"
// @Failure 503 {object} response.Response "服务异常"
// @Router /health [get]
func (h *HealthController) Check(c *gin.Context) {
	healthStatus := make(map[string]interface{})
	allHealthy := true

	// 检查数据库连接
	if err := config.PingDatabase(); err != nil {
		healthStatus["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		allHealthy = false
	} else {
		healthStatus["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// 检查 Redis 连接
	if err := config.PingRedis(); err != nil {
		healthStatus["redis"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		allHealthy = false
	} else {
		healthStatus["redis"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// 服务状态
	healthStatus["service"] = map[string]interface{}{
		"status": "healthy",
		"name":   "xingyunpan-v2",
	}

	// 返回结果
	if allHealthy {
		response.Success(c, healthStatus)
	} else {
		c.JSON(http.StatusServiceUnavailable, response.Response{
			Code:    http.StatusServiceUnavailable,
			Message: "服务部分组件异常",
			Data:    healthStatus,
		})
	}
}

// Ping 简单的 Ping 接口
// @Summary Ping
// @Description 简单的 Ping 测试
// @Tags 健康检查
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "Pong"
// @Router /ping [get]
func (h *HealthController) Ping(c *gin.Context) {
	response.SuccessWithMessage(c, "pong", nil)
}
