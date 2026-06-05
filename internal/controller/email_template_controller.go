package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// EmailTemplateController handles admin email template endpoints.
type EmailTemplateController struct {
	service service.EmailTemplateService
}

// NewEmailTemplateController creates an email template controller.
func NewEmailTemplateController(emailTemplateService service.EmailTemplateService) *EmailTemplateController {
	return &EmailTemplateController{service: emailTemplateService}
}

func (c *EmailTemplateController) List(ctx *gin.Context) {
	data, err := c.service.List()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *EmailTemplateController) Update(ctx *gin.Context) {
	templateKey := ctx.Param("templateKey")
	var req service.EmailTemplatePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	data, err := c.service.Update(templateKey, &req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "邮件模板保存成功", data)
}
