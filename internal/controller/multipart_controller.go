package controller

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/logger"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MultipartController struct {
	multipartService service.MultipartService
}

func NewMultipartController(multipartService service.MultipartService) *MultipartController {
	return &MultipartController{
		multipartService: multipartService,
	}
}

type InitMultipartUploadRequest struct {
	FileName       string `json:"file_name"`
	LegacyFileName string `json:"filename"`
	FileHash       string `json:"file_hash"`
	LegacyHash     string `json:"hash"`
	FileSize       int64  `json:"file_size" binding:"required,gt=0"`
	ChunkSize      int    `json:"chunk_size"`
	ParentID       *uint  `json:"parent_id"`
	FolderID       *uint  `json:"folder_id"`
}

type RecordChunkRequest struct {
	UploadID        string `json:"upload_id" binding:"required"`
	ChunkNumber     int    `json:"chunk_number" binding:"required,gt=0"`
	ETag            string `json:"etag"`
	Attempt         int    `json:"attempt"`
	ActiveTransfers int    `json:"active_transfers"`
}

type CompleteMultipartRequest struct {
	UploadID string `json:"upload_id" binding:"required"`
	ParentID *uint  `json:"parent_id"`
	FolderID *uint  `json:"folder_id"`
}

func (c *MultipartController) InitMultipartUpload(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req InitMultipartUploadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid params: "+err.Error())
		return
	}
	if req.FileName == "" {
		req.FileName = req.LegacyFileName
	}
	if req.FileHash == "" {
		req.FileHash = req.LegacyHash
	}
	if req.ParentID == nil {
		req.ParentID = req.FolderID
	}
	if req.FileName == "" || req.FileHash == "" {
		response.Error(ctx, http.StatusBadRequest, "file_name and file_hash are required")
		return
	}

	result, err := c.multipartService.InitMultipartUpload(
		ctx.Request.Context(),
		userID,
		req.FileName,
		req.FileHash,
		req.FileSize,
		req.ChunkSize,
		req.ParentID,
	)
	if err != nil {
		logger.Error("init multipart upload failed",
			zap.Uint("user_id", userID),
			zap.String("file_name", req.FileName),
			zap.Error(err))
		if service.IsStoragePolicyValidationError(err) {
			response.Error(ctx, http.StatusBadRequest, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "init upload failed")
		}
		return
	}

	response.Success(ctx, result)
}

func (c *MultipartController) ListUploadTasks(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	status := ctx.Query("status")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	list, total, err := c.multipartService.ListUploadTasks(ctx.Request.Context(), userID, status, page, pageSize)
	if err != nil {
		logger.Error("list multipart upload tasks failed",
			zap.Uint("user_id", userID),
			zap.String("status", status),
			zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "list upload tasks failed")
		return
	}

	response.PageSuccess(ctx, list, total, page, pageSize)
}

func (c *MultipartController) GetPresignedURLs(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	uploadID := ctx.Param("upload_id")
	if uploadID == "" {
		response.Error(ctx, http.StatusBadRequest, "upload id is required")
		return
	}

	result, err := c.multipartService.GetPresignedURLs(ctx.Request.Context(), uploadID, userID)
	if err != nil {
		logger.Error("get presigned urls failed",
			zap.Uint("user_id", userID),
			zap.String("upload_id", uploadID),
			zap.Error(err))
		if err.Error() == "no permission to access this upload task" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "get presigned urls failed")
		}
		return
	}

	response.Success(ctx, buildPresignedURLsPayload(result))
}

func (c *MultipartController) RecordChunkUpload(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req RecordChunkRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid params: "+err.Error())
		return
	}

	err = c.multipartService.RecordChunkUpload(ctx.Request.Context(), req.UploadID, userID, req.ChunkNumber, req.ETag, service.ChunkRecordOptions{
		Attempt:         req.Attempt,
		ActiveTransfers: req.ActiveTransfers,
	})
	if err != nil {
		logger.Error("record multipart chunk failed",
			zap.Uint("user_id", userID),
			zap.String("upload_id", req.UploadID),
			zap.Int("chunk_number", req.ChunkNumber),
			zap.Error(err))
		if err.Error() == "no permission to access this upload task" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else if isMultipartClientLimitError(err) {
			response.Error(ctx, http.StatusBadRequest, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "record chunk failed")
		}
		return
	}

	response.Success(ctx, gin.H{"message": "ok"})
}

func (c *MultipartController) GetCompletedChunks(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	uploadID := ctx.Param("upload_id")
	if uploadID == "" {
		response.Error(ctx, http.StatusBadRequest, "upload id is required")
		return
	}

	chunks, err := c.multipartService.GetCompletedChunks(ctx.Request.Context(), uploadID, userID)
	if err != nil {
		logger.Error("get completed chunks failed",
			zap.Uint("user_id", userID),
			zap.String("upload_id", uploadID),
			zap.Error(err))
		if err.Error() == "no permission to access this upload task" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "get completed chunks failed")
		}
		return
	}

	response.Success(ctx, gin.H{"chunks": chunks, "completed_chunks": chunks})
}

func (c *MultipartController) CompleteMultipartUpload(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CompleteMultipartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid params: "+err.Error())
		return
	}

	parentID := req.ParentID
	if parentID == nil {
		parentID = req.FolderID
	}
	userFile, err := c.multipartService.CompleteMultipartUpload(ctx.Request.Context(), req.UploadID, userID, parentID)
	if err != nil {
		logger.Error("complete multipart upload failed",
			zap.Uint("user_id", userID),
			zap.String("upload_id", req.UploadID),
			zap.Error(err))
		if err.Error() == "no permission to access this upload task" || err.Error() == "no permission to access parent folder" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "complete upload failed")
		}
		return
	}

	response.Success(ctx, buildUserFileMutationResponse(userFile))
}

func buildPresignedURLsPayload(result *service.PresignedURLsResult) gin.H {
	if result == nil {
		return gin.H{"urls": []gin.H{}}
	}

	chunkNumbers := make([]int, 0, len(result.URLs))
	for chunkNumber := range result.URLs {
		chunkNumbers = append(chunkNumbers, chunkNumber)
	}
	sort.Ints(chunkNumbers)

	urls := make([]gin.H, 0, len(chunkNumbers))
	for _, chunkNumber := range chunkNumbers {
		urls = append(urls, gin.H{
			"chunk_number": chunkNumber,
			"url":          result.URLs[chunkNumber],
		})
	}

	return gin.H{
		"upload_id":                 result.UploadID,
		"urls":                      urls,
		"expires_at":                result.ExpiresAt,
		"max_chunk_retry":           result.MaxChunkRetry,
		"transfer_parallelism":      result.TransferParallelism,
		"cache_chunks_for_retry":    result.CacheChunksForRetry,
		"blob_signed_url_ttl":       result.BlobSignedURLTTL,
		"blob_signed_url_reuse_ttl": result.BlobSignedURLReuseTTL,
	}
}

func isMultipartClientLimitError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "exceeds server limit") ||
		strings.Contains(message, "retry cache is disabled") ||
		strings.Contains(message, "invalid chunk number")
}

func (c *MultipartController) CancelMultipartUpload(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	uploadID := ctx.Param("upload_id")
	if uploadID == "" {
		response.Error(ctx, http.StatusBadRequest, "upload id is required")
		return
	}

	err = c.multipartService.CancelMultipartUpload(ctx.Request.Context(), uploadID, userID)
	if err != nil {
		logger.Error("cancel multipart upload failed",
			zap.Uint("user_id", userID),
			zap.String("upload_id", uploadID),
			zap.Error(err))
		if err.Error() == "no permission to access this upload task" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "cancel upload failed")
		}
		return
	}

	response.Success(ctx, gin.H{"message": "cancelled"})
}

func (c *MultipartController) GetUploadProgress(ctx *gin.Context) {
	userID, err := GetCurrentUserID(ctx)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "unauthorized")
		return
	}

	uploadID := ctx.Param("upload_id")
	if uploadID == "" {
		response.Error(ctx, http.StatusBadRequest, "upload id is required")
		return
	}

	progress, err := c.multipartService.GetUploadProgress(ctx.Request.Context(), uploadID, userID)
	if err != nil {
		logger.Error("get upload progress failed",
			zap.Uint("user_id", userID),
			zap.String("upload_id", uploadID),
			zap.Error(err))
		if err.Error() == "no permission to access this upload task" {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "get upload progress failed")
		}
		return
	}

	response.Success(ctx, progress)
}
