package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// VideoController handles drive video page endpoints.
type VideoController struct {
	service service.VideoService
}

// NewVideoController creates a video controller.
func NewVideoController(videoService service.VideoService) *VideoController {
	return &VideoController{service: videoService}
}

// List returns paginated videos for the current user.
func (c *VideoController) List(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := &service.VideoListQuery{
		Keyword:  ctx.Query("keyword"),
		Sort:     ctx.Query("sort"),
		Page:     intQuery(ctx, "page", 1),
		PageSize: intQuery(ctx, "page_size", 50),
		Cursor:   ctx.Query("cursor"),
	}

	result, err := c.service.ListVideos(ctx, userID, query)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

func intQuery(ctx *gin.Context, key string, fallback int) int {
	raw := ctx.Query(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
