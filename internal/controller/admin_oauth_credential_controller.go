package controller

import (
	"net/http"
	"strconv"
	"strings"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

type AdminOAuthCredentialController struct {
	service service.OAuthCredentialService
}

func NewAdminOAuthCredentialController(oauthCredentialService service.OAuthCredentialService) *AdminOAuthCredentialController {
	return &AdminOAuthCredentialController{service: oauthCredentialService}
}

func (c *AdminOAuthCredentialController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page_size", "10")))
	var appID uint
	if raw := strings.TrimSpace(ctx.Query("app_id")); raw != "" {
		if parsed, err := strconv.ParseUint(raw, 10, 64); err == nil {
			appID = uint(parsed)
		}
	}
	var userID *uint
	if raw := strings.TrimSpace(ctx.Query("user_id")); raw != "" {
		if parsed, err := strconv.ParseUint(raw, 10, 64); err == nil {
			value := uint(parsed)
			userID = &value
		}
	}
	items, total, err := c.service.List(ctx, &service.OAuthCredentialListQuery{
		Page:     page,
		PageSize: pageSize,
		Provider: strings.TrimSpace(ctx.Query("provider")),
		Status:   strings.TrimSpace(ctx.Query("status")),
		AppID:    appID,
		UserID:   userID,
	})
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.PageSuccess(ctx, items, total, max(page, 1), normalizedPageSize(pageSize))
}

func (c *AdminOAuthCredentialController) Create(ctx *gin.Context) {
	var req service.OAuthCredentialInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := c.service.Create(ctx, &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "OAuth credential created", item)
}

func (c *AdminOAuthCredentialController) Refresh(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid credential id")
		return
	}
	item, err := c.service.RefreshOne(ctx, uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "OAuth credential refreshed", item)
}
