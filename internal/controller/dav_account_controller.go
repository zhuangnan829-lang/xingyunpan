package controller

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type DavAccountController struct {
	service service.DavAccountService
}

func NewDavAccountController(service service.DavAccountService) *DavAccountController {
	return &DavAccountController{service: service}
}

func (c *DavAccountController) List(ctx *gin.Context) {
	userID, ok := currentDavUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}
	items, err := c.service.List(ctx, userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, items)
}

func (c *DavAccountController) Create(ctx *gin.Context) {
	userID, ok := currentDavUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}
	var req service.DavAccountUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	account, secret, err := c.service.Create(ctx, userID, req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "WebDAV 账号已创建", gin.H{
		"account": account,
		"secret":  secret,
	})
}

func (c *DavAccountController) Update(ctx *gin.Context) {
	userID, ok := currentDavUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}
	id, err := service.ParseDavAccountID(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	var req service.DavAccountUpsertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	account, err := c.service.Update(ctx, userID, id, req)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(ctx, account)
}

func (c *DavAccountController) Delete(ctx *gin.Context) {
	userID, ok := currentDavUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}
	id, err := service.ParseDavAccountID(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if err := c.service.Delete(ctx, userID, id); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "WebDAV 账号已删除", gin.H{"deleted": true})
}

func (c *DavAccountController) ResetSecret(ctx *gin.Context) {
	userID, ok := currentDavUserID(ctx)
	if !ok {
		response.Error(ctx, http.StatusUnauthorized, "未授权")
		return
	}
	id, err := service.ParseDavAccountID(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	account, secret, err := c.service.ResetSecret(ctx, userID, id)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(ctx, "WebDAV 密钥已重置", gin.H{
		"account": account,
		"secret":  secret,
	})
}

func (c *DavAccountController) DavProbe(ctx *gin.Context) {
	accountToken := ctx.Param("accountToken")
	secret := basicAuthPassword(ctx)
	account, err := c.service.ResolveByToken(ctx, accountToken, secret, ctx.ClientIP())
	if err != nil {
		ctx.Header("WWW-Authenticate", `Basic realm="Xingyunpan WebDAV"`)
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	response.Success(ctx, gin.H{
		"message": "WebDAV 账号可用。完整文件挂载协议将在客户端接入层处理。",
		"account": account,
	})
}

func currentDavUserID(ctx *gin.Context) (uint, bool) {
	value, ok := ctx.Get("user_id")
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok && userID > 0
}

func basicAuthPassword(ctx *gin.Context) string {
	_, password, ok := ctx.Request.BasicAuth()
	if ok {
		return password
	}
	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if !strings.HasPrefix(header, "Basic ") {
		return strings.TrimSpace(ctx.Query("secret"))
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, "Basic "))
	if err != nil {
		return ""
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
