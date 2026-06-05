package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type AdminShareController struct {
	service                  service.AdminShareService
	fileSystemSettingService service.FileSystemSettingService
}

func NewAdminShareController(adminShareService service.AdminShareService, fileSystemSettingService ...service.FileSystemSettingService) *AdminShareController {
	var settings service.FileSystemSettingService
	if len(fileSystemSettingService) > 0 {
		settings = fileSystemSettingService[0]
	}
	return &AdminShareController{service: adminShareService, fileSystemSettingService: settings}
}

func (c *AdminShareController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page_size", "10")))
	settings := c.getFileSystemSettings()
	pagination := resolveControllerPagination(settings, page, pageSize, ctx.Query("cursor"), 10)

	var ownerID uint64
	if rawOwnerID := strings.TrimSpace(ctx.Query("owner_id")); rawOwnerID != "" {
		parsed, err := strconv.ParseUint(rawOwnerID, 10, 64)
		if err != nil {
			response.Error(ctx, http.StatusBadRequest, "invalid owner id")
			return
		}
		ownerID = parsed
	}

	minDownloads, err := parseOptionalIntQuery(ctx, "min_downloads")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid min_downloads")
		return
	}
	expiringWithinDays, err := parseOptionalIntQuery(ctx, "expiring_within_days")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid expiring_within_days")
		return
	}
	maxDownloadsReached, err := parseOptionalBoolQuery(ctx, "max_downloads_reached")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid max_downloads_reached")
		return
	}
	unavailable, err := parseOptionalBoolQuery(ctx, "unavailable")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid unavailable")
		return
	}

	items, total, err := c.service.List(ctx, &service.AdminShareListQuery{
		Page:                pagination.Page,
		PageSize:            pagination.PageSize,
		Cursor:              pagination.Cursor,
		UseCursor:           pagination.UseCursor,
		Keyword:             strings.TrimSpace(ctx.Query("keyword")),
		OwnerID:             uint(ownerID),
		Status:              strings.TrimSpace(ctx.Query("status")),
		MinDownloads:        minDownloads,
		ExpiringWithinDays:  expiringWithinDays,
		MaxDownloadsReached: maxDownloadsReached,
		Unavailable:         unavailable,
	})
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	nextCursor := ""
	if pagination.UseCursor && len(items) > 0 {
		nextCursor = service.NextCursorFromUint(items[len(items)-1].ShareID)
	}
	response.Success(ctx, gin.H{
		"list":            items,
		"total":           total,
		"page":            pagination.Page,
		"page_size":       pagination.PageSize,
		"pagination_mode": pagination.Mode,
		"next_cursor":     nextCursor,
		"max_page_size":   pagination.MaxPageSize,
	})
}

func (c *AdminShareController) Metrics(ctx *gin.Context) {
	expiringWithinDays := 3
	if raw := strings.TrimSpace(ctx.Query("expiring_within_days")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			response.Error(ctx, http.StatusBadRequest, "invalid expiring_within_days")
			return
		}
		expiringWithinDays = parsed
	}

	data, err := c.service.Metrics(ctx, expiringWithinDays)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

func (c *AdminShareController) Delete(ctx *gin.Context) {
	shareID, ok := parseUintParam(ctx, "id")
	if !ok {
		response.Error(ctx, http.StatusBadRequest, "invalid share id")
		return
	}

	if err := c.service.Delete(ctx, shareID); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "share deleted", gin.H{"deleted": true})
}

func (c *AdminShareController) BatchDelete(ctx *gin.Context) {
	var req struct {
		ShareIDs []uint `json:"share_ids"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	if !validateMaxBatchActionSize(ctx, c.fileSystemSettingService, len(req.ShareIDs)) {
		return
	}

	if err := c.service.BatchDelete(ctx, req.ShareIDs); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "shares deleted", gin.H{"deleted": true})
}

func (c *AdminShareController) getFileSystemSettings() *service.FileSystemSettingPayload {
	if c.fileSystemSettingService == nil {
		return nil
	}
	settings, err := c.fileSystemSettingService.Get()
	if err != nil {
		return nil
	}
	return settings
}
