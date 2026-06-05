package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// CaptchaSettingController handles admin captcha settings endpoints.
type CaptchaSettingController struct {
	service service.CaptchaSettingService
}

// NewCaptchaSettingController creates a captcha settings controller.
func NewCaptchaSettingController(captchaSettingService service.CaptchaSettingService) *CaptchaSettingController {
	return &CaptchaSettingController{service: captchaSettingService}
}

// Get returns the current captcha settings.
func (c *CaptchaSettingController) Get(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

// Update saves the captcha settings.
func (c *CaptchaSettingController) Update(ctx *gin.Context) {
	var req service.CaptchaSettingPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.Update(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "验证码设置保存成功", data)
}
