package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type AppearanceSettingController struct {
	service service.AppearanceSettingService
}

func NewAppearanceSettingController(appearanceSettingService service.AppearanceSettingService) *AppearanceSettingController {
	return &AppearanceSettingController{service: appearanceSettingService}
}

func (c *AppearanceSettingController) Get(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

func (c *AppearanceSettingController) Update(ctx *gin.Context) {
	var req service.AppearanceSettingPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.Update(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "外观设置保存成功", data)
}
