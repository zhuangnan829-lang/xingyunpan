package controller

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type AdminBlobController struct {
	service service.AdminBlobService
}

func NewAdminBlobController(adminBlobService service.AdminBlobService) *AdminBlobController {
	return &AdminBlobController{service: adminBlobService}
}

func (c *AdminBlobController) List(ctx *gin.Context) {
	query, err := parseAdminBlobListQuery(ctx)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	items, err := c.service.List(query)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, items)
}

func (c *AdminBlobController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Error(ctx, http.StatusBadRequest, "invalid blob id")
		return
	}

	item, err := c.service.Get(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, item)
}

func (c *AdminBlobController) Download(ctx *gin.Context) {
	id, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Error(ctx, http.StatusBadRequest, "invalid blob id")
		return
	}

	reader, fileName, contentType, err := c.service.Download(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	defer reader.Close()

	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Content-Disposition", `attachment; filename="`+sanitizeAdminBlobDownloadName(fileName)+`"`)
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, reader)
}

func (c *AdminBlobController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Error(ctx, http.StatusBadRequest, "invalid blob id")
		return
	}

	if err := c.service.Delete(uint(id)); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "blob deleted", gin.H{"deleted": true})
}

func (c *AdminBlobController) BatchDelete(ctx *gin.Context) {
	var payload struct {
		IDs     []uint `json:"ids"`
		BlobIDs []uint `json:"blob_ids"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	ids := payload.IDs
	if len(ids) == 0 {
		ids = payload.BlobIDs
	}
	result, err := c.service.BatchDelete(ids)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, result)
}

func (c *AdminBlobController) Lock(ctx *gin.Context) {
	id, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Error(ctx, http.StatusBadRequest, "invalid blob id")
		return
	}
	var payload struct {
		Reason string `json:"reason"`
	}
	_ = ctx.ShouldBindJSON(&payload)
	adminID, _ := ctx.Get("user_id")
	item, err := c.service.Lock(uint(id), uintFromContextValue(adminID), payload.Reason)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, item)
}

func (c *AdminBlobController) Unlock(ctx *gin.Context) {
	id, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("id")), 10, 64)
	if err != nil || id == 0 {
		response.Error(ctx, http.StatusBadRequest, "invalid blob id")
		return
	}
	adminID, _ := ctx.Get("user_id")
	item, err := c.service.Unlock(uint(id), uintFromContextValue(adminID))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, item)
}

func (c *AdminBlobController) Scan(ctx *gin.Context) {
	task, err := c.service.Scan()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, task)
}

func (c *AdminBlobController) LatestScan(ctx *gin.Context) {
	task, err := c.service.LatestScan()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, task)
}

func sanitizeAdminBlobDownloadName(name string) string {
	safe := strings.ReplaceAll(name, `"`, "")
	safe = strings.ReplaceAll(safe, "\r", "")
	safe = strings.ReplaceAll(safe, "\n", "")
	if safe == "" {
		return "blob.bin"
	}
	return safe
}

func parseAdminBlobListQuery(ctx *gin.Context) (*service.AdminBlobListQuery, error) {
	query := &service.AdminBlobListQuery{
		Page:      parsePositiveInt(ctx.Query("page"), 1),
		PageSize:  parsePositiveInt(ctx.Query("page_size"), 20),
		Kind:      strings.TrimSpace(ctx.Query("kind")),
		Keyword:   strings.TrimSpace(ctx.Query("keyword")),
		SortBy:    strings.TrimSpace(ctx.Query("sort_by")),
		SortOrder: strings.TrimSpace(ctx.Query("sort_order")),
	}
	var err error
	if query.OwnerID, err = parseOptionalUintQuery(ctx, "owner_id"); err != nil {
		return nil, err
	}
	if query.StoragePolicyID, err = parseOptionalUintQuery(ctx, "storage_policy_id"); err != nil {
		return nil, err
	}
	if query.MinSize, err = parseOptionalInt64Query(ctx, "min_size"); err != nil {
		return nil, err
	}
	if query.MaxSize, err = parseOptionalInt64Query(ctx, "max_size"); err != nil {
		return nil, err
	}
	if query.RefCountMin, err = parseOptionalIntQuery(ctx, "ref_count_min"); err != nil {
		return nil, err
	}
	if query.RefCountMax, err = parseOptionalIntQuery(ctx, "ref_count_max"); err != nil {
		return nil, err
	}
	if query.Encrypted, err = parseOptionalBoolQuery(ctx, "encrypted"); err != nil {
		return nil, err
	}
	if query.CreatedFrom, err = parseOptionalTimeQuery(ctx, "created_from"); err != nil {
		return nil, err
	}
	if query.CreatedTo, err = parseOptionalTimeQuery(ctx, "created_to"); err != nil {
		return nil, err
	}
	return query, nil
}

func parsePositiveInt(raw string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseOptionalUintQuery(ctx *gin.Context, key string) (uint, error) {
	raw := strings.TrimSpace(ctx.Query(key))
	if raw == "" || raw == "all" {
		return 0, nil
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, strconv.ErrSyntax
	}
	return uint(parsed), nil
}

func parseOptionalInt64Query(ctx *gin.Context, key string) (*int64, error) {
	raw := strings.TrimSpace(ctx.Query(key))
	if raw == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalIntQuery(ctx *gin.Context, key string) (*int, error) {
	raw := strings.TrimSpace(ctx.Query(key))
	if raw == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalBoolQuery(ctx *gin.Context, key string) (*bool, error) {
	raw := strings.ToLower(strings.TrimSpace(ctx.Query(key)))
	if raw == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalTimeQuery(ctx *gin.Context, key string) (*time.Time, error) {
	raw := strings.TrimSpace(ctx.Query(key))
	if raw == "" {
		return nil, nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return &parsed, nil
		}
	}
	return nil, strconv.ErrSyntax
}

func uintFromContextValue(value interface{}) uint {
	switch typed := value.(type) {
	case uint:
		return typed
	case uint64:
		return uint(typed)
	case int:
		if typed > 0 {
			return uint(typed)
		}
	}
	return 0
}
