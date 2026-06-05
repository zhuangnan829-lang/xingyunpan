package controller

import (
	"net/http"
	"strconv"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

// RecycleController controls recycle-bin APIs.
type RecycleController struct {
	recycleService           service.RecycleService
	fileSystemSettingService service.FileSystemSettingService
}

func NewRecycleController(recycleService service.RecycleService, fileSystemSettingService ...service.FileSystemSettingService) *RecycleController {
	var settings service.FileSystemSettingService
	if len(fileSystemSettingService) > 0 {
		settings = fileSystemSettingService[0]
	}
	return &RecycleController{recycleService: recycleService, fileSystemSettingService: settings}
}

// MoveToRecycle moves user files into the recycle bin.
func (c *RecycleController) MoveToRecycle(ctx *gin.Context) {
	var req struct {
		FileIDs []uint `json:"file_ids" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, ok := currentUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}
	if !validateMaxBatchActionSize(ctx, c.fileSystemSettingService, len(req.FileIDs)) {
		return
	}

	if err := c.recycleService.MoveToRecycle(ctx, userID, req.FileIDs); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetRecycleList returns recycle-bin items with pagination, sorting and filtering.
func (c *RecycleController) GetRecycleList(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	settings := c.getFileSystemSettings()
	pagination := resolveControllerPagination(settings, positiveIntQuery(ctx, "page", 1, 0), positiveIntQuery(ctx, "page_size", 20, 0), ctx.Query("cursor"), 20)
	query := service.RecycleListQuery{
		Page:           pagination.Page,
		PageSize:       pagination.PageSize,
		Cursor:         pagination.Cursor,
		UseCursor:      pagination.UseCursor,
		PaginationMode: pagination.Mode,
		MaxPageSize:    pagination.MaxPageSize,
		Sort:           ctx.DefaultQuery("sort", "deleted-desc"),
		Keyword:        ctx.Query("keyword"),
		FileType:       ctx.Query("file_type"),
	}

	resp, err := c.recycleService.GetRecycleList(ctx, userID, query)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, resp)
}

func (c *RecycleController) getFileSystemSettings() *service.FileSystemSettingPayload {
	if c.fileSystemSettingService == nil {
		return nil
	}
	settings, err := c.fileSystemSettingService.Get()
	if err != nil {
		return nil
	}
	return settings
}

// RestoreFiles restores files from the recycle bin.
func (c *RecycleController) RestoreFiles(ctx *gin.Context) {
	var req struct {
		ItemIDs []uint `json:"item_ids" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, ok := currentUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}
	if !validateMaxBatchActionSize(ctx, c.fileSystemSettingService, len(req.ItemIDs)) {
		return
	}

	if err := c.recycleService.RestoreFiles(ctx, userID, req.ItemIDs); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// PermanentDelete permanently deletes recycle-bin items.
func (c *RecycleController) PermanentDelete(ctx *gin.Context) {
	var req struct {
		ItemIDs []uint `json:"item_ids" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	userID, ok := currentUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}
	if !validateMaxBatchActionSize(ctx, c.fileSystemSettingService, len(req.ItemIDs)) {
		return
	}

	if err := c.recycleService.PermanentDelete(ctx, userID, req.ItemIDs); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *RecycleController) validateBatchSize(ctx *gin.Context, count int) bool {
	limit := 3000
	if c.fileSystemSettingService != nil {
		if settings, err := c.fileSystemSettingService.Get(); err == nil && settings != nil && settings.MaxBatchActionSize > 0 {
			limit = settings.MaxBatchActionSize
		}
	}
	if count > limit {
		response.Error(ctx, http.StatusBadRequest, "批量操作数量超过上限: "+strconv.Itoa(limit))
		return false
	}
	return true
}

// EmptyRecycleBin permanently deletes all recycle-bin items for the current user.
func (c *RecycleController) EmptyRecycleBin(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	if err := c.recycleService.EmptyRecycleBin(ctx, userID); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

func currentUserID(ctx *gin.Context) (uint, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return 0, false
	}

	switch value := userID.(type) {
	case uint:
		return value, true
	case int:
		if value > 0 {
			return uint(value), true
		}
	case float64:
		if value > 0 {
			return uint(value), true
		}
	}

	return 0, false
}

func positiveIntQuery(ctx *gin.Context, key string, fallback int, max int) int {
	value, err := strconv.Atoi(ctx.Query(key))
	if err != nil || value <= 0 {
		return fallback
	}
	if max > 0 && value > max {
		return max
	}
	return value
}
