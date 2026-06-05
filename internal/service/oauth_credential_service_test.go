package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"xingyunpan-v2/internal/model"
)

func TestOAuthCredentialLocalAdapterRefreshesToken(t *testing.T) {
	svc := &oauthCredentialService{httpClient: &http.Client{Timeout: time.Second}}
	result, err := svc.refreshWithAdapter(context.Background(), &model.OAuthCredential{
		Provider:     "xingyunpan",
		RefreshToken: "refresh-token",
		ScopesJSON:   `["Files.Read"]`,
	})
	if err != nil {
		t.Fatalf("refreshWithAdapter returned error: %v", err)
	}
	if !strings.HasPrefix(result.AccessToken, "xyp_at_") {
		t.Fatalf("expected local access token prefix, got %q", result.AccessToken)
	}
	if result.RefreshToken != "refresh-token" {
		t.Fatalf("expected refresh token to be preserved")
	}
	if result.AccessTokenExpiresAt == nil || time.Until(*result.AccessTokenExpiresAt) <= 0 {
		t.Fatalf("expected future access token expiry")
	}
}

func TestOAuthCredentialHTTPAdapterRefreshesToken(t *testing.T) {
	var sawGrantType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm failed: %v", err)
		}
		sawGrantType = r.Form.Get("grant_type")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "remote-access",
			"refresh_token": "remote-refresh",
			"token_type":    "Bearer",
			"expires_in":    120,
			"scope":         "Files.Read profile",
		})
	}))
	defer server.Close()

	svc := &oauthCredentialService{httpClient: server.Client()}
	result, err := svc.refreshWithAdapter(context.Background(), &model.OAuthCredential{
		Provider:      "custom",
		TokenEndpoint: server.URL,
		RefreshToken:  "old-refresh",
		App: model.OAuthApp{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		},
	})
	if err != nil {
		t.Fatalf("refreshWithAdapter returned error: %v", err)
	}
	if sawGrantType != "refresh_token" {
		t.Fatalf("expected refresh_token grant, got %q", sawGrantType)
	}
	if result.AccessToken != "remote-access" || result.RefreshToken != "remote-refresh" {
		t.Fatalf("unexpected token result: %#v", result)
	}
	if got := parseJSONStringSlice(result.ScopesJSON); len(got) != 2 || got[0] != "Files.Read" || got[1] != "profile" {
		t.Fatalf("unexpected scopes json: %s", result.ScopesJSON)
	}
}
