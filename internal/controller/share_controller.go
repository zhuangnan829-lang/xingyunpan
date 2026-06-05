// 路径: internal/controller/share_controller.go
package controller

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"

	"github.com/gin-gonic/gin"
)

// ShareController 分享控制器
type ShareController struct {
	shareService             service.ShareService
	fileSystemSettingService service.FileSystemSettingService
}

// NewShareController 创建分享控制器实例
func NewShareController(shareService service.ShareService, fileSystemSettingService ...service.FileSystemSettingService) *ShareController {
	var settings service.FileSystemSettingService
	if len(fileSystemSettingService) > 0 {
		settings = fileSystemSettingService[0]
	}
	return &ShareController{
		shareService:             shareService,
		fileSystemSettingService: settings,
	}
}

// CreateShare 创建分享接口
// @Summary 创建分享
// @Tags 分享
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateShareRequest true "创建分享请求"
// @Success 201 {object} response.Response
// @Router /api/shares [post]
func (c *ShareController) CreateShare(ctx *gin.Context) {
	var req struct {
		FileIDs      []string `json:"file_ids" binding:"required,min=1"`
		ExpiresIn    *int     `json:"expires_in"`
		AccessCode   *string  `json:"access_code"`
		MaxDownloads *int     `json:"max_downloads"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 验证 access_code 长度（4-8 字符）
	if req.AccessCode != nil && (len(*req.AccessCode) < 4 || len(*req.AccessCode) > 8) {
		response.Error(ctx, http.StatusBadRequest, "访问密码长度必须为 4-8 字符")
		return
	}

	// 从上下文获取用户 ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	resp, err := c.shareService.CreateShare(ctx, userID.(uint), req.FileIDs, req.ExpiresIn, req.AccessCode, req.MaxDownloads)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    resp,
	})
}

// GetShareInfo 获取分享信息接口
// @Summary 获取分享信息
// @Tags 分享
// @Produce json
// @Param shareId path string true "分享令牌"
// @Success 200 {object} response.Response
// @Router /api/shares/{shareId} [get]
func (c *ShareController) GetShareInfo(ctx *gin.Context) {
	shareToken := ctx.Param("shareId")

	resp, err := c.shareService.GetShareInfo(ctx, shareToken)
	if err != nil {
		if err.Error() == "share expired" {
			response.Error(ctx, http.StatusGone, err.Error())
			return
		}
		if err.Error() == "分享已过期" {
			response.Error(ctx, http.StatusGone, err.Error())
			return
		}
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	response.Success(ctx, resp)
}

// VerifyPassword 验证分享密码接口
// @Summary 验证分享密码
// @Tags 分享
// @Accept json
// @Produce json
// @Param shareId path string true "分享令牌"
// @Param request body VerifyPasswordRequest true "验证密码请求"
// @Success 200 {object} response.Response
// @Router /api/shares/{shareId}/verify [post]
func (c *ShareController) VerifyPassword(ctx *gin.Context) {
	shareToken := ctx.Param("shareId")

	var req struct {
		AccessCode string `json:"access_code" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	clientIP := ctx.ClientIP()
	resp, err := c.shareService.VerifySharePassword(ctx, shareToken, req.AccessCode, clientIP)
	if err != nil {
		if err.Error() == "超过验证次数限制" {
			response.Error(ctx, http.StatusTooManyRequests, err.Error())
			return
		}
		response.Error(ctx, http.StatusUnauthorized, "密码错误")
		return
	}

	if !resp.Valid {
		response.Error(ctx, http.StatusUnauthorized, "密码错误")
		return
	}

	response.Success(ctx, resp)
}

// GetMyShares 获取我的分享列表接口
// @Summary 获取我的分享列表
// @Tags 分享
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /api/shares/me [get]
func (c *ShareController) GetMyShares(ctx *gin.Context) {
	// 从上下文获取用户 ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	resp, err := c.shareService.GetMyShares(ctx, userID.(uint))
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	settings := c.getFileSystemSettings()
	page, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(ctx.DefaultQuery("page_size", "20")))
	pagination := resolveControllerPagination(settings, page, pageSize, ctx.Query("cursor"), 20)
	items, nextCursor := paginateMyShares(resp, pagination)

	response.Success(ctx, gin.H{
		"list":            items,
		"total":           len(resp),
		"page":            pagination.Page,
		"page_size":       pagination.PageSize,
		"pagination_mode": pagination.Mode,
		"next_cursor":     nextCursor,
		"max_page_size":   pagination.MaxPageSize,
	})
}

// DeleteShare 删除分享接口
// @Summary 删除分享
// @Tags 分享
// @Security Bearer
// @Param shareId path string true "分享 ID"
// @Success 204 "删除成功"
// @Router /api/shares/{shareId} [delete]
func (c *ShareController) DeleteShare(ctx *gin.Context) {
	shareIDStr := ctx.Param("shareId")
	shareID, err := strconv.ParseUint(shareIDStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的分享 ID")
		return
	}

	// 从上下文获取用户 ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}

	if err := c.shareService.DeleteShare(ctx, userID.(uint), uint(shareID)); err != nil {
		if err.Error() == "无权限删除此分享" {
			response.Error(ctx, http.StatusForbidden, err.Error())
			return
		}
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// IncrementDownload 增加下载计数接口
// @Summary 增加下载计数
// @Tags 分享
// @Param shareId path string true "分享令牌"
// @Success 204 "成功"
// @Router /api/shares/{shareId}/download [post]
func (c *ShareController) IncrementDownload(ctx *gin.Context) {
	shareToken := ctx.Param("shareId")

	if err := c.shareService.IncrementDownloadCount(ctx, shareToken); err != nil {
		if err.Error() == "share expired" {
			response.Error(ctx, http.StatusGone, err.Error())
			return
		}
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	ctx.Status(http.StatusNoContent)
}

// Download streams shared files to public visitors.
func (c *ShareController) Download(ctx *gin.Context) {
	shareToken := ctx.Param("shareId")

	result, err := c.shareService.DownloadShareWithDelivery(ctx, shareToken)
	if err != nil {
		if err.Error() == "share expired" {
			response.Error(ctx, http.StatusGone, err.Error())
			return
		}
		response.Error(ctx, http.StatusNotFound, err.Error())
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

	ctx.Header("Content-Type", result.ContentType)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(result.FileName)))
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, result.Reader)
}

func (c *ShareController) getFileSystemSettings() *service.FileSystemSettingPayload {
	if c.fileSystemSettingService == nil {
		return nil
	}
	settings, err := c.fileSystemSettingService.Get()
	if err != nil {
		return nil
	}
	return settings
}

func paginateMyShares(items []*service.MyShareResponse, pagination service.PaginationRuntime) ([]*service.MyShareResponse, string) {
	if len(items) == 0 {
		return []*service.MyShareResponse{}, ""
	}
	filtered := items
	if pagination.UseCursor && pagination.Cursor > 0 {
		filtered = make([]*service.MyShareResponse, 0, len(items))
		for _, item := range items {
			if item != nil && item.ShareID < pagination.Cursor {
				filtered = append(filtered, item)
			}
		}
	}
	start := 0
	if !pagination.UseCursor {
		start = (pagination.Page - 1) * pagination.PageSize
	}
	if start >= len(filtered) {
		return []*service.MyShareResponse{}, ""
	}
	end := start + pagination.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	pageItems := filtered[start:end]
	nextCursor := ""
	if pagination.UseCursor && len(pageItems) > 0 {
		nextCursor = service.NextCursorFromUint(pageItems[len(pageItems)-1].ShareID)
	}
	return pageItems, nextCursor
}
