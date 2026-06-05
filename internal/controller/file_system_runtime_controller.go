package controller

import (
	"fmt"
	"net/http"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

type FileSystemRuntimeController struct {
	service service.FileSystemRuntimeService
}

func NewFileSystemRuntimeController(fileSystemRuntimeService service.FileSystemRuntimeService) *FileSystemRuntimeController {
	return &FileSystemRuntimeController{service: fileSystemRuntimeService}
}

func (c *FileSystemRuntimeController) MapConfig(ctx *gin.Context) {
	data, err := c.service.GetMapRuntimeConfig()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *FileSystemRuntimeController) MasterKeyStatus(ctx *gin.Context) {
	response.Success(ctx, c.service.GetMasterKeyStatus())
}

func (c *FileSystemRuntimeController) CreatePackageDownloadSession(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Unauthorized(ctx, err.Error())
		return
	}
	var req struct {
		FileIDs []uint `json:"file_ids"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	data, err := c.service.CreatePackageDownloadSession(ctx.Request.Context(), userID, req.FileIDs)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *FileSystemRuntimeController) DownloadPackage(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Unauthorized(ctx, err.Error())
		return
	}
	sessionID := ctx.Param("sessionId")
	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="package-%s.zip"`, sessionID))
	if _, err := c.service.WritePackageDownload(ctx.Request.Context(), userID, sessionID, ctx.Writer); err != nil {
		ctx.Status(http.StatusBadRequest)
		_, _ = ctx.Writer.WriteString(err.Error())
		return
	}
}

func (c *FileSystemRuntimeController) ExtractArchive(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Unauthorized(ctx, err.Error())
		return
	}
	var req struct {
		SourceFileID   uint  `json:"source_file_id"`
		TargetFolderID *uint `json:"target_folder_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	data, err := c.service.ExtractArchive(ctx.Request.Context(), userID, req.SourceFileID, req.TargetFolderID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *FileSystemRuntimeController) CreateWOPISession(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Unauthorized(ctx, err.Error())
		return
	}
	var req struct {
		FileID uint `json:"file_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	data, err := c.service.CreateWOPISession(ctx.Request.Context(), userID, req.FileID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *FileSystemRuntimeController) GetWOPISession(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Unauthorized(ctx, err.Error())
		return
	}
	data, err := c.service.GetWOPISession(ctx.Request.Context(), userID, ctx.Param("sessionId"))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *FileSystemRuntimeController) SignSlaveRequest(ctx *gin.Context) {
	var req struct {
		Method     string `json:"method"`
		Path       string `json:"path"`
		BodySHA256 string `json:"body_sha256"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	data, err := c.service.SignSlaveRequest(req.Method, req.Path, req.BodySHA256)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *FileSystemRuntimeController) VerifySlaveRequest(ctx *gin.Context) {
	var req struct {
		Method     string `json:"method"`
		Path       string `json:"path"`
		BodySHA256 string `json:"body_sha256"`
		Timestamp  int64  `json:"timestamp"`
		Signature  string `json:"signature"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	data, err := c.service.VerifySlaveRequest(req.Method, req.Path, req.BodySHA256, req.Timestamp, req.Signature)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *FileSystemRuntimeController) OAuthRefreshStatus(ctx *gin.Context) {
	data, err := c.service.GetOAuthRefreshStatus()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}

func (c *FileSystemRuntimeController) RunOAuthRefresh(ctx *gin.Context) {
	data, err := c.service.RunOAuthRefresh()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, data)
}
