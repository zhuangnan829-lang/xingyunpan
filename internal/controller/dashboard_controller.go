package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type DashboardController struct {
	service service.DashboardService
}

func NewDashboardController(dashboardService service.DashboardService) *DashboardController {
	return &DashboardController{service: dashboardService}
}

func (c *DashboardController) Overview(ctx *gin.Context) {
	data, err := c.service.GetOverview(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}
