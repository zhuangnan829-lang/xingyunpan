package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

// DocumentController handles drive document page endpoints.
type DocumentController struct {
	service service.DocumentService
}

// NewDocumentController creates a document controller.
func NewDocumentController(documentService service.DocumentService) *DocumentController {
	return &DocumentController{service: documentService}
}

// List returns paginated document files for the current user.
func (c *DocumentController) List(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := &service.DocumentListQuery{
		Keyword:  ctx.Query("keyword"),
		Sort:     ctx.Query("sort"),
		Page:     intQuery(ctx, "page", 1),
		PageSize: intQuery(ctx, "page_size", 50),
		Cursor:   ctx.Query("cursor"),
	}

	result, err := c.service.ListDocuments(ctx, userID, query)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}
