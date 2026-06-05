// 路径: internal/controller/folder_controller.go
package controller

import (
	"net/http"
	"strconv"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FolderController 文件夹控制器
type FolderController struct {
	folderService  service.FolderService
	recycleService service.RecycleService
}

// NewFolderController 创建文件夹控制器实例
func NewFolderController(folderService service.FolderService, recycleService service.RecycleService) *FolderController {
	return &FolderController{
		folderService:  folderService,
		recycleService: recycleService,
	}
}

// Create 创建文件夹接口
func (c *FolderController) Create(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var req struct {
		Name     string `json:"name" binding:"required"`
		ParentID *uint  `json:"parent_id"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	folder, err := c.folderService.Create(userID.(uint), req.Name, req.ParentID)
	if err != nil {
		// ✅ 任务 6: 记录详细错误到日志
		logger.Error("创建文件夹失败",
			zap.Uint("user_id", userID.(uint)),
			zap.String("name", req.Name),
			zap.Error(err))
		// 区分业务错误和系统错误
		if err.Error() == "无权访问此文件夹" || err.Error() == "文件夹不存在" || err.Error() == "不是文件夹" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "创建失败")
		}
		return
	}

	response.SuccessWithMessage(ctx, "创建成功", folder)
}

// Rename 重命名文件夹接口
func (c *FolderController) Rename(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	folderID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	if err := c.folderService.Rename(userID.(uint), uint(folderID), req.Name); err != nil {
		// ✅ 任务 6: 记录详细错误到日志
		logger.Error("重命名文件夹失败",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("folder_id", folderID),
			zap.String("new_name", req.Name),
			zap.Error(err))
		// 区分业务错误和系统错误
		if err.Error() == "无权操作此文件夹" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "重命名失败")
		}
		return
	}

	response.SuccessWithMessage(ctx, "重命名成功", nil)
}

// Delete 删除文件夹接口
func (c *FolderController) Delete(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	folderID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var err error
	if c.recycleService != nil {
		err = c.recycleService.MoveToRecycle(ctx, userID.(uint), []uint{uint(folderID)})
	} else {
		err = c.folderService.Delete(userID.(uint), uint(folderID))
	}
	if err != nil {
		// ✅ 任务 6: 记录详细错误到日志
		logger.Error("删除文件夹失败",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("folder_id", folderID),
			zap.Error(err))
		// 区分业务错误和系统错误
		if err.Error() == "无权删除此文件夹" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "删除失败")
		}
		return
	}

	response.SuccessWithMessage(ctx, "删除成功", nil)
}

// Move 移动文件夹接口
func (c *FolderController) Move(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	folderID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var req struct {
		ParentID *uint `json:"parent_id"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	if err := c.folderService.Move(userID.(uint), uint(folderID), req.ParentID); err != nil {
		// ✅ 任务 6: 记录详细错误到日志
		logger.Error("移动文件夹失败",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("folder_id", folderID),
			zap.Error(err))
		// 区分业务错误和系统错误
		if err.Error() == "无权移动此文件夹" || err.Error() == "无权访问此文件夹" ||
			err.Error() == "不能将文件夹移动到自己下面" || err.Error() == "不能移动到子文件夹" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "移动失败")
		}
		return
	}

	response.SuccessWithMessage(ctx, "移动成功", nil)
}

// Copy 复制文件夹接口
func (c *FolderController) Copy(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	folderID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var req struct {
		ParentID *uint `json:"parent_id"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	folder, err := c.folderService.Copy(userID.(uint), uint(folderID), req.ParentID)
	if err != nil {
		logger.Error("复制文件夹失败",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("folder_id", folderID),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "复制成功", folder)
}

// Path 获取文件夹路径接口
func (c *FolderController) Path(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	folderID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	path, err := c.folderService.Path(userID.(uint), uint(folderID))
	if err != nil {
		logger.Error("获取文件夹路径失败",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("folder_id", folderID),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, path)
}
