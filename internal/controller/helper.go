// 路径: internal/controller/helper.go
package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// GetCurrentUserID 从上下文中获取当前用户 ID
// ✅ 任务 9: 封装常用辅助函数
func GetCurrentUserID(ctx *gin.Context) (uint, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return 0, fmt.Errorf("未找到用户 ID")
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, fmt.Errorf("用户 ID 类型错误")
	}

	return id, nil
}

// GetCurrentUsername 从上下文中获取当前用户名
// ✅ 任务 9: 封装常用辅助函数
func GetCurrentUsername(ctx *gin.Context) (string, error) {
	username, exists := ctx.Get("username")
	if !exists {
		return "", fmt.Errorf("未找到用户名")
	}

	name, ok := username.(string)
	if !ok {
		return "", fmt.Errorf("用户名类型错误")
	}

	return name, nil
}
