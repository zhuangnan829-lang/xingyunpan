package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
)

func newOAuthSessionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "oauth-session.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.OAuthApp{},
		&model.OAuthAuthorizationCode{},
		&model.OAuthAccessToken{},
		&model.OAuthRefreshToken{},
		&model.OAuthGrant{},
		&model.OAuthAuditLog{},
	); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func seedOAuthSessionApp(t *testing.T, db *gorm.DB) (model.OAuthApp, string) {
	t.Helper()
	plain := "client-secret"
	hash, err := hashOAuthSecret(plain)
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}
	app := model.OAuthApp{
		Slug:                   "test-app",
		Name:                   "Test App",
		AppName:                "Test App",
		ClientID:               "client-id",
		ClientSecret:           hash,
		RedirectURIsJSON:       mustJSON([]string{"http://127.0.0.1/callback"}),
		ScopesJSON:             mustJSON([]string{"openid", "email", "profile", "offline_access", "Files.Read"}),
		PermissionsJSON:        "[]",
		Enabled:                true,
		TokenTTL:               "1 hour",
		RefreshTokenTTLSeconds: 3600,
	}
	if err := db.Create(&app).Error; err != nil {
		t.Fatalf("seed app: %v", err)
	}
	return app, plain
}

func seedOAuthSessionUser(t *testing.T, db *gorm.DB) model.User {
	t.Helper()
	user := model.User{Username: "alice", Email: "alice@example.com", Password: "x", Enabled: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}

func TestOAuthAuthorizationCodeFlowIsSingleUse(t *testing.T) {
	db := newOAuthSessionTestDB(t)
	_, secret := seedOAuthSessionApp(t, db)
	user := seedOAuthSessionUser(t, db)
	svc := NewOAuthSessionService(db)

	auth, err := svc.Authorize(context.Background(), OAuthAuthorizeRequest{
		ClientID:     "client-id",
		RedirectURI:  "http://127.0.0.1/callback",
		ResponseType: "code",
		Scope:        "openid email profile offline_access",
		State:        "state-1",
		UserID:       user.ID,
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if !strings.Contains(auth.Location, "code=") || !strings.Contains(auth.Location, "state=state-1") {
		t.Fatalf("unexpected redirect location: %s", auth.Location)
	}

	token, err := svc.ExchangeToken(context.Background(), OAuthTokenRequest{
		GrantType:    "authorization_code",
		Code:         auth.Code,
		ClientID:     "client-id",
		ClientSecret: secret,
		RedirectURI:  "http://127.0.0.1/callback",
	})
	if err != nil {
		t.Fatalf("exchange token: %v", err)
	}
	if token.AccessToken == "" || token.RefreshToken == "" || token.TokenType != "Bearer" {
		t.Fatalf("unexpected token response: %#v", token)
	}
	if _, err := svc.ExchangeToken(context.Background(), OAuthTokenRequest{
		GrantType:    "authorization_code",
		Code:         auth.Code,
		ClientID:     "client-id",
		ClientSecret: secret,
		RedirectURI:  "http://127.0.0.1/callback",
	}); err == nil {
		t.Fatalf("expected reused authorization code to fail")
	}
}

func TestOAuthAuthorizeRejectsBadRedirectAndScope(t *testing.T) {
	db := newOAuthSessionTestDB(t)
	seedOAuthSessionApp(t, db)
	user := seedOAuthSessionUser(t, db)
	svc := NewOAuthSessionService(db)

	_, err := svc.Authorize(context.Background(), OAuthAuthorizeRequest{
		ClientID:     "client-id",
		RedirectURI:  "http://evil.example/callback",
		ResponseType: "code",
		Scope:        "openid",
		UserID:       user.ID,
	})
	if err == nil {
		t.Fatalf("expected bad redirect_uri to fail")
	}

	_, err = svc.Authorize(context.Background(), OAuthAuthorizeRequest{
		ClientID:     "client-id",
		RedirectURI:  "http://127.0.0.1/callback",
		ResponseType: "code",
		Scope:        "Admin.Read",
		UserID:       user.ID,
	})
	if err == nil {
		t.Fatalf("expected disallowed scope to fail")
	}
}

func TestOAuthRefreshTokenRotatesAndDisabledAppFails(t *testing.T) {
	db := newOAuthSessionTestDB(t)
	app, secret := seedOAuthSessionApp(t, db)
	user := seedOAuthSessionUser(t, db)
	svc := NewOAuthSessionService(db)
	auth, err := svc.Authorize(context.Background(), OAuthAuthorizeRequest{
		ClientID:     app.ClientID,
		RedirectURI:  "http://127.0.0.1/callback",
		ResponseType: "code",
		Scope:        "openid offline_access",
		UserID:       user.ID,
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	token, err := svc.ExchangeToken(context.Background(), OAuthTokenRequest{
		GrantType:    "authorization_code",
		Code:         auth.Code,
		ClientID:     app.ClientID,
		ClientSecret: secret,
		RedirectURI:  "http://127.0.0.1/callback",
	})
	if err != nil {
		t.Fatalf("exchange token: %v", err)
	}
	rotated, err := svc.RefreshToken(context.Background(), OAuthTokenRequest{
		ClientID:     app.ClientID,
		ClientSecret: secret,
		RefreshToken: token.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh token: %v", err)
	}
	if rotated.RefreshToken == "" || rotated.RefreshToken == token.RefreshToken {
		t.Fatalf("expected rotated refresh token, got %#v", rotated)
	}
	if _, err := svc.RefreshToken(context.Background(), OAuthTokenRequest{
		ClientID:     app.ClientID,
		ClientSecret: secret,
		RefreshToken: token.RefreshToken,
	}); err == nil {
		t.Fatalf("expected old refresh token to fail")
	}
	if err := db.Model(&model.OAuthApp{}).Where("id = ?", app.ID).Update("enabled", false).Error; err != nil {
		t.Fatalf("disable app: %v", err)
	}
	if _, err := svc.RefreshToken(context.Background(), OAuthTokenRequest{
		ClientID:     app.ClientID,
		ClientSecret: secret,
		RefreshToken: rotated.RefreshToken,
	}); err == nil {
		t.Fatalf("expected disabled app refresh to fail")
	}
}

func TestOAuthUserInfoHonorsScopes(t *testing.T) {
	db := newOAuthSessionTestDB(t)
	app, _ := seedOAuthSessionApp(t, db)
	user := seedOAuthSessionUser(t, db)
	svc := NewOAuthSessionService(db)
	access := "access-token"
	if err := db.Create(&model.OAuthAccessToken{
		TokenHash:  hashOpaqueToken(access),
		AppID:      app.ID,
		UserID:     user.ID,
		ScopesJSON: mustJSON([]string{"openid", "email"}),
		ExpiresAt:  time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("seed access token: %v", err)
	}

	info, err := svc.UserInfo(context.Background(), access)
	if err != nil {
		t.Fatalf("userinfo: %v", err)
	}
	if info["sub"] != "1" || info["email"] != "alice@example.com" {
		t.Fatalf("unexpected userinfo: %#v", info)
	}
	if _, ok := info["nickname"]; ok {
		t.Fatalf("profile fields should not be returned without profile scope: %#v", info)
	}
}

func TestAdminOAuthAppSecretIsReturnedOnceAndStoredHashed(t *testing.T) {
	db := newOAuthSessionTestDB(t)
	svc := NewAdminOAuthAppService(db)
	created, err := svc.Create(context.Background(), &AdminOAuthAppPayload{
		Name:                   "Custom",
		AppName:                "Custom",
		RedirectURIs:           []string{"http://127.0.0.1/callback"},
		Scopes:                 []string{"openid"},
		RefreshTokenTTLSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	if created.ClientSecret == "" {
		t.Fatalf("expected one-time client secret in response")
	}
	var row model.OAuthApp
	if err := db.First(&row, created.ID).Error; err != nil {
		t.Fatalf("load app: %v", err)
	}
	if row.ClientSecret == created.ClientSecret || !strings.HasPrefix(row.ClientSecret, "$2") {
		t.Fatalf("client secret was not stored as bcrypt hash: %q", row.ClientSecret)
	}
	loaded, err := svc.Get(context.Background(), created.Slug)
	if err != nil {
		t.Fatalf("get app: %v", err)
	}
	if loaded.ClientSecret != "" {
		t.Fatalf("stored secret should not be returned after creation")
	}
}
