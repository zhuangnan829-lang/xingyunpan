// 路径: internal/controller/collaboration_controller.go
package controller

import (
	"net/http"
	"strconv"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

// CollaborationController 协作控制器
type CollaborationController struct {
	collaborationService service.CollaborationService
}

// NewCollaborationController 创建协作控制器实例
func NewCollaborationController(collaborationService service.CollaborationService) *CollaborationController {
	return &CollaborationController{
		collaborationService: collaborationService,
	}
}

// AddCollaborator 添加协作者
// @Summary 添加协作者
// @Tags 协作
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AddCollaboratorRequest true "添加协作者请求"
// @Success 201 {object} response.Response
// @Router /api/collaborations [post]
func (c *CollaborationController) AddCollaborator(ctx *gin.Context) {
	var req struct {
		FileID     uint   `json:"file_id" binding:"required"`
		Username   string `json:"username" binding:"required"`
		Permission string `json:"permission" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	result, err := c.collaborationService.AddCollaborator(ctx, userID.(uint), req.FileID, req.Username, req.Permission)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "文件不存在":
			response.Error(ctx, http.StatusNotFound, errMsg)
		case "无权限管理此文件的协作者":
			response.Error(ctx, http.StatusForbidden, errMsg)
		case "该用户已经是协作者":
			response.Error(ctx, http.StatusConflict, errMsg)
		case "用户不存在: " + req.Username:
			response.Error(ctx, http.StatusNotFound, "用户不存在")
		default:
			response.Error(ctx, http.StatusBadRequest, errMsg)
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "添加成功",
		"data":    result,
	})
}

// GetCollaborators 获取协作者列表
// @Summary 获取协作者列表
// @Tags 协作
// @Produce json
// @Security Bearer
// @Param fileId path int true "文件ID"
// @Success 200 {object} response.Response
// @Router /api/files/{fileId}/collaborators [get]
func (c *CollaborationController) GetCollaborators(ctx *gin.Context) {
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	result, err := c.collaborationService.GetCollaborators(ctx, userID.(uint), uint(fileID))
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "文件不存在":
			response.Error(ctx, http.StatusNotFound, errMsg)
		case "无权限查看此文件的协作者":
			response.Error(ctx, http.StatusForbidden, errMsg)
		default:
			response.Error(ctx, http.StatusInternalServerError, errMsg)
		}
		return
	}

	response.Success(ctx, result)
}

// UpdateCollaboratorPermission 更新协作者权限
// @Summary 更新协作者权限
// @Tags 协作
// @Accept json
// @Produce json
// @Security Bearer
// @Param fileId path int true "文件ID"
// @Param userId path int true "协作者用户ID"
// @Param request body UpdatePermissionRequest true "更新权限请求"
// @Success 204 "更新成功"
// @Router /api/files/{fileId}/collaborators/{userId} [put]
func (c *CollaborationController) UpdateCollaboratorPermission(ctx *gin.Context) {
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	collaboratorIDStr := ctx.Param("userId")
	collaboratorID, err := strconv.ParseUint(collaboratorIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的用户 ID")
		return
	}

	var req struct {
		Permission string `json:"permission" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	if err := c.collaborationService.UpdateCollaboratorPermission(ctx, userID.(uint), uint(fileID), uint(collaboratorID), req.Permission); err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "文件不存在":
			response.Error(ctx, http.StatusNotFound, errMsg)
		case "无权限管理此文件的协作者":
			response.Error(ctx, http.StatusForbidden, errMsg)
		case "协作关系不存在":
			response.Error(ctx, http.StatusNotFound, errMsg)
		default:
			response.Error(ctx, http.StatusBadRequest, errMsg)
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

// RemoveCollaborator 移除协作者
// @Summary 移除协作者
// @Tags 协作
// @Produce json
// @Security Bearer
// @Param fileId path int true "文件ID"
// @Param userId path int true "协作者用户ID"
// @Success 204 "移除成功"
// @Router /api/files/{fileId}/collaborators/{userId} [delete]
func (c *CollaborationController) RemoveCollaborator(ctx *gin.Context) {
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	collaboratorIDStr := ctx.Param("userId")
	collaboratorID, err := strconv.ParseUint(collaboratorIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的用户 ID")
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	if err := c.collaborationService.RemoveCollaborator(ctx, userID.(uint), uint(fileID), uint(collaboratorID)); err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "文件不存在":
			response.Error(ctx, http.StatusNotFound, errMsg)
		case "无权限管理此文件的协作者":
			response.Error(ctx, http.StatusForbidden, errMsg)
		case "协作关系不存在":
			response.Error(ctx, http.StatusNotFound, errMsg)
		default:
			response.Error(ctx, http.StatusInternalServerError, errMsg)
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetMyCollaborations 获取我的协作文件列表
// @Summary 获取我的协作文件列表
// @Tags 协作
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/collaborations/me [get]
func (c *CollaborationController) GetMyCollaborations(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	result, err := c.collaborationService.GetMyCollaborations(ctx, userID.(uint))
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// CheckFilePermission 检查文件权限
// @Summary 检查文件权限
// @Tags 协作
// @Produce json
// @Security Bearer
// @Param fileId path int true "文件ID"
// @Success 200 {object} response.Response
// @Router /api/files/{fileId}/permissions [get]
func (c *CollaborationController) CheckFilePermission(ctx *gin.Context) {
	fileIDStr := ctx.Param("fileId")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文件 ID")
		return
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	result, err := c.collaborationService.CheckFilePermission(ctx, userID.(uint), uint(fileID))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	response.Success(ctx, result)
}

// GetSharedWithMe returns files shared with the current user.
func (c *CollaborationController) GetSharedWithMe(ctx *gin.Context) {
	c.GetMyCollaborations(ctx)
}
