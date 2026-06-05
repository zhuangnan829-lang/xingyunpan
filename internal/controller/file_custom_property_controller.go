package controller

import (
	"strconv"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FileCustomPropertyController handles per-file custom property endpoints.
type FileCustomPropertyController struct {
	service service.FileCustomPropertyService
}

// NewFileCustomPropertyController creates a file custom property controller.
func NewFileCustomPropertyController(service service.FileCustomPropertyService) *FileCustomPropertyController {
	return &FileCustomPropertyController{service: service}
}

// Get returns custom property definitions and values for one file.
func (c *FileCustomPropertyController) Get(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	data, err := c.service.Get(userID.(uint), uint(fileID))
	if err != nil {
		logger.Warn("get file custom properties failed", zap.Uint("user_id", userID.(uint)), zap.Uint64("file_id", fileID), zap.Error(err))
		response.Error(ctx, response.CodeInvalidParams, err.Error())
		return
	}

	response.Success(ctx, data)
}

// Update saves custom property values for one file.
func (c *FileCustomPropertyController) Update(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var req service.UpdateFileCustomPropertyPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.CodeInvalidParams, "参数错误")
		return
	}

	data, err := c.service.Update(userID.(uint), uint(fileID), &req)
	if err != nil {
		logger.Warn("update file custom properties failed", zap.Uint("user_id", userID.(uint)), zap.Uint64("file_id", fileID), zap.Error(err))
		response.Error(ctx, response.CodeInvalidParams, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "文件自定义属性已保存", data)
}
