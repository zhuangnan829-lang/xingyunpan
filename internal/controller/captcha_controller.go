package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type CaptchaController struct {
	service service.CaptchaRuntimeService
}

func NewCaptchaController(captchaService service.CaptchaRuntimeService) *CaptchaController {
	return &CaptchaController{service: captchaService}
}

func (c *CaptchaController) Config(ctx *gin.Context) {
	scene := service.CaptchaScene(strings.TrimSpace(ctx.Query("scene")))
	identity := strings.TrimSpace(ctx.Query("identity"))
	data, err := c.service.PublicConfig(scene, ctx.Query("path"), identity, ctx.ClientIP())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *CaptchaController) Challenge(ctx *gin.Context) {
	var req struct {
		Scene    service.CaptchaScene `json:"scene" binding:"required"`
		Path     string               `json:"path"`
		Identity string               `json:"identity"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	data, err := c.service.CreateChallenge(req.Scene, req.Path, req.Identity, ctx.ClientIP())
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}
