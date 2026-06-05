package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"xingyunpan-v2/internal/service"
	"xingyunpan-v2/pkg/response"
)

type OAuthSessionController struct {
	service service.OAuthSessionService
}

func NewOAuthSessionController(oauthSessionService service.OAuthSessionService) *OAuthSessionController {
	return &OAuthSessionController{service: oauthSessionService}
}

func (c *OAuthSessionController) Authorize(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		response.Unauthorized(ctx, "unauthorized")
		return
	}
	result, err := c.service.Authorize(ctx, service.OAuthAuthorizeRequest{
		ClientID:     strings.TrimSpace(ctx.Query("client_id")),
		RedirectURI:  strings.TrimSpace(ctx.Query("redirect_uri")),
		ResponseType: strings.TrimSpace(ctx.Query("response_type")),
		Scope:        strings.TrimSpace(ctx.Query("scope")),
		State:        strings.TrimSpace(ctx.Query("state")),
		UserID:       userID,
	})
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	if strings.EqualFold(ctx.Query("redirect"), "1") || strings.Contains(ctx.GetHeader("Accept"), "text/html") {
		ctx.Redirect(http.StatusFound, result.Location)
		return
	}
	response.Success(ctx, result)
}

func (c *OAuthSessionController) Token(ctx *gin.Context) {
	req := parseOAuthTokenRequest(ctx)
	var (
		result *service.OAuthTokenResponse
		err    error
	)
	switch strings.TrimSpace(req.GrantType) {
	case "authorization_code":
		result, err = c.service.ExchangeToken(ctx, req)
	case "refresh_token":
		result, err = c.service.RefreshToken(ctx, req)
	default:
		response.Error(ctx, http.StatusBadRequest, "unsupported grant_type")
		return
	}
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *OAuthSessionController) Refresh(ctx *gin.Context) {
	req := parseOAuthTokenRequest(ctx)
	req.GrantType = "refresh_token"
	result, err := c.service.RefreshToken(ctx, req)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *OAuthSessionController) UserInfo(ctx *gin.Context) {
	token := bearerToken(ctx)
	if token == "" {
		response.Unauthorized(ctx, "missing access token")
		return
	}
	result, err := c.service.UserInfo(ctx, token)
	if err != nil {
		response.Unauthorized(ctx, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func parseOAuthTokenRequest(ctx *gin.Context) service.OAuthTokenRequest {
	var body struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		RefreshToken string `json:"refresh_token"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
	}
	if strings.Contains(strings.ToLower(ctx.GetHeader("Content-Type")), "application/json") {
		_ = ctx.ShouldBindJSON(&body)
	}
	req := service.OAuthTokenRequest{
		GrantType:    firstRequestValue(body.GrantType, ctx.PostForm("grant_type"), ctx.Query("grant_type")),
		Code:         firstRequestValue(body.Code, ctx.PostForm("code"), ctx.Query("code")),
		RefreshToken: firstRequestValue(body.RefreshToken, ctx.PostForm("refresh_token"), ctx.Query("refresh_token")),
		ClientID:     firstRequestValue(body.ClientID, ctx.PostForm("client_id"), ctx.Query("client_id")),
		ClientSecret: firstRequestValue(body.ClientSecret, ctx.PostForm("client_secret"), ctx.Query("client_secret")),
		RedirectURI:  firstRequestValue(body.RedirectURI, ctx.PostForm("redirect_uri"), ctx.Query("redirect_uri")),
	}
	if req.ClientID == "" || req.ClientSecret == "" {
		if id, secret, ok := ctx.Request.BasicAuth(); ok {
			if req.ClientID == "" {
				req.ClientID = id
			}
			if req.ClientSecret == "" {
				req.ClientSecret = secret
			}
		}
	}
	return req
}

func bearerToken(ctx *gin.Context) string {
	auth := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if auth == "" {
		return strings.TrimSpace(ctx.Query("access_token"))
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func firstRequestValue(values ...string) string {
	for _, value := range values {
		if clean := strings.TrimSpace(value); clean != "" {
			return clean
		}
	}
	return ""
}
