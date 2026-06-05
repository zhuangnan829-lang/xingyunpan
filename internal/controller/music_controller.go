package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// MusicController handles drive music page endpoints.
type MusicController struct {
	service service.MusicService
}

// NewMusicController creates a music controller.
func NewMusicController(musicService service.MusicService) *MusicController {
	return &MusicController{service: musicService}
}

// List returns paginated audio files for the current user.
func (c *MusicController) List(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := &service.MusicListQuery{
		Keyword:  ctx.Query("keyword"),
		Sort:     ctx.Query("sort"),
		Page:     intQuery(ctx, "page", 1),
		PageSize: intQuery(ctx, "page_size", 50),
		Cursor:   ctx.Query("cursor"),
	}

	result, err := c.service.ListMusic(ctx, userID, query)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}
