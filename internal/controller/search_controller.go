package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// SearchController handles authenticated file search APIs.
type SearchController struct {
	searchService service.SearchService
}

// NewSearchController creates a controller instance.
func NewSearchController(searchService service.SearchService) *SearchController {
	return &SearchController{searchService: searchService}
}

// SearchFiles supports both legacy GET query search and JSON POST search.
func (c *SearchController) SearchFiles(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "invalid user context")
		return
	}

	params, err := c.parseSearchParams(ctx)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid search request: "+err.Error())
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}

	result, err := c.searchService.SearchFiles(ctx, userID, params)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// GetSuggestions returns lightweight keyword suggestions.
func (c *SearchController) GetSuggestions(ctx *gin.Context) {
	userIDValue, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "invalid user context")
		return
	}

	prefix := ctx.Query("prefix")
	if len(prefix) < 2 {
		response.Error(ctx, http.StatusBadRequest, "prefix must contain at least 2 characters")
		return
	}

	suggestions, err := c.searchService.GetSuggestions(ctx, userID, prefix)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, suggestions)
}

func (c *SearchController) parseSearchParams(ctx *gin.Context) (*service.SearchParams, error) {
	if ctx.Request.Method == http.MethodPost {
		var params service.SearchParams
		if err := ctx.ShouldBindJSON(&params); err != nil {
			return nil, err
		}
		if params.Page == 0 {
			params.Page = 1
		}
		if params.PageSize == 0 {
			params.PageSize = 20
		}
		return &params, nil
	}

	params := &service.SearchParams{
		Keyword:  ctx.Query("keyword"),
		FileType: ctx.Query("file_type"),
		Page:     c.getIntQuery(ctx, "page", 1),
		PageSize: c.getIntQuery(ctx, "page_size", 20),
		Cursor:   ctx.Query("cursor"),
	}

	if sizeMin := ctx.Query("size_min"); sizeMin != "" {
		val, err := strconv.ParseInt(sizeMin, 10, 64)
		if err == nil {
			params.SizeMin = &val
		}
	}

	if sizeMax := ctx.Query("size_max"); sizeMax != "" {
		val, err := strconv.ParseInt(sizeMax, 10, 64)
		if err == nil {
			params.SizeMax = &val
		}
	}

	if dateFrom := ctx.Query("date_from"); dateFrom != "" {
		val, err := time.Parse(time.RFC3339, dateFrom)
		if err == nil {
			params.DateFrom = &val
		}
	}

	if dateTo := ctx.Query("date_to"); dateTo != "" {
		val, err := time.Parse(time.RFC3339, dateTo)
		if err == nil {
			params.DateTo = &val
		}
	}

	if folderID := ctx.Query("folder_id"); folderID != "" {
		params.FolderID = &folderID
	}

	return params, nil
}

func (c *SearchController) getIntQuery(ctx *gin.Context, key string, defaultValue int) int {
	if val := ctx.Query(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}
