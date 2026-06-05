package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// QueueController handles admin queue settings and stats endpoints.
type QueueController struct {
	settings                 service.QueueSettingService
	stats                    service.QueueStatsService
	jobs                     service.QueueJobService
	fileSystemSettingService service.FileSystemSettingService
	runtime                  service.QueueRuntimeService
}

// NewQueueController creates a queue controller.
func NewQueueController(settings service.QueueSettingService, stats service.QueueStatsService, jobs service.QueueJobService, fileSystemSettingService ...service.FileSystemSettingService) *QueueController {
	var fsSettings service.FileSystemSettingService
	if len(fileSystemSettingService) > 0 {
		fsSettings = fileSystemSettingService[0]
	}
	return &QueueController{settings: settings, stats: stats, jobs: jobs, fileSystemSettingService: fsSettings}
}

func (c *QueueController) SetRuntimeService(runtime service.QueueRuntimeService) {
	c.runtime = runtime
}

// GetSettings returns current queue settings.
func (c *QueueController) GetSettings(ctx *gin.Context) {
	data, err := c.settings.GetAll()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

// UpdateSettings saves queue settings.
func (c *QueueController) UpdateSettings(ctx *gin.Context) {
	var req []service.QueueSettingItemPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.settings.UpdateAll(req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "queue settings saved", data)
}

// GetStats returns current queue stats.
func (c *QueueController) GetStats(ctx *gin.Context) {
	data, err := c.stats.GetAll()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

func (c *QueueController) GetRuntime(ctx *gin.Context) {
	if c.runtime == nil {
		response.Error(ctx, http.StatusInternalServerError, "queue runtime service is not initialized")
		return
	}
	data, err := c.runtime.GetStatus()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

// GetJobs returns paged queue job details.
func (c *QueueController) GetJobs(ctx *gin.Context) {
	queueKey := ctx.Query("queue_key")
	status := ctx.Query("status")
	var nodeID uint
	if value := ctx.Query("node_id"); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
			nodeID = uint(parsed)
		}
	}

	page := 1
	if value := ctx.Query("page"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 20
	if value := ctx.Query("page_size"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	settings := c.getFileSystemSettings()
	pagination := resolveControllerPagination(settings, page, pageSize, ctx.Query("cursor"), 20)

	data, err := c.jobs.List(queueKey, status, nodeID, pagination.Page, pagination.PageSize, pagination.Cursor, pagination.UseCursor)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	data.PaginationMode = pagination.Mode
	data.PageSize = pagination.PageSize
	data.MaxPageSize = pagination.MaxPageSize
	if pagination.UseCursor && len(data.List) > 0 {
		data.NextCursor = service.NextCursorFromUint(data.List[len(data.List)-1].ID)
	}

	response.Success(ctx, data)
}

// GetJob returns one queue job detail.
func (c *QueueController) GetJob(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(ctx, http.StatusBadRequest, "invalid queue job id")
		return
	}

	data, err := c.jobs.Get(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, data)
}

// RetryJob requeues one queue job for another execution.
func (c *QueueController) RetryJob(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(ctx, http.StatusBadRequest, "invalid queue job id")
		return
	}

	data, err := c.jobs.Retry(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "queue job requeued", data)
}

// RecoverStaleJobs restores processing jobs that exceeded their execution window.
func (c *QueueController) RecoverStaleJobs(ctx *gin.Context) {
	data, err := c.jobs.RecoverStale(ctx.Query("queue_key"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "stale queue jobs recovered", data)
}

// DeleteJob removes one queue job record.
func (c *QueueController) DeleteJob(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(ctx, http.StatusBadRequest, "invalid queue job id")
		return
	}

	data, err := c.jobs.Delete(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "queue job deleted", data)
}

// BatchDeleteJobs removes selected queue job records.
func (c *QueueController) BatchDeleteJobs(ctx *gin.Context) {
	var req service.QueueJobBatchDeletePayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	if !validateMaxBatchActionSize(ctx, c.fileSystemSettingService, len(req.JobIDs)) {
		return
	}

	data, err := c.jobs.BatchDelete(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "queue jobs deleted", data)
}

// ClearJobs removes completed or failed queue job records matching a filter.
func (c *QueueController) ClearJobs(ctx *gin.Context) {
	var req service.QueueJobClearPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := c.jobs.Clear(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "queue jobs cleared", data)
}

func (c *QueueController) getFileSystemSettings() *service.FileSystemSettingPayload {
	if c.fileSystemSettingService == nil {
		return nil
	}
	settings, err := c.fileSystemSettingService.Get()
	if err != nil {
		return nil
	}
	return settings
}
