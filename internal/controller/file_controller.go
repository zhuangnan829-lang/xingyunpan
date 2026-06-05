package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/mimetype"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FileController handles file endpoints.
type FileController struct {
	fileService              service.FileService
	fileSystemSettingService service.FileSystemSettingService
	recycleService           service.RecycleService
	fileEventService         service.FileEventService
}

// NewFileController creates a file controller.
func NewFileController(fileService service.FileService, fileSystemSettingService service.FileSystemSettingService, recycleService service.RecycleService, fileEventService ...service.FileEventService) *FileController {
	var eventService service.FileEventService
	if len(fileEventService) > 0 {
		eventService = fileEventService[0]
	}
	return &FileController{
		fileService:              fileService,
		fileSystemSettingService: fileSystemSettingService,
		recycleService:           recycleService,
		fileEventService:         eventService,
	}
}

func validatePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 10000 {
		pageSize = 10000
	}
	return page, pageSize
}

// Upload handles direct file upload.
func (c *FileController) Upload(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	file, err := ctx.FormFile("file")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "failed to get upload file")
		return
	}

	src, err := file.Open()
	if err != nil {
		logger.Error("open upload file failed",
			zap.Uint("user_id", userID.(uint)),
			zap.String("filename", file.Filename),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "failed to open upload file")
		return
	}
	defer src.Close()

	var parentID *uint
	if rawParentID := ctx.PostForm("parent_id"); rawParentID != "" {
		if parsed, parseErr := strconv.ParseUint(rawParentID, 10, 32); parseErr != nil {
			response.Error(ctx, http.StatusBadRequest, "invalid parent_id")
			return
		} else {
			value := uint(parsed)
			parentID = &value
		}
	}

	userFile, err := c.fileService.Upload(userID.(uint), file.Filename, file.Size, src, parentID)
	if err != nil {
		logger.Error("upload file failed",
			zap.Uint("user_id", userID.(uint)),
			zap.String("filename", file.Filename),
			zap.Int64("size", file.Size),
			zap.Error(err))
		if service.IsStoragePolicyValidationError(err) {
			response.Error(ctx, http.StatusBadRequest, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "upload failed")
		}
		return
	}

	response.SuccessWithMessage(ctx, "upload success", buildUserFileMutationResponse(userFile))
}

// Create handles creating a text-like file from a browser-app template.
func (c *FileController) Create(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	var req struct {
		Name     string `json:"name"`
		Kind     string `json:"kind"`
		Content  string `json:"content"`
		ParentID *uint  `json:"parent_id"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid params")
		return
	}

	fileName, content, err := resolveCreateFilePayload(req.Kind, req.Name, req.Content)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userFile, err := c.fileService.CreateFile(userID.(uint), fileName, []byte(content), req.ParentID)
	if err != nil {
		logger.Error("create file failed",
			zap.Uint("user_id", userID.(uint)),
			zap.String("file_name", fileName),
			zap.String("kind", req.Kind),
			zap.Error(err))
		if service.IsStoragePolicyValidationError(err) {
			response.Error(ctx, http.StatusBadRequest, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.SuccessWithMessage(ctx, "create file success", buildUserFileMutationResponse(userFile))
}

func resolveCreateFilePayload(kind, name, content string) (string, string, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	name = strings.TrimSpace(name)

	defaults := map[string]struct {
		name    string
		content string
	}{
		"file":       {name: "新建文件.txt", content: ""},
		"text":       {name: "新建文本.txt", content: ""},
		"markdown":   {name: "新建 Markdown.md", content: "# 新建 Markdown\n"},
		"drawio":     {name: "新建图表.drawio", content: `<mxfile host="星云盘"><diagram name="Page-1"><mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/></root></mxGraphModel></diagram></mxfile>`},
		"dwb":        {name: "新建白板.dwb", content: `{"type":"dwb","version":1,"elements":[]}`},
		"excalidraw": {name: "新建 Excalidraw.excalidraw", content: `{"type":"excalidraw","version":2,"source":"xingyunpan","elements":[],"appState":{},"files":{}}`},
	}

	item, ok := defaults[kind]
	if !ok {
		return "", "", fmt.Errorf("unsupported file kind")
	}
	if name == "" {
		name = item.name
	}
	if content == "" {
		content = item.content
	}
	if len([]byte(content)) > 5*1024*1024 {
		return "", "", fmt.Errorf("file content is too large")
	}

	return name, content, nil
}

// List handles file listing.
func (c *FileController) List(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	settings := c.getFileSystemSettings()
	pagination := resolveControllerPagination(settings, page, pageSize, ctx.Query("cursor"), 20)
	page = pagination.Page
	pageSize = pagination.PageSize

	var parentID *uint
	if rawParentID := ctx.Query("parent_id"); rawParentID != "" {
		if parsed, err := strconv.ParseUint(rawParentID, 10, 32); err == nil {
			value := uint(parsed)
			parentID = &value
		}
	}

	files, total, err := c.fileService.List(userID.(uint), parentID, page, pageSize)
	var nextCursor string
	if pagination.UseCursor {
		files, total, err = c.fileService.ListAfterID(userID.(uint), parentID, pagination.Cursor, pageSize)
		if len(files) > 0 {
			nextCursor = service.NextCursorFromUint(files[len(files)-1].ID)
		}
	}
	if err != nil {
		logger.Error("list files failed",
			zap.Uint("user_id", userID.(uint)),
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "failed to list files")
		return
	}

	var browserApps []service.BrowserAppGroupPayload
	if c.fileSystemSettingService != nil {
		browserApps, err = c.fileSystemSettingService.GetBrowserApps()
		if err != nil {
			logger.Warn("load browser apps for file list failed", zap.Error(err))
		}
	}

	directoryStats := make(map[uint]service.DirectoryStatsPayload)
	encryptionStatuses := make(map[uint]service.EncryptionStatusPayload)
	folderIDs := make([]uint, 0, len(files))
	fileIDs := make([]uint, 0, len(files))
	for _, file := range files {
		if file.IsFolder {
			folderIDs = append(folderIDs, file.ID)
		} else {
			fileIDs = append(fileIDs, file.ID)
		}
	}
	if len(folderIDs) > 0 {
		directoryStats, err = c.fileService.GetDirectoryStats(userID.(uint), folderIDs)
		if err != nil {
			logger.Warn("load directory stats failed", zap.Error(err))
			directoryStats = make(map[uint]service.DirectoryStatsPayload)
		}
	}
	if len(fileIDs) > 0 {
		encryptionStatuses, err = c.fileService.GetEncryptionStatuses(userID.(uint), fileIDs)
		if err != nil {
			logger.Warn("load encryption statuses failed", zap.Error(err))
			encryptionStatuses = make(map[uint]service.EncryptionStatusPayload)
		}
	}

	response.Success(ctx, gin.H{
		"list":            buildFileListResponse(files, browserApps, settings, directoryStats, encryptionStatuses),
		"total":           total,
		"page":            page,
		"page_size":       pageSize,
		"mode":            pagination.Mode,
		"pagination_mode": pagination.Mode,
		"next_cursor":     nextCursor,
		"max_page_size":   pagination.MaxPageSize,
	})
}

// Events streams debounced file-system change events for the current user.
func (c *FileController) Events(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	if c.fileEventService == nil {
		response.Error(ctx, http.StatusNotImplemented, "file events are not enabled")
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	events, cleanup := c.fileEventService.Subscribe(ctx.Request.Context(), userID.(uint))
	defer cleanup()

	ctx.Stream(func(w io.Writer) bool {
		select {
		case payload, ok := <-events:
			if !ok {
				return false
			}
			_, _ = fmt.Fprintf(w, "event: file-change\ndata: %s\n\n", service.MarshalFileEvent(payload))
			return true
		case <-ctx.Request.Context().Done():
			return false
		}
	})
}

// Rename handles file rename.
func (c *FileController) Rename(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid params")
		return
	}

	if err := c.fileService.Rename(userID.(uint), uint(fileID), req.Name); err != nil {
		logger.Error("rename file failed",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("file_id", fileID),
			zap.String("new_name", req.Name),
			zap.Error(err))
		if service.IsStoragePolicyValidationError(err) {
			response.Error(ctx, http.StatusBadRequest, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.SuccessWithMessage(ctx, "rename success", nil)
}

// Download streams one file to the client.
func (c *FileController) Download(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	inline := ctx.Query("inline") == "1"

	result, err := c.fileService.DownloadWithDelivery(userID.(uint), uint(fileID), inline)
	if err != nil {
		logger.Error("download file failed",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("file_id", fileID),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if strings.TrimSpace(result.RedirectURL) != "" {
		ctx.Redirect(http.StatusFound, result.RedirectURL)
		return
	}
	if result.Reader == nil {
		response.Error(ctx, http.StatusInternalServerError, "download is not available")
		return
	}
	defer result.Reader.Close()

	disposition := "attachment"
	if inline {
		disposition = "inline"
	}

	ctx.Header("Content-Type", result.ContentType)
	ctx.Header("Content-Disposition", fmt.Sprintf("%s; filename*=UTF-8''%s", disposition, url.PathEscape(result.FileName)))
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, result.Reader)
}

// PreviewPDF converts a supported document to PDF and streams it inline.
func (c *FileController) PreviewPDF(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	reader, fileName, err := c.fileService.PreviewPDF(userID.(uint), uint(fileID))
	if err != nil {
		logger.Error("preview pdf failed",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("file_id", fileID),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer reader.Close()

	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename*=UTF-8''%s", url.PathEscape(fileName)))
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, reader)
}

// Delete handles file deletion.
func (c *FileController) Delete(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var err error
	if c.recycleService != nil {
		err = c.recycleService.MoveToRecycle(ctx, userID.(uint), []uint{uint(fileID)})
	} else {
		err = c.fileService.Delete(userID.(uint), uint(fileID))
	}
	if err != nil {
		logger.Error("delete file failed",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("file_id", fileID),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "delete failed")
		return
	}

	response.SuccessWithMessage(ctx, "delete success", nil)
}

// Move handles moving a file to another folder.
func (c *FileController) Move(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var req struct {
		ParentID *uint `json:"parent_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid params")
		return
	}

	if err := c.fileService.Move(userID.(uint), uint(fileID), req.ParentID); err != nil {
		logger.Error("move file failed",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("file_id", fileID),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "move success", nil)
}

// Copy handles copying a file to another folder.
func (c *FileController) Copy(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	var req struct {
		ParentID *uint `json:"parent_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid params")
		return
	}

	userFile, err := c.fileService.Copy(userID.(uint), uint(fileID), req.ParentID)
	if err != nil {
		logger.Error("copy file failed",
			zap.Uint("user_id", userID.(uint)),
			zap.Uint64("file_id", fileID),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "copy success", buildUserFileMutationResponse(userFile))
}

func buildUserFileMutationResponse(file *model.UserFile) gin.H {
	if file == nil {
		return gin.H{}
	}

	fileSize := file.FileSize
	if fileSize <= 0 && file.PhysicalFile != nil && file.PhysicalFile.FileSize > 0 {
		fileSize = file.PhysicalFile.FileSize
	}

	payload := gin.H{
		"id":               file.ID,
		"name":             file.FileName,
		"file_name":        file.FileName,
		"is_folder":        file.IsFolder,
		"parent_id":        file.ParentID,
		"size":             fileSize,
		"physical_file_id": file.PhysicalFileID,
		"created_at":       file.CreatedAt,
		"updated_at":       file.UpdatedAt,
	}
	if file.PhysicalFile != nil {
		payload["hash"] = file.PhysicalFile.FileHash
		payload["mime_type"] = file.PhysicalFile.ContentType
		payload["content_type"] = file.PhysicalFile.ContentType
	}
	return payload
}

// Thumbnail serves one file thumbnail image.
func (c *FileController) Thumbnail(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	fileID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)

	data, contentType, err := c.fileService.GetThumbnail(userID.(uint), uint(fileID))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	ctx.Data(http.StatusOK, contentType, data)
}

// Check handles instant-upload existence checks.
func (c *FileController) Check(ctx *gin.Context) {
	var req struct {
		Hash string `json:"hash" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid params")
		return
	}

	file, exists, err := c.fileService.CheckExists(req.Hash)
	if err != nil {
		logger.Error("check file exists failed", zap.String("hash", req.Hash), zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "check failed")
		return
	}

	if exists {
		response.SuccessWithMessage(ctx, "file exists", gin.H{
			"exists":  true,
			"file_id": file.ID,
		})
		return
	}

	response.SuccessWithMessage(ctx, "file not found", gin.H{
		"exists": false,
	})
}

func buildFileListResponse(files []*model.UserFile, browserApps []service.BrowserAppGroupPayload, settings *service.FileSystemSettingPayload, directoryStats map[uint]service.DirectoryStatsPayload, encryptionStatuses map[uint]service.EncryptionStatusPayload) []gin.H {
	mimeMap := parseMimeMap(settings)
	onlineEditLimitBytes := onlineEditorLimitBytes(settings)
	iconRules := parseFileIconRules(settings)
	emojiOptions := parseFileEmojiOptions(settings)
	items := make([]gin.H, 0, len(files))
	for _, file := range files {
		contentType := "application/octet-stream"
		if file.PhysicalFile != nil && !isGenericListContentType(file.PhysicalFile.ContentType) {
			contentType = strings.TrimSpace(file.PhysicalFile.ContentType)
		} else if resolved := resolveMimeTypeFromMap(file.FileName, mimeMap); resolved != "" {
			contentType = resolved
		}

		fileSize := file.FileSize
		if fileSize <= 0 && file.PhysicalFile != nil && file.PhysicalFile.FileSize > 0 {
			fileSize = file.PhysicalFile.FileSize
		}

		item := gin.H{
			"id":         file.ID,
			"name":       file.FileName,
			"file_name":  file.FileName,
			"is_folder":  file.IsFolder,
			"parent_id":  file.ParentID,
			"size":       fileSize,
			"created_at": file.CreatedAt,
			"updated_at": file.UpdatedAt,
		}
		displayIcon := resolveDisplayIcon(file.FileName, file.IsFolder, contentType, iconRules, emojiOptions)
		if displayIcon.Icon != "" {
			item["display_icon"] = displayIcon.Icon
			item["display_icon_tint"] = displayIcon.Tint
			item["display_icon_label"] = displayIcon.Label
			item["display_icon_source"] = displayIcon.Source
		}

		if !file.IsFolder && onlineEditLimitBytes > 0 {
			item["online_editable"] = fileSize > 0 && fileSize <= onlineEditLimitBytes
		}
		if !file.IsFolder {
			if status, ok := encryptionStatuses[file.ID]; ok {
				item["encryption_status"] = status
			}
		}

		if file.PhysicalFileID != nil {
			item["physical_file_id"] = *file.PhysicalFileID
			item["thumbnail_url"] = "/api/v1/file/" + strconv.FormatUint(uint64(file.ID), 10) + "/thumbnail"
		}

		if file.PhysicalFile != nil {
			item["hash"] = file.PhysicalFile.FileHash
			item["mime_type"] = contentType
			item["content_type"] = contentType
			item["storage_type"] = file.PhysicalFile.StorageType
			item["browser_app"] = service.ResolveBrowserAppFromGroups(
				browserApps,
				file.FileName,
				contentType,
				service.BrowserAppPlatformAll,
			)
		} else {
			item["hash"] = ""
			item["mime_type"] = "application/octet-stream"
			item["content_type"] = "application/octet-stream"
			item["storage_type"] = ""
			item["browser_app"] = nil
		}

		if file.IsFolder {
			item["browser_app"] = nil
			if stats, ok := directoryStats[file.ID]; ok {
				item["directory_stats"] = stats
			}
		}

		items = append(items, item)
	}

	return items
}

type fileIconRule struct {
	Label string `json:"label"`
	Icon  string `json:"icon"`
	Match string `json:"match"`
	Tint  string `json:"tint"`
}

type fileEmojiOptions struct {
	Enabled         bool                `json:"enabled"`
	ShowInList      bool                `json:"showInList"`
	FallbackUnknown bool                `json:"fallbackUnknown"`
	FolderEmoji     string              `json:"folderEmoji"`
	UnknownEmoji    string              `json:"unknownEmoji"`
	Categories      []fileEmojiCategory `json:"categories"`
}

type fileEmojiCategory struct {
	Label  string        `json:"label"`
	Icon   string        `json:"icon"`
	Match  string        `json:"match"`
	Emojis fileEmojiList `json:"emojis"`
}

type fileEmojiList []string

func (l *fileEmojiList) UnmarshalJSON(data []byte) error {
	var values []string
	if err := json.Unmarshal(data, &values); err == nil {
		*l = values
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*l = splitIconMatch(raw)
	return nil
}

type fileDisplayIcon struct {
	Icon   string
	Tint   string
	Label  string
	Source string
}

func parseFileIconRules(settings *service.FileSystemSettingPayload) []fileIconRule {
	if settings == nil || strings.TrimSpace(settings.FileIconRules) == "" {
		return nil
	}
	var rules []fileIconRule
	if err := json.Unmarshal([]byte(settings.FileIconRules), &rules); err != nil {
		return nil
	}
	return rules
}

func parseFileEmojiOptions(settings *service.FileSystemSettingPayload) fileEmojiOptions {
	options := fileEmojiOptions{
		Enabled:         true,
		ShowInList:      true,
		FallbackUnknown: true,
		FolderEmoji:     "📁",
		UnknownEmoji:    "🗂️",
	}
	if settings == nil || strings.TrimSpace(settings.EmojiOptions) == "" {
		return options
	}
	_ = json.Unmarshal([]byte(settings.EmojiOptions), &options)
	if strings.TrimSpace(options.FolderEmoji) == "" {
		options.FolderEmoji = "📁"
	}
	if strings.TrimSpace(options.UnknownEmoji) == "" {
		options.UnknownEmoji = "🗂️"
	}
	return options
}

func resolveDisplayIcon(fileName string, isFolder bool, contentType string, rules []fileIconRule, emojiOptions fileEmojiOptions) fileDisplayIcon {
	if isFolder {
		if emojiOptions.Enabled && emojiOptions.ShowInList {
			return fileDisplayIcon{Icon: emojiOptions.FolderEmoji, Label: "Folder", Source: "emoji"}
		}
		return fileDisplayIcon{}
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	lowerMime := strings.ToLower(strings.TrimSpace(contentType))
	for _, rule := range rules {
		if matchesIconMatch(rule.Match, ext, lowerMime) {
			return fileDisplayIcon{
				Icon:   strings.TrimSpace(rule.Icon),
				Tint:   strings.TrimSpace(rule.Tint),
				Label:  strings.TrimSpace(rule.Label),
				Source: "rule",
			}
		}
	}

	if emojiOptions.Enabled && emojiOptions.ShowInList {
		for _, category := range emojiOptions.Categories {
			if strings.TrimSpace(category.Match) == "" || strings.TrimSpace(category.Icon) == "" {
				continue
			}
			if matchesIconMatch(category.Match, ext, lowerMime) {
				return fileDisplayIcon{
					Icon:   strings.TrimSpace(category.Icon),
					Label:  strings.TrimSpace(category.Label),
					Source: "emoji_category",
				}
			}
		}
	}

	if emojiOptions.Enabled && emojiOptions.ShowInList && emojiOptions.FallbackUnknown {
		return fileDisplayIcon{Icon: emojiOptions.UnknownEmoji, Label: "Unknown file", Source: "emoji"}
	}
	return fileDisplayIcon{}
}

func matchesIconMatch(raw string, ext string, lowerMime string) bool {
	for _, token := range splitIconMatch(raw) {
		if token == "" {
			continue
		}
		if token == ext || token == lowerMime || (strings.HasSuffix(token, "/*") && strings.HasPrefix(lowerMime, strings.TrimSuffix(token, "*"))) {
			return true
		}
	}
	return false
}

func splitIconMatch(raw string) []string {
	normalized := strings.NewReplacer("，", ",", "；", ",", ";", ",", "|", ",", " ", ",").Replace(raw)
	parts := strings.Split(normalized, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.ToLower(strings.TrimSpace(part))
		value = strings.TrimPrefix(value, ".")
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func (c *FileController) getFileSystemSettings() *service.FileSystemSettingPayload {
	if c.fileSystemSettingService == nil {
		return nil
	}

	settings, err := c.fileSystemSettingService.Get()
	if err != nil {
		return nil
	}

	return settings
}

func parseMimeMap(settings *service.FileSystemSettingPayload) map[string]string {
	if settings == nil || strings.TrimSpace(settings.MimeMap) == "" {
		return nil
	}

	result := make(map[string]string)
	_ = json.Unmarshal([]byte(settings.MimeMap), &result)
	return result
}

func resolveMimeTypeFromMap(fileName string, mimeMap map[string]string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		return ""
	}

	if len(mimeMap) > 0 {
		if resolved := strings.TrimSpace(mimeMap[ext]); resolved != "" {
			return resolved
		}
	}

	return strings.TrimSpace(mimetype.FromFileName(fileName))
}

func isGenericListContentType(contentType string) bool {
	return mimetype.IsGeneric(contentType)
}

func onlineEditorLimitBytes(settings *service.FileSystemSettingPayload) int64 {
	if settings == nil || settings.OnlineEditorSize <= 0 {
		return 0
	}

	multiplier := int64(1)
	switch strings.ToUpper(strings.TrimSpace(settings.OnlineEditorUnit)) {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return int64(settings.OnlineEditorSize) * multiplier
}
