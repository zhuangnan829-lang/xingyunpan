package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// FullTextSearchSettingController handles admin full text search settings endpoints.
type FullTextSearchSettingController struct {
	service service.FullTextSearchSettingService
}

// NewFullTextSearchSettingController creates a controller instance.
func NewFullTextSearchSettingController(searchSettingService service.FullTextSearchSettingService) *FullTextSearchSettingController {
	return &FullTextSearchSettingController{service: searchSettingService}
}

// Get returns current settings.
func (c *FullTextSearchSettingController) Get(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

// Update saves current settings.
func (c *FullTextSearchSettingController) Update(ctx *gin.Context) {
	var req service.FullTextSearchSettingPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Update(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "full text search settings saved", data)
}

// RebuildIndex enqueues a full text index rebuild task.
func (c *FullTextSearchSettingController) RebuildIndex(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "invalid user context")
		return
	}

	data, err := c.service.TriggerRebuild(userID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "full text rebuild job queued", data)
}
