package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// FileSystemSettingController handles admin file system settings endpoints.
type FileSystemSettingController struct {
	service service.FileSystemSettingService
}

// NewFileSystemSettingController creates a file system settings controller.
func NewFileSystemSettingController(fileSystemSettingService service.FileSystemSettingService) *FileSystemSettingController {
	return &FileSystemSettingController{service: fileSystemSettingService}
}

// Get returns the current file system settings.
func (c *FileSystemSettingController) Get(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

// GetClientSettings returns non-sensitive file-system settings needed by regular clients.
func (c *FileSystemSettingController) GetClientSettings(ctx *gin.Context) {
	data, err := c.service.Get()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"online_editor_size":     data.OnlineEditorSize,
		"online_editor_unit":     data.OnlineEditorUnit,
		"max_page_size":          data.MaxPageSize,
		"max_chunk_retry":        data.MaxChunkRetry,
		"cache_chunks_for_retry": data.CacheChunksForRetry,
		"transfer_parallelism":   data.TransferParallelism,
		"max_batch_action_size":  data.MaxBatchActionSize,
		"map_provider":           data.MapProvider,
		"show_encryption_status": data.ShowEncryptionStatus,
		"enable_event_push":      data.EnableEventPush,
		"debounce_delay":         data.DebounceDelay,
		"file_icon_rules":        data.FileIconRules,
		"emoji_options":          data.EmojiOptions,
	})
}

// Update saves the file system settings.
func (c *FileSystemSettingController) Update(ctx *gin.Context) {
	var req service.FileSystemSettingPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.Update(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "file system settings saved", data)
}

func (c *FileSystemSettingController) UpdateIcons(ctx *gin.Context) {
	var req service.FileSystemIconSettingsPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	data, err := c.service.UpdateIcons(&req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "file system icon settings saved", data)
}

// ClearBlobURLCache removes cached Blob presigned URLs.
func (c *FileSystemSettingController) ClearBlobURLCache(ctx *gin.Context) {
	if err := c.service.ClearBlobURLCache(); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "blob url cache cleared", gin.H{"cleared": true})
}

// GetBrowserApps returns normalized browser app groups for admin usage.
func (c *FileSystemSettingController) GetBrowserApps(ctx *gin.Context) {
	data, err := c.service.GetBrowserApps()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, data)
}

// ResolveBrowserApp resolves one browser app by filename/extension and platform.
func (c *FileSystemSettingController) ResolveBrowserApp(ctx *gin.Context) {
	filename := ctx.Query("filename")
	if filename == "" {
		filename = ctx.Query("extension")
	}
	if filename == "" {
		response.Error(ctx, http.StatusBadRequest, "filename or extension is required")
		return
	}

	mimeType := ctx.Query("mime_type")
	platform := service.BrowserAppPlatform(ctx.DefaultQuery("platform", "all"))

	data, err := c.service.ResolveBrowserApp(filename, mimeType, platform)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"matched": data != nil,
		"app":     data,
	})
}
