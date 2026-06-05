package controller

import (
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type AdminFileController struct {
	service                  service.AdminFileService
	fileSystemSettingService service.FileSystemSettingService
}

func NewAdminFileController(adminFileService service.AdminFileService, fileSystemSettingService ...service.FileSystemSettingService) *AdminFileController {
	var settings service.FileSystemSettingService
	if len(fileSystemSettingService) > 0 {
		settings = fileSystemSettingService[0]
	}
	return &AdminFileController{service: adminFileService, fileSystemSettingService: settings}
}

func (c *AdminFileController) List(ctx *gin.Context) {
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

	var storagePolicyID uint64
	if rawStoragePolicyID := strings.TrimSpace(ctx.Query("storage_policy_id")); rawStoragePolicyID != "" {
		parsed, err := strconv.ParseUint(rawStoragePolicyID, 10, 64)
		if err != nil {
			response.Error(ctx, http.StatusBadRequest, "invalid storage policy id")
			return
		}
		storagePolicyID = parsed
	}

	items, total, err := c.service.List(&service.AdminFileListQuery{
		Page:            pagination.Page,
		PageSize:        pagination.PageSize,
		Cursor:          pagination.Cursor,
		UseCursor:       pagination.UseCursor,
		OwnerID:         uint(ownerID),
		Keyword:         strings.TrimSpace(ctx.Query("keyword")),
		StoragePolicyID: uint(storagePolicyID),
	})
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	c.applyDisplayIcons(items)

	nextCursor := ""
	if pagination.UseCursor && len(items) > 0 {
		nextCursor = service.NextCursorFromUint(items[len(items)-1].ID)
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

func (c *AdminFileController) Get(ctx *gin.Context) {
	fileID, ok := parseUintParam(ctx, "id")
	if !ok {
		response.Error(ctx, http.StatusBadRequest, "invalid file id")
		return
	}

	item, err := c.service.Get(fileID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	c.applyDisplayIcon(item)

	response.Success(ctx, item)
}

func (c *AdminFileController) Import(ctx *gin.Context) {
	ownerID, ok := parseUintForm(ctx, "owner_id")
	if !ok {
		response.Error(ctx, http.StatusBadRequest, "owner_id is required")
		return
	}

	parentID, hasParent, err := optionalUintForm(ctx, "parent_id")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid parent id")
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "failed to parse upload form")
		return
	}

	files := append([]*multipartFileHeaderRef{}, toMultipartRefs(form.File["file"])...)
	files = append(files, toMultipartRefs(form.File["files"])...)
	if len(files) == 0 {
		response.Error(ctx, http.StatusBadRequest, "no file uploaded")
		return
	}

	items := make([]*service.AdminFilePayload, 0, len(files))
	for _, file := range files {
		stream, err := file.Open()
		if err != nil {
			response.Error(ctx, http.StatusBadRequest, "failed to open uploaded file")
			return
		}

		var targetParentID *uint
		if hasParent {
			targetParentID = &parentID
		}

		item, importErr := c.service.Import(ownerID, targetParentID, file.Filename, file.Size, stream)
		closeErr := stream.Close()
		if importErr != nil {
			response.Error(ctx, http.StatusBadRequest, importErr.Error())
			return
		}
		if closeErr != nil {
			response.Error(ctx, http.StatusInternalServerError, "failed to close uploaded file stream")
			return
		}
		c.applyDisplayIcon(item)

		items = append(items, item)
	}

	response.SuccessWithMessage(ctx, "files imported", gin.H{
		"items": items,
		"count": len(items),
	})
}

func (c *AdminFileController) Rename(ctx *gin.Context) {
	fileID, ok := parseUintParam(ctx, "id")
	if !ok {
		response.Error(ctx, http.StatusBadRequest, "invalid file id")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := c.service.Rename(fileID, req.Name); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "file renamed", gin.H{"renamed": true})
}

func (c *AdminFileController) Download(ctx *gin.Context) {
	fileID, ok := parseUintParam(ctx, "id")
	if !ok {
		response.Error(ctx, http.StatusBadRequest, "invalid file id")
		return
	}

	result, err := c.service.DownloadWithDelivery(fileID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
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

	contentType := result.ContentType
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Content-Disposition", `attachment; filename="`+sanitizeDownloadName(result.FileName)+`"`)
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, result.Reader)
}

func (c *AdminFileController) CreateShare(ctx *gin.Context) {
	fileID, ok := parseUintParam(ctx, "id")
	if !ok {
		response.Error(ctx, http.StatusBadRequest, "invalid file id")
		return
	}

	var req struct {
		ExpiresIn  *int    `json:"expires_in"`
		AccessCode *string `json:"access_code"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AccessCode != nil && (len(*req.AccessCode) < 4 || len(*req.AccessCode) > 8) {
		response.Error(ctx, http.StatusBadRequest, "access code length must be between 4 and 8")
		return
	}

	resp, err := c.service.CreateShare(ctx, fileID, req.ExpiresIn, req.AccessCode)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "share created", resp)
}

func (c *AdminFileController) Delete(ctx *gin.Context) {
	fileID, ok := parseUintParam(ctx, "id")
	if !ok {
		response.Error(ctx, http.StatusBadRequest, "invalid file id")
		return
	}

	if err := c.service.Delete(fileID); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "file deleted", gin.H{"deleted": true})
}

func normalizedPageSize(value int) int {
	if value <= 0 {
		return 10
	}
	if value > 100 {
		return 100
	}
	return value
}

func (c *AdminFileController) getFileSystemSettings() *service.FileSystemSettingPayload {
	if c.fileSystemSettingService == nil {
		return nil
	}
	settings, err := c.fileSystemSettingService.Get()
	if err != nil {
		return nil
	}
	return settings
}

func (c *AdminFileController) applyDisplayIcons(items []service.AdminFilePayload) {
	if len(items) == 0 {
		return
	}
	settings := c.getFileSystemSettings()
	iconRules := parseFileIconRules(settings)
	emojiOptions := parseFileEmojiOptions(settings)
	mimeMap := parseMimeMap(settings)
	for index := range items {
		applyAdminFileDisplayIcon(&items[index], iconRules, emojiOptions, mimeMap)
	}
}

func (c *AdminFileController) applyDisplayIcon(item *service.AdminFilePayload) {
	if item == nil {
		return
	}
	settings := c.getFileSystemSettings()
	applyAdminFileDisplayIcon(item, parseFileIconRules(settings), parseFileEmojiOptions(settings), parseMimeMap(settings))
}

func applyAdminFileDisplayIcon(item *service.AdminFilePayload, iconRules []fileIconRule, emojiOptions fileEmojiOptions, mimeMap map[string]string) {
	contentType := strings.TrimSpace(item.ContentType)
	if contentType == "" || isGenericListContentType(contentType) {
		contentType = resolveMimeTypeFromMap(item.FileName, mimeMap)
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	displayIcon := resolveDisplayIcon(item.FileName, item.IsFolder, contentType, iconRules, emojiOptions)
	item.DisplayIcon = displayIcon.Icon
	item.DisplayIconTint = displayIcon.Tint
	item.DisplayIconLabel = displayIcon.Label
	item.DisplayIconSource = displayIcon.Source
}

func max(value, fallback int) int {
	if value < fallback {
		return fallback
	}
	return value
}

func parseUintParam(ctx *gin.Context, key string) (uint, bool) {
	value, err := strconv.ParseUint(strings.TrimSpace(ctx.Param(key)), 10, 64)
	if err != nil || value == 0 {
		return 0, false
	}
	return uint(value), true
}

func parseUintForm(ctx *gin.Context, key string) (uint, bool) {
	value := strings.TrimSpace(ctx.PostForm(key))
	if value == "" {
		return 0, false
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, false
	}

	return uint(parsed), true
}

func optionalUintForm(ctx *gin.Context, key string) (uint, bool, error) {
	value := strings.TrimSpace(ctx.PostForm(key))
	if value == "" {
		return 0, false, nil
	}

	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, false, err
	}

	return uint(parsed), true, nil
}

func sanitizeDownloadName(name string) string {
	safe := strings.ReplaceAll(name, `"`, "")
	safe = strings.ReplaceAll(safe, "\r", "")
	safe = strings.ReplaceAll(safe, "\n", "")
	if safe == "" {
		return "download.bin"
	}
	return safe
}

type multipartFileHeaderRef struct {
	Filename string
	Size     int64
	Open     func() (io.ReadCloser, error)
}

func toMultipartRefs(files []*multipart.FileHeader) []*multipartFileHeaderRef {
	refs := make([]*multipartFileHeaderRef, 0, len(files))
	for _, file := range files {
		current := file
		refs = append(refs, &multipartFileHeaderRef{
			Filename: current.Filename,
			Size:     current.Size,
			Open: func() (io.ReadCloser, error) {
				return current.Open()
			},
		})
	}
	return refs
}
