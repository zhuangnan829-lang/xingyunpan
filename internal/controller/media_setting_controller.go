package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// MediaSettingController handles admin media settings endpoints.
type MediaSettingController struct {
	service service.MediaSettingService
}

// NewMediaSettingController creates a media settings controller.
func NewMediaSettingController(mediaSettingService service.MediaSettingService) *MediaSettingController {
	return &MediaSettingController{service: mediaSettingService}
}

// Get returns the current media settings.
func (c *MediaSettingController) Get(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

// Update saves the media settings.
func (c *MediaSettingController) Update(ctx *gin.Context) {
	var req service.MediaSettingPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Update(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "media settings saved", data)
}
