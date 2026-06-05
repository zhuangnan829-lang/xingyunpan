package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

type OAuthCredentialPayload struct {
	ID                    uint       `json:"id"`
	AppID                 uint       `json:"app_id"`
	UserID                uint       `json:"user_id"`
	Provider              string     `json:"provider"`
	Subject               string     `json:"subject"`
	TokenType             string     `json:"token_type"`
	Scopes                []string   `json:"scopes"`
	TokenEndpoint         string     `json:"token_endpoint"`
	AccessTokenPreview    string     `json:"access_token_preview"`
	RefreshTokenPreview   string     `json:"refresh_token_preview"`
	AccessTokenExpiresAt  *time.Time `json:"access_token_expires_at,omitempty"`
	RefreshTokenExpiresAt *time.Time `json:"refresh_token_expires_at,omitempty"`
	LastRefreshedAt       *time.Time `json:"last_refreshed_at,omitempty"`
	NextRefreshAt         *time.Time `json:"next_refresh_at,omitempty"`
	Status                string     `json:"status"`
	RefreshError          string     `json:"refresh_error,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type OAuthCredentialInput struct {
	AppID                 uint       `json:"app_id"`
	UserID                uint       `json:"user_id"`
	Provider              string     `json:"provider"`
	Subject               string     `json:"subject"`
	AccessToken           string     `json:"access_token"`
	RefreshToken          string     `json:"refresh_token"`
	TokenType             string     `json:"token_type"`
	Scopes                []string   `json:"scopes"`
	TokenEndpoint         string     `json:"token_endpoint"`
	AccessTokenExpiresAt  *time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt *time.Time `json:"refresh_token_expires_at"`
	MetadataJSON          string     `json:"metadata_json"`
}

type OAuthCredentialListQuery struct {
	Page     int
	PageSize int
	Provider string
	Status   string
	AppID    uint
	UserID   *uint
}

type OAuthCredentialRefreshSummary struct {
	Checked   int                         `json:"checked"`
	Refreshed int                         `json:"refreshed"`
	Failed    int                         `json:"failed"`
	Skipped   int                         `json:"skipped"`
	Items     []OAuthCredentialRunPayload `json:"items"`
}

type OAuthCredentialRunPayload struct {
	ID      uint   `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type OAuthCredentialService interface {
	List(ctx context.Context, query *OAuthCredentialListQuery) ([]OAuthCredentialPayload, int64, error)
	Create(ctx context.Context, input *OAuthCredentialInput) (*OAuthCredentialPayload, error)
	RefreshOne(ctx context.Context, id uint) (*OAuthCredentialPayload, error)
	RefreshDue(ctx context.Context, lookahead time.Duration) (*OAuthCredentialRefreshSummary, error)
	Stats(ctx context.Context, lookahead time.Duration) (total, active, due int64, err error)
}

type oauthCredentialService struct {
	db         *gorm.DB
	httpClient *http.Client
}

func NewOAuthCredentialService(db *gorm.DB) OAuthCredentialService {
	return &oauthCredentialService{
		db:         db,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *oauthCredentialService) List(ctx context.Context, query *OAuthCredentialListQuery) ([]OAuthCredentialPayload, int64, error) {
	if err := s.ensureTable(ctx); err != nil {
		return nil, 0, err
	}
	if query == nil {
		query = &OAuthCredentialListQuery{}
	}
	page := normalizeOAuthPage(query.Page)
	pageSize := normalizeOAuthPageSize(query.PageSize)
	base := s.db.WithContext(ctx).Model(&model.OAuthCredential{})
	if query.AppID > 0 {
		base = base.Where("app_id = ?", query.AppID)
	}
	if query.UserID != nil {
		base = base.Where("user_id = ?", *query.UserID)
	}
	if provider := strings.TrimSpace(query.Provider); provider != "" {
		base = base.Where("provider = ?", provider)
	}
	if status := strings.TrimSpace(query.Status); status != "" {
		base = base.Where("status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.OAuthCredential
	if err := base.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]OAuthCredentialPayload, 0, len(rows))
	for i := range rows {
		items = append(items, toOAuthCredentialPayload(&rows[i]))
	}
	return items, total, nil
}

func (s *oauthCredentialService) Create(ctx context.Context, input *OAuthCredentialInput) (*OAuthCredentialPayload, error) {
	if err := s.ensureTable(ctx); err != nil {
		return nil, err
	}
	if input == nil {
		return nil, fmt.Errorf("credential cannot be nil")
	}
	if input.AppID == 0 {
		return nil, fmt.Errorf("app_id is required")
	}
	var app model.OAuthApp
	if err := s.db.WithContext(ctx).First(&app, input.AppID).Error; err != nil {
		return nil, fmt.Errorf("oauth app not found: %w", err)
	}
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	if provider == "" {
		provider = "xingyunpan"
	}
	tokenType := strings.TrimSpace(input.TokenType)
	if tokenType == "" {
		tokenType = "Bearer"
	}
	if strings.TrimSpace(input.RefreshToken) == "" {
		return nil, fmt.Errorf("refresh_token is required")
	}
	nextRefresh := nextOAuthRefreshTime(input.AccessTokenExpiresAt, 5*time.Minute)
	row := &model.OAuthCredential{
		AppID:                 input.AppID,
		UserID:                input.UserID,
		Provider:              provider,
		Subject:               strings.TrimSpace(input.Subject),
		AccessToken:           strings.TrimSpace(input.AccessToken),
		RefreshToken:          strings.TrimSpace(input.RefreshToken),
		TokenType:             tokenType,
		ScopesJSON:            mustJSON(normalizeStringList(input.Scopes)),
		TokenEndpoint:         strings.TrimSpace(input.TokenEndpoint),
		AccessTokenExpiresAt:  input.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: input.RefreshTokenExpiresAt,
		NextRefreshAt:         nextRefresh,
		Status:                "active",
		MetadataJSON:          normalizeJSON(input.MetadataJSON),
	}
	if row.AccessToken == "" {
		row.AccessToken = "pending"
	}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	payload := toOAuthCredentialPayload(row)
	return &payload, nil
}

func (s *oauthCredentialService) RefreshOne(ctx context.Context, id uint) (*OAuthCredentialPayload, error) {
	if err := s.ensureTable(ctx); err != nil {
		return nil, err
	}
	var credential model.OAuthCredential
	if err := s.db.WithContext(ctx).Preload("App").First(&credential, id).Error; err != nil {
		return nil, err
	}
	if err := s.refreshCredential(ctx, &credential); err != nil {
		return nil, err
	}
	payload := toOAuthCredentialPayload(&credential)
	return &payload, nil
}

func (s *oauthCredentialService) RefreshDue(ctx context.Context, lookahead time.Duration) (*OAuthCredentialRefreshSummary, error) {
	if err := s.ensureTable(ctx); err != nil {
		return nil, err
	}
	if lookahead <= 0 {
		lookahead = 5 * time.Minute
	}
	deadline := time.Now().Add(lookahead)
	var rows []model.OAuthCredential
	if err := s.db.WithContext(ctx).
		Preload("App").
		Where("status = ?", "active").
		Where("refresh_token <> ''").
		Where("(next_refresh_at IS NULL OR next_refresh_at <= ? OR access_token_expires_at IS NULL OR access_token_expires_at <= ?)", deadline, deadline).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	summary := &OAuthCredentialRefreshSummary{Checked: len(rows), Items: make([]OAuthCredentialRunPayload, 0, len(rows))}
	for i := range rows {
		err := s.refreshCredential(ctx, &rows[i])
		if err != nil {
			summary.Failed++
			summary.Items = append(summary.Items, OAuthCredentialRunPayload{ID: rows[i].ID, Status: "failed", Message: err.Error()})
			continue
		}
		summary.Refreshed++
		summary.Items = append(summary.Items, OAuthCredentialRunPayload{ID: rows[i].ID, Status: "refreshed", Message: "access token refreshed"})
	}
	return summary, nil
}

func (s *oauthCredentialService) Stats(ctx context.Context, lookahead time.Duration) (total, active, due int64, err error) {
	if err = s.ensureTable(ctx); err != nil {
		return
	}
	if lookahead <= 0 {
		lookahead = 5 * time.Minute
	}
	deadline := time.Now().Add(lookahead)
	err = s.db.WithContext(ctx).Model(&model.OAuthCredential{}).Count(&total).Error
	if err != nil {
		return
	}
	err = s.db.WithContext(ctx).Model(&model.OAuthCredential{}).Where("status = ?", "active").Count(&active).Error
	if err != nil {
		return
	}
	err = s.db.WithContext(ctx).Model(&model.OAuthCredential{}).
		Where("status = ?", "active").
		Where("refresh_token <> ''").
		Where("(next_refresh_at IS NULL OR next_refresh_at <= ? OR access_token_expires_at IS NULL OR access_token_expires_at <= ?)", deadline, deadline).
		Count(&due).Error
	return
}

func (s *oauthCredentialService) refreshCredential(ctx context.Context, credential *model.OAuthCredential) error {
	if credential.RefreshTokenExpiresAt != nil && time.Now().After(*credential.RefreshTokenExpiresAt) {
		credential.Status = "expired"
		credential.RefreshError = "refresh token expired"
		_ = s.db.WithContext(ctx).Save(credential).Error
		return fmt.Errorf("refresh token expired")
	}
	result, err := s.refreshWithAdapter(ctx, credential)
	now := time.Now()
	credential.LastRefreshedAt = &now
	if err != nil {
		credential.RefreshError = err.Error()
		if saveErr := s.db.WithContext(ctx).Save(credential).Error; saveErr != nil {
			return saveErr
		}
		return err
	}
	credential.AccessToken = result.AccessToken
	if result.RefreshToken != "" {
		credential.RefreshToken = result.RefreshToken
	}
	credential.TokenType = firstNonEmpty(result.TokenType, credential.TokenType, "Bearer")
	if result.ScopesJSON != "" {
		credential.ScopesJSON = result.ScopesJSON
	}
	credential.AccessTokenExpiresAt = result.AccessTokenExpiresAt
	credential.NextRefreshAt = nextOAuthRefreshTime(result.AccessTokenExpiresAt, 5*time.Minute)
	credential.Status = "active"
	credential.RefreshError = ""
	return s.db.WithContext(ctx).Save(credential).Error
}

type oauthRefreshResult struct {
	AccessToken          string
	RefreshToken         string
	TokenType            string
	ScopesJSON           string
	AccessTokenExpiresAt *time.Time
}

func (s *oauthCredentialService) refreshWithAdapter(ctx context.Context, credential *model.OAuthCredential) (*oauthRefreshResult, error) {
	if strings.TrimSpace(credential.TokenEndpoint) != "" {
		return s.refreshViaTokenEndpoint(ctx, credential)
	}
	switch strings.ToLower(strings.TrimSpace(credential.Provider)) {
	case "", "xingyunpan", "local", "internal":
		token, err := randomOAuthCredentialToken(32)
		if err != nil {
			return nil, err
		}
		expiresAt := time.Now().Add(time.Hour)
		return &oauthRefreshResult{
			AccessToken:          "xyp_at_" + token,
			RefreshToken:         credential.RefreshToken,
			TokenType:            "Bearer",
			ScopesJSON:           credential.ScopesJSON,
			AccessTokenExpiresAt: &expiresAt,
		}, nil
	default:
		return nil, fmt.Errorf("provider %q requires token_endpoint", credential.Provider)
	}
}

func (s *oauthCredentialService) refreshViaTokenEndpoint(ctx context.Context, credential *model.OAuthCredential) (*oauthRefreshResult, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", credential.RefreshToken)
	if credential.App.ClientID != "" {
		form.Set("client_id", credential.App.ClientID)
	}
	if credential.App.ClientSecret != "" {
		form.Set("client_secret", credential.App.ClientSecret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, credential.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		Description  string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("provider refresh failed: %s %s", body.Error, body.Description)
	}
	if strings.TrimSpace(body.AccessToken) == "" {
		return nil, fmt.Errorf("provider response missing access_token")
	}
	var expiresAt *time.Time
	if body.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(body.ExpiresIn) * time.Second)
		expiresAt = &t
	}
	scopesJSON := credential.ScopesJSON
	if strings.TrimSpace(body.Scope) != "" {
		scopesJSON = mustJSON(strings.Fields(body.Scope))
	}
	return &oauthRefreshResult{
		AccessToken:          body.AccessToken,
		RefreshToken:         body.RefreshToken,
		TokenType:            firstNonEmpty(body.TokenType, "Bearer"),
		ScopesJSON:           scopesJSON,
		AccessTokenExpiresAt: expiresAt,
	}, nil
}

func (s *oauthCredentialService) ensureTable(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&model.OAuthCredential{})
}

func toOAuthCredentialPayload(row *model.OAuthCredential) OAuthCredentialPayload {
	return OAuthCredentialPayload{
		ID:                    row.ID,
		AppID:                 row.AppID,
		UserID:                row.UserID,
		Provider:              row.Provider,
		Subject:               row.Subject,
		TokenType:             row.TokenType,
		Scopes:                parseJSONStringSlice(row.ScopesJSON),
		TokenEndpoint:         row.TokenEndpoint,
		AccessTokenPreview:    tokenPreview(row.AccessToken),
		RefreshTokenPreview:   tokenPreview(row.RefreshToken),
		AccessTokenExpiresAt:  row.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: row.RefreshTokenExpiresAt,
		LastRefreshedAt:       row.LastRefreshedAt,
		NextRefreshAt:         row.NextRefreshAt,
		Status:                row.Status,
		RefreshError:          row.RefreshError,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}

func nextOAuthRefreshTime(expiresAt *time.Time, skew time.Duration) *time.Time {
	if expiresAt == nil {
		return nil
	}
	next := expiresAt.Add(-skew)
	if next.Before(time.Now()) {
		now := time.Now()
		return &now
	}
	return &next
}

func tokenPreview(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func normalizeJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !json.Valid([]byte(raw)) {
		return "{}"
	}
	return raw
}

func randomOAuthCredentialToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func uintPtrFromQuery(value string) *uint {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil
	}
	result := uint(parsed)
	return &result
}
