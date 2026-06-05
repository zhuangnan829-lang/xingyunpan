package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type AdminOAuthAppController struct {
	service service.AdminOAuthAppService
}

func NewAdminOAuthAppController(adminOAuthAppService service.AdminOAuthAppService) *AdminOAuthAppController {
	return &AdminOAuthAppController{service: adminOAuthAppService}
}

func (c *AdminOAuthAppController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page_size", "10")))
	items, total, err := c.service.List(ctx, &service.AdminOAuthAppListQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  strings.TrimSpace(ctx.Query("keyword")),
		Status:   strings.TrimSpace(ctx.Query("status")),
	})
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.PageSuccess(ctx, items, total, max(page, 1), normalizedPageSize(pageSize))
}

func (c *AdminOAuthAppController) Get(ctx *gin.Context) {
	item, err := c.service.Get(ctx, ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.Success(ctx, item)
}

func (c *AdminOAuthAppController) Create(ctx *gin.Context) {
	var req service.AdminOAuthAppPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := c.service.Create(ctx, &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "OAuth app created", item)
}

func (c *AdminOAuthAppController) Update(ctx *gin.Context) {
	var req service.AdminOAuthAppPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := c.service.Update(ctx, ctx.Param("id"), &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "OAuth app updated", item)
}

func (c *AdminOAuthAppController) UpdateStatus(ctx *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := c.service.UpdateStatus(ctx, ctx.Param("id"), req.Enabled)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "OAuth app status updated", item)
}

func (c *AdminOAuthAppController) Delete(ctx *gin.Context) {
	if err := c.service.Delete(ctx, ctx.Param("id")); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "OAuth app deleted", gin.H{"deleted": true})
}

func (c *AdminOAuthAppController) RegenerateSecret(ctx *gin.Context) {
	item, err := c.service.RegenerateSecret(ctx, ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "OAuth app secret regenerated", item)
}
