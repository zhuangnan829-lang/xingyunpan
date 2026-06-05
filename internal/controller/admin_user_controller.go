package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type AdminUserController struct {
	service service.AdminUserService
}

func NewAdminUserController(adminUserService service.AdminUserService) *AdminUserController {
	return &AdminUserController{service: adminUserService}
}

func (c *AdminUserController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page_size", "10")))

	var userGroupID uint
	if rawUserGroupID := strings.TrimSpace(ctx.Query("user_group_id")); rawUserGroupID != "" {
		parsed, err := strconv.ParseUint(rawUserGroupID, 10, 64)
		if err != nil {
			response.Error(ctx, http.StatusBadRequest, "invalid user group id")
			return
		}
		userGroupID = uint(parsed)
	}

	data, err := c.service.ListUsers(&service.AdminUserListQuery{
		Keyword:       strings.TrimSpace(ctx.Query("keyword")),
		Role:          strings.TrimSpace(ctx.DefaultQuery("role", "all")),
		Status:        strings.TrimSpace(ctx.DefaultQuery("status", "all")),
		UserGroupID:   userGroupID,
		UserGroupName: strings.TrimSpace(ctx.Query("user_group_name")),
		Page:          page,
		PageSize:      pageSize,
	})
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

func (c *AdminUserController) Create(ctx *gin.Context) {
	var req service.AdminUserUpsertPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Create(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "user created", data)
}

func (c *AdminUserController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	operatorID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var req service.AdminUserUpsertPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Update(uint(id), &req, operatorID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "user updated", data)
}

func (c *AdminUserController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	operatorID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	if err := c.service.Delete(uint(id), operatorID); err != nil {
		var blocked *service.AdminUserDeleteBlockedError
		if errors.As(err, &blocked) {
			response.ErrorWithData(ctx, http.StatusBadRequest, err.Error(), blocked.Preview)
			return
		}
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "user deleted", gin.H{"deleted": true})
}

func (c *AdminUserController) DeletePreview(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	data, err := c.service.GetDeletePreview(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, data)
}

func (c *AdminUserController) ResetPassword(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	var req service.AdminUserResetPasswordPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := c.service.ResetPassword(uint(id), &req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "password reset", gin.H{"updated": true})
}

func (c *AdminUserController) BatchDelete(ctx *gin.Context) {
	operatorID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var req service.AdminUserBatchDeletePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if err := c.service.BatchDelete(&req, operatorID); err != nil {
		var blocked *service.AdminUserDeleteBlockedError
		if errors.As(err, &blocked) {
			response.ErrorWithData(ctx, http.StatusBadRequest, err.Error(), blocked.Preview)
			return
		}
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "users deleted", gin.H{"deleted": true})
}

func (c *AdminUserController) BatchUpdateGroup(ctx *gin.Context) {
	operatorID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var req service.AdminUserBatchGroupPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.BatchUpdateGroup(&req, operatorID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "user groups updated", data)
}

func (c *AdminUserController) BatchUpdateRole(ctx *gin.Context) {
	operatorID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var req service.AdminUserBatchRolePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.BatchUpdateRole(&req, operatorID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "user roles updated", data)
}

func (c *AdminUserController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	operatorID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var req service.AdminUserStatusPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.UpdateStatus(uint(id), &req, operatorID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "user status updated", data)
}

func (c *AdminUserController) BatchUpdateStatus(ctx *gin.Context) {
	operatorID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	var req service.AdminUserBatchStatusPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.BatchUpdateStatus(&req, operatorID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "user status batch updated", data)
}
