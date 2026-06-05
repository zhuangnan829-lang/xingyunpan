package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// EmailSettingController handles admin email settings endpoints.
type EmailSettingController struct {
	service service.EmailSettingService
}

// NewEmailSettingController creates an email settings controller.
func NewEmailSettingController(emailSettingService service.EmailSettingService) *EmailSettingController {
	return &EmailSettingController{service: emailSettingService}
}

func (c *EmailSettingController) Get(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *EmailSettingController) Update(ctx *gin.Context) {
	var req service.EmailSettingPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.Update(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "邮件设置保存成功", data)
}

func (c *EmailSettingController) SendTestEmail(ctx *gin.Context) {
	var req struct {
		ToEmail string `json:"to_email" binding:"required,email"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := c.service.SendTestEmail(req.ToEmail); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "测试邮件发送成功", gin.H{"to_email": req.ToEmail})
}
