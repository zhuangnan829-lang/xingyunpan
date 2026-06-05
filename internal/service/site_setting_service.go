package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// SiteSettingPayload is the API-facing shape for admin site settings.
type SiteSettingPayload struct {
	SiteName            string   `json:"site_name"`
	Tagline             string   `json:"tagline"`
	Description         string   `json:"description"`
	TermsURL            string   `json:"terms_url"`
	PrivacyURL          string   `json:"privacy_url"`
	PrimaryURL          string   `json:"primary_url"`
	BackupURLs          []string `json:"backup_urls"`
	LogoLight           string   `json:"logo_light"`
	LogoDark            string   `json:"logo_dark"`
	Favicon             string   `json:"favicon"`
	Logo192             string   `json:"logo_192"`
	InjectionCode       string   `json:"injection_code"`
	MobileGuideEnabled  bool     `json:"mobile_guide_enabled"`
	MobileFeedbackURL   string   `json:"mobile_feedback_url"`
	DesktopGuideEnabled bool     `json:"desktop_guide_enabled"`
	DesktopCommunityURL string   `json:"desktop_community_url"`
	AllowRegistration   bool     `json:"allow_registration"`
	EmailActivation     bool     `json:"email_activation"`
	PasskeyLoginEnabled bool     `json:"passkey_login_enabled"`
	DefaultGroup        string   `json:"default_group"`
	AvatarPath          string   `json:"avatar_path"`
	AvatarSizeLimitMB   int      `json:"avatar_size_limit_mb"`
	AvatarDimension     int      `json:"avatar_dimension"`
	GravatarServer      string   `json:"gravatar_server"`
}

// SiteSettingService provides admin access to site settings.
type SiteSettingService interface {
	Get() (*SiteSettingPayload, error)
	Update(payload *SiteSettingPayload) (*SiteSettingPayload, error)
}

type siteSettingService struct {
	repo repository.SiteSettingRepository
}

// NewSiteSettingService creates a site settings service.
func NewSiteSettingService(repo repository.SiteSettingRepository) SiteSettingService {
	return &siteSettingService{repo: repo}
}

// Get returns site settings, creating a default payload when no row exists yet.
func (s *siteSettingService) Get() (*SiteSettingPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}

	if setting == nil {
		return defaultSiteSettingPayload(), nil
	}

	return toSiteSettingPayload(setting)
}

// Update persists site settings.
func (s *siteSettingService) Update(payload *SiteSettingPayload) (*SiteSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("站点设置不能为空")
	}

	normalized := normalizeSiteSettingPayload(payload)
	avatarPath, err := NormalizeAvatarPath(normalized.AvatarPath)
	if err != nil {
		return nil, err
	}
	normalized.AvatarPath = avatarPath
	if normalized.SiteName == "" {
		return nil, fmt.Errorf("站点名称不能为空")
	}
	if normalized.DefaultGroup == "" {
		return nil, fmt.Errorf("默认用户组不能为空")
	}
	if normalized.AvatarPath == "" {
		return nil, fmt.Errorf("头像存储路径不能为空")
	}
	if normalized.GravatarServer == "" {
		return nil, fmt.Errorf("Gravatar 服务器不能为空")
	}
	if normalized.AvatarSizeLimitMB <= 0 {
		return nil, fmt.Errorf("头像文件大小限制必须大于 0")
	}
	if normalized.AvatarDimension <= 0 {
		return nil, fmt.Errorf("图像尺寸必须大于 0")
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.SiteSetting{}
	}

	backupJSON, err := json.Marshal(normalized.BackupURLs)
	if err != nil {
		return nil, fmt.Errorf("序列化备选域名失败: %w", err)
	}

	setting.SiteName = normalized.SiteName
	setting.Tagline = normalized.Tagline
	setting.Description = normalized.Description
	setting.TermsURL = normalized.TermsURL
	setting.PrivacyURL = normalized.PrivacyURL
	setting.PrimaryURL = normalized.PrimaryURL
	setting.BackupURLsJSON = string(backupJSON)
	setting.LogoLight = normalized.LogoLight
	setting.LogoDark = normalized.LogoDark
	setting.Favicon = normalized.Favicon
	setting.Logo192 = normalized.Logo192
	setting.InjectionCode = normalized.InjectionCode
	setting.MobileGuideEnabled = normalized.MobileGuideEnabled
	setting.MobileFeedbackURL = normalized.MobileFeedbackURL
	setting.DesktopGuideEnabled = normalized.DesktopGuideEnabled
	setting.DesktopCommunityURL = normalized.DesktopCommunityURL
	setting.AllowRegistration = normalized.AllowRegistration
	setting.EmailActivation = normalized.EmailActivation
	setting.PasskeyLoginEnabled = normalized.PasskeyLoginEnabled
	setting.DefaultGroup = normalized.DefaultGroup
	setting.AvatarPath = normalized.AvatarPath
	setting.AvatarSizeLimitMB = normalized.AvatarSizeLimitMB
	setting.AvatarDimension = normalized.AvatarDimension
	setting.GravatarServer = normalized.GravatarServer

	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}

	return toSiteSettingPayload(setting)
}

func toSiteSettingPayload(setting *model.SiteSetting) (*SiteSettingPayload, error) {
	backupURLs := make([]string, 0)
	if strings.TrimSpace(setting.BackupURLsJSON) != "" {
		if err := json.Unmarshal([]byte(setting.BackupURLsJSON), &backupURLs); err != nil {
			return nil, fmt.Errorf("解析备选域名失败: %w", err)
		}
	}

	return &SiteSettingPayload{
		SiteName:            setting.SiteName,
		Tagline:             setting.Tagline,
		Description:         setting.Description,
		TermsURL:            setting.TermsURL,
		PrivacyURL:          setting.PrivacyURL,
		PrimaryURL:          setting.PrimaryURL,
		BackupURLs:          backupURLs,
		LogoLight:           setting.LogoLight,
		LogoDark:            setting.LogoDark,
		Favicon:             setting.Favicon,
		Logo192:             setting.Logo192,
		InjectionCode:       setting.InjectionCode,
		MobileGuideEnabled:  setting.MobileGuideEnabled,
		MobileFeedbackURL:   setting.MobileFeedbackURL,
		DesktopGuideEnabled: setting.DesktopGuideEnabled,
		DesktopCommunityURL: setting.DesktopCommunityURL,
		AllowRegistration:   setting.AllowRegistration,
		EmailActivation:     setting.EmailActivation,
		PasskeyLoginEnabled: setting.PasskeyLoginEnabled,
		DefaultGroup:        setting.DefaultGroup,
		AvatarPath:          setting.AvatarPath,
		AvatarSizeLimitMB:   setting.AvatarSizeLimitMB,
		AvatarDimension:     setting.AvatarDimension,
		GravatarServer:      setting.GravatarServer,
	}, nil
}

func defaultSiteSettingPayload() *SiteSettingPayload {
	baseURL := strings.TrimSpace(config.Config.Server.BaseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &SiteSettingPayload{
		SiteName:            "星云盘",
		Tagline:             "新一代私有云盘控制台",
		Description:         "面向多节点云盘场景的现代化后台，强调品牌控制、访问路由与多端引导体验。",
		TermsURL:            baseURL + "/terms",
		PrivacyURL:          baseURL + "/privacy-policy",
		PrimaryURL:          baseURL,
		BackupURLs:          []string{},
		LogoLight:           "/assets/branding/logo-light.svg",
		LogoDark:            "/assets/branding/logo-dark.svg",
		Favicon:             "/favicon.ico",
		Logo192:             "/logo192.png",
		InjectionCode:       "",
		MobileGuideEnabled:  true,
		MobileFeedbackURL:   baseURL + "/support/mobile-feedback",
		DesktopGuideEnabled: true,
		DesktopCommunityURL: baseURL + "/community",
		AllowRegistration:   true,
		EmailActivation:     false,
		PasskeyLoginEnabled: false,
		DefaultGroup:        "User",
		AvatarPath:          "avatar",
		AvatarSizeLimitMB:   4,
		AvatarDimension:     200,
		GravatarServer:      "https://www.gravatar.com/",
	}
}

func normalizeSiteSettingPayload(payload *SiteSettingPayload) *SiteSettingPayload {
	normalized := *payload
	normalized.SiteName = strings.TrimSpace(normalized.SiteName)
	normalized.Tagline = strings.TrimSpace(normalized.Tagline)
	normalized.Description = strings.TrimSpace(normalized.Description)
	normalized.TermsURL = strings.TrimSpace(normalized.TermsURL)
	normalized.PrivacyURL = strings.TrimSpace(normalized.PrivacyURL)
	normalized.PrimaryURL = strings.TrimSpace(normalized.PrimaryURL)
	normalized.LogoLight = strings.TrimSpace(normalized.LogoLight)
	normalized.LogoDark = strings.TrimSpace(normalized.LogoDark)
	normalized.Favicon = strings.TrimSpace(normalized.Favicon)
	normalized.Logo192 = strings.TrimSpace(normalized.Logo192)
	normalized.InjectionCode = strings.TrimSpace(normalized.InjectionCode)
	normalized.MobileFeedbackURL = strings.TrimSpace(normalized.MobileFeedbackURL)
	normalized.DesktopCommunityURL = strings.TrimSpace(normalized.DesktopCommunityURL)
	normalized.DefaultGroup = strings.TrimSpace(normalized.DefaultGroup)
	normalized.AvatarPath = strings.TrimSpace(normalized.AvatarPath)
	normalized.GravatarServer = strings.TrimSpace(normalized.GravatarServer)

	cleanURLs := make([]string, 0, len(normalized.BackupURLs))
	for _, item := range normalized.BackupURLs {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		cleanURLs = append(cleanURLs, item)
	}
	normalized.BackupURLs = cleanURLs

	return &normalized
}
