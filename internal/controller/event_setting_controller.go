package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// EventSettingController handles admin event settings endpoints.
type EventSettingController struct {
	service service.EventSettingService
}

// NewEventSettingController creates an event settings controller.
func NewEventSettingController(eventSettingService service.EventSettingService) *EventSettingController {
	return &EventSettingController{service: eventSettingService}
}

// Get returns the current event settings.
func (c *EventSettingController) Get(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

// Update saves all event settings.
func (c *EventSettingController) Update(ctx *gin.Context) {
	var req service.EventSettingPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.Update(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "事件设置保存成功", data)
}

// Reset restores default event settings.
func (c *EventSettingController) Reset(ctx *gin.Context) {
	data, err := c.service.Reset()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "事件设置已恢复默认", data)
}

// ToggleAll enables or disables every event.
func (c *EventSettingController) ToggleAll(ctx *gin.Context) {
	var req service.EventTogglePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.ToggleAll(req.Enabled)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "事件设置已更新", data)
}

// ToggleCategory enables or disables all events in one category.
func (c *EventSettingController) ToggleCategory(ctx *gin.Context) {
	var req service.EventTogglePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.ToggleCategory(ctx.Param("categoryKey"), req.Enabled)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "事件分类设置已更新", data)
}

// ToggleEvent enables or disables one event.
func (c *EventSettingController) ToggleEvent(ctx *gin.Context) {
	var req service.EventTogglePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.ToggleEvent(ctx.Param("eventKey"), req.Enabled)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "事件设置已更新", data)
}
