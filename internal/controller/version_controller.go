// 路径: internal/controller/version_controller.go
package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

// VersionController 版本管理控制器
type VersionController struct {
	versionService service.VersionService
}

// NewVersionController 创建版本管理控制器实例
func NewVersionController(versionService service.VersionService) *VersionController {
	return &VersionController{
		versionService: versionService,
	}
}

// GetVersionHistory 获取版本历史接口
// @Summary 获取版本历史
// @Tags 版本管理
// @Produce json
// @Security Bearer
// @Param fileId path int true "文件ID"
// @Success 200 {object} response.Response
// @Router /api/files/{fileId}/versions [get]
func (c *VersionController) GetVersionHistory(ctx *gin.Context) {
	// 获取文件 ID
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "文件ID格式错误")
		return
	}

	// 从上下文获取用户 ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	// 获取版本历史
	versions, err := c.versionService.GetVersionHistory(ctx, userID.(uint), uint(fileID))
	if err != nil {
		if err.Error() == "文件不存在" {
			response.Error(ctx, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "无权限访问此文件" {
			response.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, versions)
}

// DownloadVersion 下载版本接口
// @Summary 下载版本
// @Tags 版本管理
// @Produce octet-stream
// @Security Bearer
// @Param fileId path int true "文件ID"
// @Param versionId path int true "版本ID"
// @Success 200 {file} binary
// @Router /api/files/{fileId}/versions/{versionId}/download [get]
func (c *VersionController) DownloadVersion(ctx *gin.Context) {
	// 获取文件 ID
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "文件ID格式错误")
		return
	}

	// 获取版本 ID
	versionIDStr := ctx.Param("versionId")
	versionID, err := strconv.ParseUint(versionIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "版本ID格式错误")
		return
	}

	// 从上下文获取用户 ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	// 获取物理文件路径
	filePath, err := c.versionService.DownloadVersion(ctx, userID.(uint), uint(fileID), uint(versionID))
	if err != nil {
		if err.Error() == "文件不存在" || err.Error() == "版本不属于该文件" {
			response.Error(ctx, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "无权限下载此文件" {
			response.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 设置响应头
	filename := filepath.Base(filePath)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	ctx.File(filePath)
}

// RestoreVersion 恢复版本接口
// @Summary 恢复版本
// @Tags 版本管理
// @Produce json
// @Security Bearer
// @Param fileId path int true "文件ID"
// @Param versionId path int true "版本ID"
// @Success 200 {object} response.Response
// @Router /api/files/{fileId}/versions/{versionId}/restore [post]
func (c *VersionController) RestoreVersion(ctx *gin.Context) {
	// 获取文件 ID
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "文件ID格式错误")
		return
	}

	// 获取版本 ID
	versionIDStr := ctx.Param("versionId")
	versionID, err := strconv.ParseUint(versionIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "版本ID格式错误")
		return
	}

	// 从上下文获取用户 ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	// 恢复版本
	resp, err := c.versionService.RestoreVersion(ctx, userID.(uint), uint(fileID), uint(versionID))
	if err != nil {
		if err.Error() == "文件不存在" || err.Error() == "版本不属于该文件" {
			response.Error(ctx, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "无权限恢复此文件版本" {
			response.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		if err.Error() == "存储空间不足" {
			ctx.JSON(http.StatusInsufficientStorage, gin.H{
				"code":    507,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, resp)
}

// DeleteVersion 删除版本接口
// @Summary 删除版本
// @Tags 版本管理
// @Produce json
// @Security Bearer
// @Param fileId path int true "文件ID"
// @Param versionId path int true "版本ID"
// @Success 204 "删除成功"
// @Router /api/files/{fileId}/versions/{versionId} [delete]
func (c *VersionController) DeleteVersion(ctx *gin.Context) {
	// 获取文件 ID
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "文件ID格式错误")
		return
	}

	// 获取版本 ID
	versionIDStr := ctx.Param("versionId")
	versionID, err := strconv.ParseUint(versionIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "版本ID格式错误")
		return
	}

	// 从上下文获取用户 ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	// 删除版本
	err = c.versionService.DeleteVersion(ctx, userID.(uint), uint(fileID), uint(versionID))
	if err != nil {
		if err.Error() == "文件不存在" || err.Error() == "版本不属于该文件" {
			response.Error(ctx, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "只有文件所有者可以删除版本" {
			response.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		if err.Error() == "不能删除当前版本" {
			response.Error(ctx, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}
