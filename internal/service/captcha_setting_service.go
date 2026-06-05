package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// CaptchaSettingPayload is the API-facing shape for captcha settings.
type CaptchaSettingPayload struct {
	LoginEnabled         bool     `json:"login_enabled"`
	RegisterEnabled      bool     `json:"register_enabled"`
	ResetPasswordEnabled bool     `json:"reset_password_enabled"`
	Provider             string   `json:"provider"`
	SecurityLevel        string   `json:"security_level"`
	SiteKey              string   `json:"site_key"`
	SecretKey            string   `json:"secret_key"`
	FailureThreshold     int      `json:"failure_threshold"`
	CooldownSeconds      int      `json:"cooldown_seconds"`
	WhitelistPaths       []string `json:"whitelist_paths"`
}

// CaptchaSettingService provides admin access to captcha settings.
type CaptchaSettingService interface {
	Get() (*CaptchaSettingPayload, error)
	Update(payload *CaptchaSettingPayload) (*CaptchaSettingPayload, error)
}

type captchaSettingService struct {
	repo repository.CaptchaSettingRepository
}

// NewCaptchaSettingService creates a captcha settings service.
func NewCaptchaSettingService(repo repository.CaptchaSettingRepository) CaptchaSettingService {
	return &captchaSettingService{repo: repo}
}

// Get returns captcha settings, creating a default payload when no row exists yet.
func (s *captchaSettingService) Get() (*CaptchaSettingPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}

	if setting == nil {
		return defaultCaptchaSettingPayload(), nil
	}

	return toCaptchaSettingPayload(setting)
}

// Update persists the captcha settings.
func (s *captchaSettingService) Update(payload *CaptchaSettingPayload) (*CaptchaSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("验证码设置信息不能为空")
	}

	normalized, err := normalizeCaptchaSettingPayload(payload)
	if err != nil {
		return nil, err
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.CaptchaSetting{}
	}

	whitelistJSON, err := json.Marshal(normalized.WhitelistPaths)
	if err != nil {
		return nil, fmt.Errorf("序列化白名单路径失败: %w", err)
	}

	setting.LoginEnabled = normalized.LoginEnabled
	setting.RegisterEnabled = normalized.RegisterEnabled
	setting.ResetPasswordEnabled = normalized.ResetPasswordEnabled
	setting.Provider = normalized.Provider
	setting.SecurityLevel = normalized.SecurityLevel
	setting.SiteKey = normalized.SiteKey
	setting.SecretKey = normalized.SecretKey
	setting.FailureThreshold = normalized.FailureThreshold
	setting.CooldownSeconds = normalized.CooldownSeconds
	setting.WhitelistPathsJSON = string(whitelistJSON)

	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}

	return toCaptchaSettingPayload(setting)
}

func toCaptchaSettingPayload(setting *model.CaptchaSetting) (*CaptchaSettingPayload, error) {
	whitelistPaths := make([]string, 0)
	if strings.TrimSpace(setting.WhitelistPathsJSON) != "" {
		if err := json.Unmarshal([]byte(setting.WhitelistPathsJSON), &whitelistPaths); err != nil {
			return nil, fmt.Errorf("解析白名单路径失败: %w", err)
		}
	}

	return &CaptchaSettingPayload{
		LoginEnabled:         setting.LoginEnabled,
		RegisterEnabled:      setting.RegisterEnabled,
		ResetPasswordEnabled: setting.ResetPasswordEnabled,
		Provider:             setting.Provider,
		SecurityLevel:        setting.SecurityLevel,
		SiteKey:              setting.SiteKey,
		SecretKey:            setting.SecretKey,
		FailureThreshold:     setting.FailureThreshold,
		CooldownSeconds:      setting.CooldownSeconds,
		WhitelistPaths:       whitelistPaths,
	}, nil
}

func defaultCaptchaSettingPayload() *CaptchaSettingPayload {
	return &CaptchaSettingPayload{
		LoginEnabled:         true,
		RegisterEnabled:      true,
		ResetPasswordEnabled: true,
		Provider:             "image",
		SecurityLevel:        "balanced",
		SiteKey:              "",
		SecretKey:            "",
		FailureThreshold:     3,
		CooldownSeconds:      60,
		WhitelistPaths:       []string{"/api/v1/health"},
	}
}

func normalizeCaptchaSettingPayload(payload *CaptchaSettingPayload) (*CaptchaSettingPayload, error) {
	normalized := *payload
	normalized.Provider = strings.TrimSpace(strings.ToLower(normalized.Provider))
	normalized.SecurityLevel = strings.TrimSpace(strings.ToLower(normalized.SecurityLevel))
	normalized.SiteKey = strings.TrimSpace(normalized.SiteKey)
	normalized.SecretKey = strings.TrimSpace(normalized.SecretKey)

	if !normalized.LoginEnabled && !normalized.RegisterEnabled && !normalized.ResetPasswordEnabled {
		return nil, fmt.Errorf("至少需要启用一个验证码场景")
	}

	switch normalized.Provider {
	case "image", "slider", "turnstile", "recaptcha":
	default:
		return nil, fmt.Errorf("不支持的验证码类型")
	}

	switch normalized.SecurityLevel {
	case "balanced", "strict", "relaxed":
	default:
		return nil, fmt.Errorf("不支持的难度等级")
	}

	if normalized.Provider == "turnstile" || normalized.Provider == "recaptcha" {
		if normalized.SiteKey == "" {
			return nil, fmt.Errorf("当前验证码类型要求填写站点 Key")
		}
		if normalized.SecretKey == "" {
			return nil, fmt.Errorf("当前验证码类型要求填写服务端 Secret")
		}
	}

	if normalized.FailureThreshold < 1 {
		normalized.FailureThreshold = 1
	}
	if normalized.FailureThreshold > 10 {
		normalized.FailureThreshold = 10
	}
	if normalized.CooldownSeconds < 0 {
		normalized.CooldownSeconds = 0
	}
	if normalized.CooldownSeconds > 3600 {
		normalized.CooldownSeconds = 3600
	}

	cleanPaths := make([]string, 0, len(normalized.WhitelistPaths))
	for _, item := range normalized.WhitelistPaths {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		cleanPaths = append(cleanPaths, item)
	}
	normalized.WhitelistPaths = cleanPaths

	return &normalized, nil
}
