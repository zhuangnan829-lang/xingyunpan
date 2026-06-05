package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

const defaultMaxBatchActionSize = 3000

func validateMaxBatchActionSize(ctx *gin.Context, settingsService service.FileSystemSettingService, count int) bool {
	limit := defaultMaxBatchActionSize
	if settingsService != nil {
		if settings, err := settingsService.Get(); err == nil && settings != nil && settings.MaxBatchActionSize > 0 {
			limit = settings.MaxBatchActionSize
		}
	}
	if count > limit {
		response.Error(ctx, http.StatusBadRequest, fmt.Sprintf("最多允许 %d 项", limit))
		return false
	}
	return true
}
