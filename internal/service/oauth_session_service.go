package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"xingyunpan-v2/internal/model"
)

const (
	oauthAccessTokenTTL = time.Hour
	oauthCodeTTL        = 10 * time.Minute
)

var BuiltinOAuthScopes = []string{
	"openid", "email", "profile", "offline_access",
	"UserInfo.Read", "UserInfo.Write",
	"UserSecurityInfo.Read", "UserSecurityInfo.Write",
	"Files.Read", "Files.Write",
	"Shares.Read", "Shares.Write",
	"Workflow.Read", "Workflow.Write",
	"Finance.Read", "Finance.Write",
	"DavAccount.Read", "DavAccount.Write",
	"Admin.Read",
}

type OAuthAuthorizeRequest struct {
	ClientID     string
	RedirectURI  string
	ResponseType string
	Scope        string
	State        string
	UserID       uint
}

type OAuthAuthorizeResponse struct {
	Code        string `json:"code"`
	State       string `json:"state,omitempty"`
	RedirectURI string `json:"redirect_uri"`
	Location    string `json:"location"`
	ExpiresIn   int64  `json:"expires_in"`
}

type OAuthTokenRequest struct {
	GrantType    string
	Code         string
	RefreshToken string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

type OAuthUserInfoResponse map[string]interface{}

type OAuthSessionService interface {
	Authorize(ctx context.Context, req OAuthAuthorizeRequest) (*OAuthAuthorizeResponse, error)
	ExchangeToken(ctx context.Context, req OAuthTokenRequest) (*OAuthTokenResponse, error)
	RefreshToken(ctx context.Context, req OAuthTokenRequest) (*OAuthTokenResponse, error)
	UserInfo(ctx context.Context, accessToken string) (OAuthUserInfoResponse, error)
	EnsureTables(ctx context.Context) error
}

type oauthSessionService struct {
	db *gorm.DB
}

func NewOAuthSessionService(db *gorm.DB) OAuthSessionService {
	return &oauthSessionService{db: db}
}

func (s *oauthSessionService) EnsureTables(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(
		&model.OAuthAuthorizationCode{},
		&model.OAuthAccessToken{},
		&model.OAuthRefreshToken{},
		&model.OAuthGrant{},
		&model.OAuthAuditLog{},
	)
}

func (s *oauthSessionService) Authorize(ctx context.Context, req OAuthAuthorizeRequest) (*OAuthAuthorizeResponse, error) {
	if err := s.EnsureTables(ctx); err != nil {
		return nil, err
	}
	if req.UserID == 0 {
		return nil, fmt.Errorf("user is required")
	}
	if strings.TrimSpace(req.ResponseType) != "code" {
		return nil, fmt.Errorf("unsupported response_type")
	}
	app, err := s.findEnabledAppByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}
	if !redirectURIAllowed(app, req.RedirectURI) {
		return nil, fmt.Errorf("redirect_uri is not allowed")
	}
	scopes, err := normalizeRequestedScopes(req.Scope, parseJSONStringSlice(app.ScopesJSON))
	if err != nil {
		return nil, err
	}
	if err := s.ensureUserEnabled(ctx, req.UserID); err != nil {
		return nil, err
	}

	code, err := secureOpaqueToken(32)
	if err != nil {
		return nil, err
	}
	row := &model.OAuthAuthorizationCode{
		CodeHash:    hashOpaqueToken(code),
		AppID:       app.ID,
		UserID:      req.UserID,
		RedirectURI: strings.TrimSpace(req.RedirectURI),
		ScopesJSON:  mustJSON(scopes),
		State:       strings.TrimSpace(req.State),
		ExpiresAt:   time.Now().Add(oauthCodeTTL),
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		if err := upsertOAuthGrant(tx, app.ID, req.UserID, scopes); err != nil {
			return err
		}
		return recordOAuthAudit(tx, 0, app.ID, req.UserID, "oauth.app.authorized", map[string]interface{}{"scopes": scopes})
	})
	if err != nil {
		return nil, err
	}

	location := appendOAuthRedirectParam(req.RedirectURI, code, req.State)
	return &OAuthAuthorizeResponse{
		Code:        code,
		State:       strings.TrimSpace(req.State),
		RedirectURI: strings.TrimSpace(req.RedirectURI),
		Location:    location,
		ExpiresIn:   int64(oauthCodeTTL.Seconds()),
	}, nil
}

func (s *oauthSessionService) ExchangeToken(ctx context.Context, req OAuthTokenRequest) (*OAuthTokenResponse, error) {
	if strings.TrimSpace(req.GrantType) != "authorization_code" {
		return nil, fmt.Errorf("unsupported grant_type")
	}
	if err := s.EnsureTables(ctx); err != nil {
		return nil, err
	}
	app, err := s.authenticateClient(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var code model.OAuthAuthorizationCode
	if err := s.db.WithContext(ctx).Where("code_hash = ?", hashOpaqueToken(req.Code)).First(&code).Error; err != nil {
		return nil, fmt.Errorf("authorization code is invalid")
	}
	if code.AppID != app.ID || code.UsedAt != nil || now.After(code.ExpiresAt) {
		return nil, fmt.Errorf("authorization code is invalid or expired")
	}
	if strings.TrimSpace(req.RedirectURI) != code.RedirectURI {
		return nil, fmt.Errorf("redirect_uri does not match authorization code")
	}
	if err := s.ensureUserAndAppStillValid(ctx, app.ID, code.UserID); err != nil {
		return nil, err
	}
	scopes := parseJSONStringSlice(code.ScopesJSON)
	token, refresh, err := s.issueTokens(ctx, app, code.UserID, scopes)
	if err != nil {
		return nil, err
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		code.UsedAt = &now
		if err := tx.Save(&code).Error; err != nil {
			return err
		}
		return recordOAuthAudit(tx, 0, app.ID, code.UserID, "oauth.token.exchanged", map[string]interface{}{"scopes": scopes})
	})
	if err != nil {
		return nil, err
	}
	return tokenResponse(token, refresh, scopes), nil
}

func (s *oauthSessionService) RefreshToken(ctx context.Context, req OAuthTokenRequest) (*OAuthTokenResponse, error) {
	if strings.TrimSpace(req.GrantType) != "" && strings.TrimSpace(req.GrantType) != "refresh_token" {
		return nil, fmt.Errorf("unsupported grant_type")
	}
	if err := s.EnsureTables(ctx); err != nil {
		return nil, err
	}
	app, err := s.authenticateClient(ctx, req.ClientID, req.ClientSecret)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var old model.OAuthRefreshToken
	if err := s.db.WithContext(ctx).Where("token_hash = ?", hashOpaqueToken(req.RefreshToken)).First(&old).Error; err != nil {
		return nil, fmt.Errorf("refresh_token is invalid")
	}
	if old.AppID != app.ID || old.RevokedAt != nil || old.UsedAt != nil || now.After(old.ExpiresAt) {
		return nil, fmt.Errorf("refresh_token is invalid or expired")
	}
	if err := s.ensureUserAndAppStillValid(ctx, app.ID, old.UserID); err != nil {
		return nil, err
	}
	scopes := parseJSONStringSlice(old.ScopesJSON)
	if !scopesAllowed(scopes, parseJSONStringSlice(app.ScopesJSON)) {
		return nil, fmt.Errorf("refresh_token scopes are no longer allowed")
	}
	token, refresh, err := s.issueTokens(ctx, app, old.UserID, scopes)
	if err != nil {
		return nil, err
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		old.UsedAt = &now
		old.RevokedAt = &now
		if err := tx.Save(&old).Error; err != nil {
			return err
		}
		return recordOAuthAudit(tx, 0, app.ID, old.UserID, "oauth.token.refreshed", map[string]interface{}{"scopes": scopes})
	})
	if err != nil {
		return nil, err
	}
	return tokenResponse(token, refresh, scopes), nil
}

func (s *oauthSessionService) UserInfo(ctx context.Context, accessToken string) (OAuthUserInfoResponse, error) {
	if err := s.EnsureTables(ctx); err != nil {
		return nil, err
	}
	now := time.Now()
	var token model.OAuthAccessToken
	if err := s.db.WithContext(ctx).Where("token_hash = ?", hashOpaqueToken(accessToken)).First(&token).Error; err != nil {
		return nil, fmt.Errorf("access_token is invalid")
	}
	if token.RevokedAt != nil || now.After(token.ExpiresAt) {
		return nil, fmt.Errorf("access_token is expired or revoked")
	}
	if err := s.ensureUserAndAppStillValid(ctx, token.AppID, token.UserID); err != nil {
		return nil, err
	}
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, token.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}
	scopes := scopeSet(parseJSONStringSlice(token.ScopesJSON))
	info := OAuthUserInfoResponse{}
	if scopes["openid"] {
		info["sub"] = fmt.Sprintf("%d", user.ID)
	}
	if scopes["email"] {
		info["email"] = user.Email
		info["email_verified"] = true
	}
	if scopes["profile"] {
		info["name"] = user.Username
		info["nickname"] = user.Username
		info["avatar"] = user.AvatarURL
		info["picture"] = user.AvatarURL
	}
	return info, nil
}

func (s *oauthSessionService) findEnabledAppByClientID(ctx context.Context, clientID string) (*model.OAuthApp, error) {
	var app model.OAuthApp
	if err := s.db.WithContext(ctx).Where("client_id = ?", strings.TrimSpace(clientID)).First(&app).Error; err != nil {
		return nil, fmt.Errorf("oauth client not found")
	}
	if !app.Enabled {
		return nil, fmt.Errorf("oauth client is disabled")
	}
	return &app, nil
}

func (s *oauthSessionService) authenticateClient(ctx context.Context, clientID, secret string) (*model.OAuthApp, error) {
	app, err := s.findEnabledAppByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	stored := strings.TrimSpace(app.ClientSecret)
	if stored == "" {
		return nil, fmt.Errorf("client_secret is not configured")
	}
	if !verifyOAuthSecret(stored, strings.TrimSpace(secret)) {
		return nil, fmt.Errorf("client_secret is invalid")
	}
	return app, nil
}

func (s *oauthSessionService) ensureUserEnabled(ctx context.Context, userID uint) error {
	var user model.User
	if err := s.db.WithContext(ctx).Select("id", "enabled").First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found")
	}
	if !user.Enabled {
		return fmt.Errorf("user is disabled")
	}
	return nil
}

func (s *oauthSessionService) ensureUserAndAppStillValid(ctx context.Context, appID, userID uint) error {
	var app model.OAuthApp
	if err := s.db.WithContext(ctx).First(&app, appID).Error; err != nil {
		return fmt.Errorf("oauth client not found")
	}
	if !app.Enabled {
		return fmt.Errorf("oauth client is disabled")
	}
	return s.ensureUserEnabled(ctx, userID)
}

func (s *oauthSessionService) issueTokens(ctx context.Context, app *model.OAuthApp, userID uint, scopes []string) (*model.OAuthAccessToken, *model.OAuthRefreshToken, error) {
	access, err := secureOpaqueToken(48)
	if err != nil {
		return nil, nil, err
	}
	refresh, err := secureOpaqueToken(48)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	accessRow := &model.OAuthAccessToken{
		TokenHash:  hashOpaqueToken(access),
		AppID:      app.ID,
		UserID:     userID,
		ScopesJSON: mustJSON(scopes),
		ExpiresAt:  now.Add(oauthAccessTokenTTL),
	}
	refreshTTL := time.Duration(app.RefreshTokenTTLSeconds) * time.Second
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	refreshRow := &model.OAuthRefreshToken{
		TokenHash:  hashOpaqueToken(refresh),
		AppID:      app.ID,
		UserID:     userID,
		ScopesJSON: mustJSON(scopes),
		ExpiresAt:  now.Add(refreshTTL),
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(accessRow).Error; err != nil {
			return err
		}
		if containsString(scopes, "offline_access") {
			if err := tx.Create(refreshRow).Error; err != nil {
				return err
			}
		} else {
			refreshRow = nil
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	accessRow.TokenHash = access
	if refreshRow != nil {
		refreshRow.TokenHash = refresh
	}
	return accessRow, refreshRow, nil
}

func tokenResponse(access *model.OAuthAccessToken, refresh *model.OAuthRefreshToken, scopes []string) *OAuthTokenResponse {
	resp := &OAuthTokenResponse{
		AccessToken: access.TokenHash,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(access.ExpiresAt).Seconds()),
		Scope:       strings.Join(scopes, " "),
	}
	if refresh != nil {
		resp.RefreshToken = refresh.TokenHash
	}
	return resp
}

func normalizeRequestedScopes(raw string, allowed []string) ([]string, error) {
	requested := strings.Fields(strings.TrimSpace(raw))
	if len(requested) == 0 {
		requested = []string{"openid"}
	}
	if !containsString(requested, "openid") {
		requested = append([]string{"openid"}, requested...)
	}
	if !scopesAllowed(requested, allowed) {
		return nil, fmt.Errorf("requested scope is not allowed")
	}
	return normalizeStringList(requested), nil
}

func scopesAllowed(requested, allowed []string) bool {
	allowedSet := scopeSet(allowed)
	builtin := scopeSet(BuiltinOAuthScopes)
	for _, scope := range requested {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if !builtin[scope] || !allowedSet[scope] {
			return false
		}
	}
	return true
}

func redirectURIAllowed(app *model.OAuthApp, uri string) bool {
	uri = strings.TrimSpace(uri)
	for _, allowed := range parseJSONStringSlice(app.RedirectURIsJSON) {
		if subtle.ConstantTimeCompare([]byte(uri), []byte(allowed)) == 1 {
			return true
		}
	}
	return false
}

func verifyOAuthSecret(stored, secret string) bool {
	if strings.HasPrefix(stored, "$2a$") || strings.HasPrefix(stored, "$2b$") || strings.HasPrefix(stored, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(secret)) == nil
	}
	return subtle.ConstantTimeCompare([]byte(stored), []byte(secret)) == 1
}

func hashOAuthSecret(secret string) (string, error) {
	data, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(secret)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func secureOpaqueToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashOpaqueToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func appendOAuthRedirectParam(redirectURI, code, state string) string {
	separator := "?"
	if strings.Contains(redirectURI, "?") {
		separator = "&"
	}
	location := redirectURI + separator + "code=" + queryEscapeLite(code)
	if strings.TrimSpace(state) != "" {
		location += "&state=" + queryEscapeLite(state)
	}
	return location
}

func queryEscapeLite(value string) string {
	replacer := strings.NewReplacer(" ", "%20", "+", "%2B", "&", "%26", "=", "%3D", "?", "%3F", "#", "%23")
	return replacer.Replace(value)
}

func upsertOAuthGrant(tx *gorm.DB, appID, userID uint, scopes []string) error {
	var grant model.OAuthGrant
	err := tx.Where("app_id = ? AND user_id = ?", appID, userID).First(&grant).Error
	if err == nil {
		grant.ScopesJSON = mustJSON(scopes)
		grant.RevokedAt = nil
		return tx.Save(&grant).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return tx.Create(&model.OAuthGrant{AppID: appID, UserID: userID, ScopesJSON: mustJSON(scopes)}).Error
}

func recordOAuthAudit(tx *gorm.DB, actorUserID, appID, userID uint, action string, metadata map[string]interface{}) error {
	raw := "{}"
	if metadata != nil {
		if data, err := json.Marshal(metadata); err == nil {
			raw = string(data)
		}
	}
	return tx.Create(&model.OAuthAuditLog{
		ActorUserID:  actorUserID,
		AppID:        appID,
		UserID:       userID,
		Action:       action,
		MetadataJSON: raw,
	}).Error
}

func scopeSet(scopes []string) map[string]bool {
	result := make(map[string]bool, len(scopes))
	for _, scope := range scopes {
		if clean := strings.TrimSpace(scope); clean != "" {
			result[clean] = true
		}
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}
