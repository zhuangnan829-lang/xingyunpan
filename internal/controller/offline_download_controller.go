package controller

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type OfflineDownloadController struct {
	service                  service.OfflineDownloadService
	fileSystemSettingService service.FileSystemSettingService
}

func NewOfflineDownloadController(offlineDownloadService service.OfflineDownloadService, fileSystemSettingService ...service.FileSystemSettingService) *OfflineDownloadController {
	var settings service.FileSystemSettingService
	if len(fileSystemSettingService) > 0 {
		settings = fileSystemSettingService[0]
	}
	return &OfflineDownloadController{service: offlineDownloadService, fileSystemSettingService: settings}
}

func (c *OfflineDownloadController) List(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := c.service.List(ctx, userID, ctx.Query("status"), ctx.Query("keyword"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, items)
}

func (c *OfflineDownloadController) Create(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req service.OfflineDownloadCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	item, err := c.service.Create(ctx, userID, req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "offline download task created", item)
}

func (c *OfflineDownloadController) Refresh(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := c.service.Refresh(ctx, userID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, items)
}

func (c *OfflineDownloadController) Pause(ctx *gin.Context) {
	c.mutateOne(ctx, c.service.Pause)
}

func (c *OfflineDownloadController) Resume(ctx *gin.Context) {
	c.mutateOne(ctx, c.service.Resume)
}

func (c *OfflineDownloadController) Retry(ctx *gin.Context) {
	c.mutateOne(ctx, c.service.Retry)
}

func (c *OfflineDownloadController) Delete(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := service.ParseOfflineDownloadTaskID(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.service.Delete(ctx, userID, id); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, gin.H{"deleted": true})
}

func (c *OfflineDownloadController) BatchDelete(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req service.OfflineDownloadBatchDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if !validateMaxBatchActionSize(ctx, c.fileSystemSettingService, len(req.IDs)) {
		return
	}
	deleted, err := c.service.BatchDelete(ctx, userID, req.IDs)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, gin.H{"deleted": deleted})
}

func (c *OfflineDownloadController) mutateOne(
	ctx *gin.Context,
	action func(ctx context.Context, userID uint, id uint) (*service.OfflineDownloadTaskPayload, error),
) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, err := service.ParseOfflineDownloadTaskID(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	item, err := action(ctx, userID, id)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, item)
}
