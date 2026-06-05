package service

import (
	"fmt"
	"strings"

	"xingyunpan-v2/internal/config"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

// EmailSettingPayload is the API-facing shape for admin email settings.
type EmailSettingPayload struct {
	Enabled             bool   `json:"enabled"`
	Provider            string `json:"provider"`
	Host                string `json:"host"`
	Port                int    `json:"port"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	FromName            string `json:"from_name"`
	FromAddress         string `json:"from_address"`
	ReplyTo             string `json:"reply_to"`
	ForceSSL            bool   `json:"force_ssl"`
	ConnectionTimeout   int    `json:"connection_timeout"`
	CodeTTLSeconds      int    `json:"code_ttl_seconds"`
	SendIntervalSeconds int    `json:"send_interval_seconds"`
}

// EmailSettingService provides admin access to email settings.
type EmailSettingService interface {
	Get() (*EmailSettingPayload, error)
	Update(payload *EmailSettingPayload) (*EmailSettingPayload, error)
	SendTestEmail(toEmail string) error
	GetSMTPConfig() (SMTPConfig, error)
}

type emailSettingService struct {
	repo       repository.EmailSettingRepository
	emailSender EmailSender
}

// NewEmailSettingService creates an email settings service.
func NewEmailSettingService(repo repository.EmailSettingRepository, emailSender EmailSender) EmailSettingService {
	return &emailSettingService{repo: repo, emailSender: emailSender}
}

func (s *emailSettingService) Get() (*EmailSettingPayload, error) {
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return defaultEmailSettingPayload(), nil
	}
	return toEmailSettingPayload(setting), nil
}

func (s *emailSettingService) Update(payload *EmailSettingPayload) (*EmailSettingPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("邮件设置不能为空")
	}

	normalized := normalizeEmailSettingPayload(payload)
	if normalized.Username == "" {
		return nil, fmt.Errorf("SMTP 用户名不能为空")
	}
	if normalized.Password == "" {
		return nil, fmt.Errorf("SMTP 密码不能为空")
	}
	if normalized.FromAddress == "" {
		return nil, fmt.Errorf("发件邮箱不能为空")
	}
	if normalized.Port <= 0 {
		return nil, fmt.Errorf("SMTP 端口必须大于 0")
	}
	if normalized.ConnectionTimeout <= 0 {
		return nil, fmt.Errorf("连接超时时间必须大于 0")
	}

	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &model.EmailSetting{}
	}

	setting.Enabled = normalized.Enabled
	setting.Provider = normalized.Provider
	setting.Host = normalized.Host
	setting.Port = normalized.Port
	setting.Username = normalized.Username
	setting.Password = normalized.Password
	setting.FromName = normalized.FromName
	setting.FromAddress = normalized.FromAddress
	setting.ReplyTo = normalized.ReplyTo
	setting.ForceSSL = normalized.ForceSSL
	setting.ConnectionTimeout = normalized.ConnectionTimeout
	setting.CodeTTLSeconds = normalized.CodeTTLSeconds
	setting.SendIntervalSeconds = normalized.SendIntervalSeconds

	if err := s.repo.Save(setting); err != nil {
		return nil, err
	}
	return toEmailSettingPayload(setting), nil
}

func (s *emailSettingService) SendTestEmail(toEmail string) error {
	if s.emailSender == nil {
		return fmt.Errorf("邮件发送器未初始化")
	}
	return s.emailSender.SendTestEmail(toEmail, "星云盘测试邮件", "这是一封来自星云盘管理后台的测试邮件，用于验证 SMTP 发信配置是否生效。")
}

func (s *emailSettingService) GetSMTPConfig() (SMTPConfig, error) {
	payload, err := s.Get()
	if err != nil {
		return SMTPConfig{}, err
	}
	return SMTPConfig{
		Enabled:           payload.Enabled,
		Provider:          payload.Provider,
		Host:              payload.Host,
		Port:              payload.Port,
		Username:          payload.Username,
		Password:          payload.Password,
		FromName:          payload.FromName,
		FromAddress:       payload.FromAddress,
		ReplyTo:           payload.ReplyTo,
		ForceSSL:          payload.ForceSSL,
		ConnectionTimeout: payload.ConnectionTimeout,
	}, nil
}

func toEmailSettingPayload(setting *model.EmailSetting) *EmailSettingPayload {
	return &EmailSettingPayload{
		Enabled:             setting.Enabled,
		Provider:            setting.Provider,
		Host:                setting.Host,
		Port:                setting.Port,
		Username:            setting.Username,
		Password:            setting.Password,
		FromName:            setting.FromName,
		FromAddress:         setting.FromAddress,
		ReplyTo:             setting.ReplyTo,
		ForceSSL:            setting.ForceSSL,
		ConnectionTimeout:   setting.ConnectionTimeout,
		CodeTTLSeconds:      setting.CodeTTLSeconds,
		SendIntervalSeconds: setting.SendIntervalSeconds,
	}
}

func defaultEmailSettingPayload() *EmailSettingPayload {
	return &EmailSettingPayload{
		Enabled:             config.Config.Email.Enabled,
		Provider:            config.Config.Email.Provider,
		Host:                config.Config.Email.Host,
		Port:                config.Config.Email.Port,
		Username:            config.Config.Email.Username,
		Password:            config.Config.Email.Password,
		FromName:            config.Config.Email.FromName,
		FromAddress:         config.Config.Email.FromAddress,
		ReplyTo:             config.Config.Email.FromAddress,
		ForceSSL:            false,
		ConnectionTimeout:   30,
		CodeTTLSeconds:      config.Config.Email.CodeTTLSeconds,
		SendIntervalSeconds: config.Config.Email.SendIntervalSeconds,
	}
}

func normalizeEmailSettingPayload(payload *EmailSettingPayload) *EmailSettingPayload {
	normalized := *payload
	normalized.Provider = strings.ToLower(strings.TrimSpace(normalized.Provider))
	normalized.Host = strings.TrimSpace(normalized.Host)
	normalized.Username = strings.TrimSpace(normalized.Username)
	normalized.Password = strings.TrimSpace(normalized.Password)
	normalized.FromName = strings.TrimSpace(normalized.FromName)
	normalized.FromAddress = strings.TrimSpace(normalized.FromAddress)
	normalized.ReplyTo = strings.TrimSpace(normalized.ReplyTo)
	return &normalized
}
