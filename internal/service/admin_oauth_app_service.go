package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"

	"gorm.io/gorm"
)

type OAuthPermissionPayload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
	Enabled     bool   `json:"enabled"`
	Required    bool   `json:"required,omitempty"`
	Mode        string `json:"mode,omitempty"`
	ModeSwitch  bool   `json:"modeSwitch,omitempty"`
	Icon        string `json:"icon"`
}

type AdminOAuthAppPayload struct {
	ID                     uint                     `json:"id"`
	Slug                   string                   `json:"slug"`
	Name                   string                   `json:"name"`
	Description            string                   `json:"description"`
	AppName                string                   `json:"appName"`
	IconPath               string                   `json:"iconPath"`
	ClientID               string                   `json:"clientId"`
	ClientSecret           string                   `json:"clientSecret,omitempty"`
	RedirectURIs           []string                 `json:"redirectUris"`
	Scopes                 []string                 `json:"scopes"`
	IsSystem               bool                     `json:"isSystem"`
	Enabled                bool                     `json:"enabled"`
	CreatedAt              time.Time                `json:"createdAt"`
	UpdatedAt              time.Time                `json:"updatedAt"`
	TokenTTL               string                   `json:"tokenTtl"`
	RefreshTokenTTLSeconds int                      `json:"refreshTokenTtlSeconds"`
	Permissions            []OAuthPermissionPayload `json:"permissions"`
}

type AdminOAuthAppListQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
}

type AdminOAuthAppService interface {
	List(ctx context.Context, query *AdminOAuthAppListQuery) ([]AdminOAuthAppPayload, int64, error)
	Get(ctx context.Context, id string) (*AdminOAuthAppPayload, error)
	Create(ctx context.Context, payload *AdminOAuthAppPayload) (*AdminOAuthAppPayload, error)
	Update(ctx context.Context, id string, payload *AdminOAuthAppPayload) (*AdminOAuthAppPayload, error)
	UpdateStatus(ctx context.Context, id string, enabled bool) (*AdminOAuthAppPayload, error)
	Delete(ctx context.Context, id string) error
	RegenerateSecret(ctx context.Context, id string) (*AdminOAuthAppPayload, error)
}

type adminOAuthAppService struct {
	db *gorm.DB
}

func NewAdminOAuthAppService(db *gorm.DB) AdminOAuthAppService {
	return &adminOAuthAppService{db: db}
}

func (s *adminOAuthAppService) List(ctx context.Context, query *AdminOAuthAppListQuery) ([]AdminOAuthAppPayload, int64, error) {
	if err := s.ensureDefaults(ctx); err != nil {
		return nil, 0, err
	}
	if query == nil {
		query = &AdminOAuthAppListQuery{}
	}
	page := normalizeOAuthPage(query.Page)
	pageSize := normalizeOAuthPageSize(query.PageSize)

	base := s.db.WithContext(ctx).Model(&model.OAuthApp{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("name LIKE ? OR app_name LIKE ? OR client_id LIKE ? OR scopes_json LIKE ?", like, like, like, like)
	}
	switch strings.ToLower(strings.TrimSpace(query.Status)) {
	case "enabled":
		base = base.Where("enabled = ?", true)
	case "disabled":
		base = base.Where("enabled = ?", false)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var apps []model.OAuthApp
	if err := base.Order("is_system DESC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	items := make([]AdminOAuthAppPayload, 0, len(apps))
	for _, app := range apps {
		items = append(items, toOAuthPayload(&app))
	}
	return items, total, nil
}

func (s *adminOAuthAppService) Get(ctx context.Context, id string) (*AdminOAuthAppPayload, error) {
	if err := s.ensureDefaults(ctx); err != nil {
		return nil, err
	}
	app, err := s.find(ctx, id)
	if err != nil {
		return nil, err
	}
	payload := toOAuthPayload(app)
	return &payload, nil
}

func (s *adminOAuthAppService) Create(ctx context.Context, payload *AdminOAuthAppPayload) (*AdminOAuthAppPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("OAuth app cannot be empty")
	}
	normalized, err := s.normalizePayload(payload, false)
	if err != nil {
		return nil, err
	}
	if normalized.Slug == "" {
		normalized.Slug = "custom-" + randomHex(6)
	}
	normalized.ClientID = firstNonEmpty(normalized.ClientID, "xyp_custom_"+randomHex(8))
	plainSecret := firstNonEmpty(normalized.ClientSecret, randomHex(24))
	secretHash, err := hashOAuthSecret(plainSecret)
	if err != nil {
		return nil, err
	}
	normalized.ClientSecret = secretHash
	app, err := payloadToOAuthModel(normalized)
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Create(app).Error; err != nil {
		return nil, err
	}
	result := toOAuthPayload(app)
	result.ClientSecret = plainSecret
	return &result, nil
}

func (s *adminOAuthAppService) Update(ctx context.Context, id string, payload *AdminOAuthAppPayload) (*AdminOAuthAppPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("OAuth app cannot be empty")
	}
	app, err := s.find(ctx, id)
	if err != nil {
		return nil, err
	}
	normalized, err := s.normalizePayload(payload, app.IsSystem)
	if err != nil {
		return nil, err
	}
	app.Name = normalized.Name
	app.Description = normalized.Description
	app.AppName = normalized.AppName
	app.IconPath = normalized.IconPath
	if !app.IsSystem {
		app.ClientID = firstNonEmpty(normalized.ClientID, app.ClientID)
	}
	if strings.TrimSpace(normalized.ClientSecret) != "" {
		secretHash, err := hashOAuthSecret(normalized.ClientSecret)
		if err != nil {
			return nil, err
		}
		app.ClientSecret = secretHash
	}
	app.RedirectURIsJSON = mustJSON(normalized.RedirectURIs)
	app.PermissionsJSON = mustJSON(normalized.Permissions)
	app.ScopesJSON = mustJSON(scopesFromPermissions(normalized.Permissions, normalized.Scopes))
	app.Enabled = normalized.Enabled
	app.TokenTTL = firstNonEmpty(normalized.TokenTTL, app.TokenTTL)
	app.RefreshTokenTTLSeconds = normalized.RefreshTokenTTLSeconds
	if app.RefreshTokenTTLSeconds < 0 {
		app.RefreshTokenTTLSeconds = 0
	}
	if err := s.db.WithContext(ctx).Save(app).Error; err != nil {
		return nil, err
	}
	result := toOAuthPayload(app)
	return &result, nil
}

func (s *adminOAuthAppService) UpdateStatus(ctx context.Context, id string, enabled bool) (*AdminOAuthAppPayload, error) {
	app, err := s.find(ctx, id)
	if err != nil {
		return nil, err
	}
	app.Enabled = enabled
	if err := s.db.WithContext(ctx).Save(app).Error; err != nil {
		return nil, err
	}
	result := toOAuthPayload(app)
	return &result, nil
}

func (s *adminOAuthAppService) Delete(ctx context.Context, id string) error {
	app, err := s.find(ctx, id)
	if err != nil {
		return err
	}
	if app.IsSystem {
		return fmt.Errorf("system OAuth app cannot be deleted")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(app).Error; err != nil {
			return err
		}
		return recordOAuthAudit(tx, 0, app.ID, 0, "oauth.app.deleted", map[string]interface{}{"client_id": app.ClientID})
	})
}

func (s *adminOAuthAppService) RegenerateSecret(ctx context.Context, id string) (*AdminOAuthAppPayload, error) {
	app, err := s.find(ctx, id)
	if err != nil {
		return nil, err
	}
	plainSecret := randomHex(24)
	secretHash, err := hashOAuthSecret(plainSecret)
	if err != nil {
		return nil, err
	}
	app.ClientSecret = secretHash
	if err := s.db.WithContext(ctx).Save(app).Error; err != nil {
		return nil, err
	}
	result := toOAuthPayload(app)
	result.ClientSecret = plainSecret
	return &result, nil
}

func (s *adminOAuthAppService) ensureDefaults(ctx context.Context) error {
	if err := s.ensureTable(ctx); err != nil {
		return err
	}

	for _, payload := range defaultOAuthApps() {
		var existing model.OAuthApp
		err := s.db.WithContext(ctx).Where("slug = ? OR client_id = ?", payload.Slug, payload.ClientID).First(&existing).Error
		if err == nil {
			changed := false
			if !existing.IsSystem {
				existing.IsSystem = true
				changed = true
			}
			if strings.TrimSpace(existing.Name) == "" {
				existing.Name = payload.Name
				changed = true
			}
			if strings.TrimSpace(existing.Description) == "" {
				existing.Description = payload.Description
				changed = true
			}
			if strings.TrimSpace(existing.AppName) == "" {
				existing.AppName = payload.AppName
				changed = true
			}
			if strings.TrimSpace(existing.IconPath) == "" {
				existing.IconPath = payload.IconPath
				changed = true
			}
			if strings.TrimSpace(existing.RedirectURIsJSON) == "" || strings.TrimSpace(existing.RedirectURIsJSON) == "[]" {
				existing.RedirectURIsJSON = mustJSON(payload.RedirectURIs)
				changed = true
			}
			if strings.TrimSpace(existing.ScopesJSON) == "" || strings.TrimSpace(existing.ScopesJSON) == "[]" {
				existing.ScopesJSON = mustJSON(payload.Scopes)
				changed = true
			}
			if strings.TrimSpace(existing.PermissionsJSON) == "" || strings.TrimSpace(existing.PermissionsJSON) == "[]" {
				existing.PermissionsJSON = mustJSON(payload.Permissions)
				changed = true
			}
			if strings.TrimSpace(existing.TokenTTL) == "" {
				existing.TokenTTL = payload.TokenTTL
				changed = true
			}
			if strings.TrimSpace(existing.ClientSecret) != "" && !strings.HasPrefix(existing.ClientSecret, "$2") {
				if secretHash, err := hashOAuthSecret(existing.ClientSecret); err == nil {
					existing.ClientSecret = secretHash
					changed = true
				}
			}
			if existing.RefreshTokenTTLSeconds <= 0 {
				existing.RefreshTokenTTLSeconds = payload.RefreshTokenTTLSeconds
				changed = true
			}
			if changed {
				if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
					return err
				}
			}
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		app, err := payloadToOAuthModel(&payload)
		if err != nil {
			return err
		}
		if err := s.db.WithContext(ctx).Create(app).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *adminOAuthAppService) ensureTable(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(
		&model.OAuthApp{},
		&model.OAuthAuthorizationCode{},
		&model.OAuthAccessToken{},
		&model.OAuthRefreshToken{},
		&model.OAuthGrant{},
		&model.OAuthAuditLog{},
	)
}

func (s *adminOAuthAppService) find(ctx context.Context, id string) (*model.OAuthApp, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("OAuth app ID cannot be empty")
	}
	var app model.OAuthApp
	query := s.db.WithContext(ctx)
	if parsed, err := strconv.ParseUint(id, 10, 64); err == nil {
		query = query.Where("id = ? OR slug = ?", parsed, id)
	} else {
		query = query.Where("slug = ?", id)
	}
	if err := query.First(&app).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("OAuth app not found")
		}
		return nil, err
	}
	return &app, nil
}

func (s *adminOAuthAppService) normalizePayload(payload *AdminOAuthAppPayload, isSystem bool) (*AdminOAuthAppPayload, error) {
	normalized := *payload
	normalized.Name = strings.TrimSpace(normalized.Name)
	normalized.Description = strings.TrimSpace(normalized.Description)
	normalized.AppName = strings.TrimSpace(normalized.AppName)
	normalized.IconPath = strings.TrimSpace(normalized.IconPath)
	normalized.ClientID = strings.TrimSpace(normalized.ClientID)
	normalized.ClientSecret = strings.TrimSpace(normalized.ClientSecret)
	normalized.TokenTTL = strings.TrimSpace(normalized.TokenTTL)
	if normalized.Name == "" {
		return nil, fmt.Errorf("app name cannot be empty")
	}
	if normalized.AppName == "" {
		normalized.AppName = normalized.Name
	}
	if normalized.IconPath == "" {
		normalized.IconPath = "/static/img/xingyunpan_oauth.svg"
	}
	normalized.RedirectURIs = normalizeStringList(normalized.RedirectURIs)
	if len(normalized.RedirectURIs) == 0 {
		return nil, fmt.Errorf("闁煎嘲鍟块惃顖炴閳ь剛鎲版担椋庮伇濞戞搩浜崳鍝モ偓瑙勮壘閹?URI")
	}
	if normalized.RefreshTokenTTLSeconds < 0 {
		return nil, fmt.Errorf("闁告帡鏀遍弻濠冪閵堝洤顤傞柡鍫濐槹閺呫儵寮甸悢鍓佺憹闁煎疇妫勯惃顒佺?0")
	}
	if normalized.TokenTTL == "" {
		normalized.TokenTTL = "7 days"
	}
	if !isSystem {
		normalized.IsSystem = false
	}
	normalized.Permissions = normalizePermissions(normalized.Permissions, normalized.Scopes)
	normalized.Scopes = scopesFromPermissions(normalized.Permissions, normalized.Scopes)
	return &normalized, nil
}

func payloadToOAuthModel(payload *AdminOAuthAppPayload) (*model.OAuthApp, error) {
	return &model.OAuthApp{
		Slug:                   strings.TrimSpace(payload.Slug),
		Name:                   strings.TrimSpace(payload.Name),
		Description:            strings.TrimSpace(payload.Description),
		AppName:                strings.TrimSpace(payload.AppName),
		IconPath:               strings.TrimSpace(payload.IconPath),
		ClientID:               strings.TrimSpace(payload.ClientID),
		ClientSecret:           strings.TrimSpace(payload.ClientSecret),
		RedirectURIsJSON:       mustJSON(payload.RedirectURIs),
		ScopesJSON:             mustJSON(payload.Scopes),
		PermissionsJSON:        mustJSON(payload.Permissions),
		IsSystem:               payload.IsSystem,
		Enabled:                payload.Enabled,
		TokenTTL:               strings.TrimSpace(payload.TokenTTL),
		RefreshTokenTTLSeconds: payload.RefreshTokenTTLSeconds,
	}, nil
}

func toOAuthPayload(app *model.OAuthApp) AdminOAuthAppPayload {
	redirectURIs := parseJSONStringSlice(app.RedirectURIsJSON)
	scopes := parseJSONStringSlice(app.ScopesJSON)
	permissions := parseOAuthPermissions(app.PermissionsJSON)
	return AdminOAuthAppPayload{
		ID:                     app.ID,
		Slug:                   app.Slug,
		Name:                   app.Name,
		Description:            app.Description,
		AppName:                app.AppName,
		IconPath:               app.IconPath,
		ClientID:               app.ClientID,
		RedirectURIs:           redirectURIs,
		Scopes:                 scopes,
		IsSystem:               app.IsSystem,
		Enabled:                app.Enabled,
		CreatedAt:              app.CreatedAt,
		UpdatedAt:              app.UpdatedAt,
		TokenTTL:               app.TokenTTL,
		RefreshTokenTTLSeconds: app.RefreshTokenTTLSeconds,
		Permissions:            permissions,
	}
}

func defaultOAuthApps() []AdminOAuthAppPayload {
	return []AdminOAuthAppPayload{
		{
			Slug:                   "ios-ipados",
			Name:                   "iOS/iPadOS Client",
			Description:            "Official mobile login and file sync entry.",
			AppName:                "application:setting.iOSApp",
			IconPath:               "/static/img/xingyunpan_ios.svg",
			ClientID:               "220db97a-44a3-44f7-99b6-d767262b4daa",
			ClientSecret:           "",
			RedirectURIs:           []string{"/callback/ios", "xingyunpan://oauth/callback"},
			Scopes:                 []string{"profile", "email", "openid", "offline_access", "UserInfo.Write", "UserSecurityInfo.Write", "Files.Write", "Shares.Write", "Workflow.Write", "Finance.Write", "DavAccount.Write"},
			IsSystem:               true,
			Enabled:                true,
			TokenTTL:               "30 days",
			RefreshTokenTTLSeconds: 7776000,
			Permissions: defaultOAuthPermissions(
				[]string{"email", "profile", "offline_access", "userinfo", "security", "files", "shares", "workflow", "finance", "dav"},
				[]string{"userinfo", "security", "files", "shares", "workflow", "finance", "dav"},
			),
		},
		{
			Slug:                   "xingyunpan-desktop",
			Name:                   "Xingyunpan Desktop Client",
			Description:            "Desktop OAuth client for mount, sync, and external client access.",
			AppName:                "application:oauth.desktop",
			IconPath:               "/static/img/xingyunpan_desktop.svg",
			ClientID:               "393a1839-f52e-498e-9972-e77cc2241eee",
			ClientSecret:           "",
			RedirectURIs:           []string{"/callback/desktop", "http://127.0.0.1:5212/callback", "xingyunpan://oauth/callback"},
			Scopes:                 []string{"profile", "email", "openid", "offline_access", "UserInfo.Write", "Workflow.Write", "Files.Write", "Shares.Write"},
			IsSystem:               true,
			Enabled:                true,
			TokenTTL:               "14 days",
			RefreshTokenTTLSeconds: 7776000,
			Permissions: defaultOAuthPermissions(
				[]string{"email", "profile", "offline_access", "userinfo", "files", "shares", "workflow"},
				[]string{"userinfo", "files", "shares", "workflow"},
			),
		},
	}
}

func defaultOAuthPermissions(enabledIDs []string, writeIDs []string) []OAuthPermissionPayload {
	enabled := make(map[string]bool)
	for _, id := range enabledIDs {
		enabled[id] = true
	}
	writes := make(map[string]bool)
	for _, id := range writeIDs {
		writes[id] = true
	}
	return []OAuthPermissionPayload{
		{ID: "openid", Title: "openid (OIDC)", Description: "OpenID Connect identity access.", Scope: "openid", Enabled: true, Required: true, Icon: "document"},
		{ID: "email", Title: "email (OIDC)", Description: "Read user email address and verification status.", Scope: "email", Enabled: enabled["email"], Icon: "document"},
		{ID: "profile", Title: "profile (OIDC)", Description: "Read user display name, avatar, and profile basics.", Scope: "profile", Enabled: enabled["profile"], Icon: "document"},
		{ID: "offline_access", Title: "offline access", Description: "Keep access when the user is not actively using the app.", Scope: "offline_access", Enabled: enabled["offline_access"], Icon: "offline"},
		modePermission("userinfo", "profile", "Read and update account profile information.", "UserInfo", "user", enabled, writes),
		modePermission("security", "security", "Access account security information.", "UserSecurityInfo", "lock", enabled, writes),
		modePermission("files", "files", "Access owned and permitted files.", "Files", "folder", enabled, writes),
		modePermission("shares", "shares", "Access file share links.", "Shares", "share", enabled, writes),
		modePermission("workflow", "workflow", "Access background file processing tasks.", "Workflow", "task", enabled, writes),
		modePermission("finance", "Finance.Read", "Access points, redemption codes, and orders.", "Finance", "finance", enabled, writes),
		modePermission("dav", "WebDAV account", "Access WebDAV account information.", "DavAccount", "dav", enabled, writes),
		{ID: "admin", Title: "admin", Description: "Access site settings and all user data.", Scope: "Admin.Read", Enabled: enabled["admin"], Icon: "admin"},
	}
}

func modePermission(id, title, description, prefix, icon string, enabled, writes map[string]bool) OAuthPermissionPayload {
	mode := "read"
	if writes[id] {
		mode = "write"
	}
	return OAuthPermissionPayload{
		ID:          id,
		Title:       title,
		Description: description,
		Scope:       prefix + "." + strings.Title(mode),
		Enabled:     enabled[id],
		Mode:        mode,
		ModeSwitch:  true,
		Icon:        icon,
	}
}

func normalizePermissions(permissions []OAuthPermissionPayload, scopes []string) []OAuthPermissionPayload {
	if len(permissions) == 0 {
		return defaultOAuthPermissions([]string{"email", "profile"}, nil)
	}
	for i := range permissions {
		permissions[i].ID = strings.TrimSpace(permissions[i].ID)
		permissions[i].Title = strings.TrimSpace(permissions[i].Title)
		permissions[i].Description = strings.TrimSpace(permissions[i].Description)
		permissions[i].Scope = strings.TrimSpace(permissions[i].Scope)
		permissions[i].Icon = strings.TrimSpace(permissions[i].Icon)
		if permissions[i].Required {
			permissions[i].Enabled = true
		}
		if permissions[i].ModeSwitch {
			mode := strings.ToLower(strings.TrimSpace(permissions[i].Mode))
			if mode != "write" {
				mode = "read"
			}
			permissions[i].Mode = mode
			if permissions[i].Scope != "" && strings.Contains(permissions[i].Scope, ".") {
				parts := strings.SplitN(permissions[i].Scope, ".", 2)
				if mode == "write" {
					permissions[i].Scope = parts[0] + ".Write"
				} else {
					permissions[i].Scope = parts[0] + ".Read"
				}
			}
		}
	}
	_ = scopes
	return permissions
}

func scopesFromPermissions(permissions []OAuthPermissionPayload, fallback []string) []string {
	scopes := make([]string, 0, len(permissions))
	seen := map[string]struct{}{}
	for _, permission := range permissions {
		if !permission.Enabled && !permission.Required {
			continue
		}
		scope := strings.TrimSpace(permission.Scope)
		if scope == "" {
			continue
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		scopes = append(scopes, scope)
	}
	if len(scopes) == 0 {
		return normalizeStringList(fallback)
	}
	return scopes
}

func parseOAuthPermissions(raw string) []OAuthPermissionPayload {
	var result []OAuthPermissionPayload
	if strings.TrimSpace(raw) == "" {
		return result
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return []OAuthPermissionPayload{}
	}
	return result
}

func parseJSONStringSlice(raw string) []string {
	var result []string
	if strings.TrimSpace(raw) == "" {
		return result
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return []string{}
	}
	return result
}

func mustJSON(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func normalizeStringList(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean == "" {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}
	return result
}

func randomHex(bytes int) string {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buffer)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func normalizeOAuthPage(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}

func normalizeOAuthPageSize(value int) int {
	if value <= 0 {
		return 10
	}
	if value > 100 {
		return 100
	}
	return value
}
