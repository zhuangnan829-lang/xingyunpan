package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// UserGroupController handles admin user group endpoints.
type UserGroupController struct {
	service service.UserGroupService
}

// NewUserGroupController creates a user group controller.
func NewUserGroupController(userGroupService service.UserGroupService) *UserGroupController {
	return &UserGroupController{service: userGroupService}
}

func (c *UserGroupController) List(ctx *gin.Context) {
	data, err := c.service.List()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *UserGroupController) Summary(ctx *gin.Context) {
	data, err := c.service.Summary()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *UserGroupController) ListMembers(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid user group id")
		return
	}

	data, err := c.service.ListMembers(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *UserGroupController) Create(ctx *gin.Context) {
	var req service.UserGroupPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Create(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "user group created", data)
}

func (c *UserGroupController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid user group id")
		return
	}

	var req service.UserGroupPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Update(uint(id), &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "user group saved", data)
}

func (c *UserGroupController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid user group id")
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "user group deleted", gin.H{"deleted": true})
}
