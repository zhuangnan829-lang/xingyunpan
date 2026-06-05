package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// SiteSettingController handles admin site settings endpoints.
type SiteSettingController struct {
	service service.SiteSettingService
}

// NewSiteSettingController creates a site settings controller.
func NewSiteSettingController(siteSettingService service.SiteSettingService) *SiteSettingController {
	return &SiteSettingController{service: siteSettingService}
}

// Get returns the current site settings.
func (c *SiteSettingController) Get(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

// Update saves the site settings.
func (c *SiteSettingController) Update(ctx *gin.Context) {
	var req service.SiteSettingPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.Update(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "站点设置保存成功", data)
}
