// 路径: internal/middleware/admin_middleware.go
package middleware

import (
	"net/http"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware 管理员权限中间件
// 必须在 AuthMiddleware 之后使用
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户 ID（由 AuthMiddleware 设置）
		userID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "未授权")
			c.Abort()
			return
		}

		// 查询用户信息
		db := config.GetDB()
		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			response.Error(c, http.StatusUnauthorized, "用户不存在")
			c.Abort()
			return
		}

		if !user.Enabled {
			response.Error(c, http.StatusForbidden, "当前账号已被禁用")
			c.Abort()
			return
		}

		// 检查用户角色
		if user.Role != "admin" {
			response.Error(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}

		// 将用户角色存入上下文
		c.Set("user_role", user.Role)
		c.Next()
	}
}
