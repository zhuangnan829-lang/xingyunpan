package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
	texttemplate "text/template"
	"time"
)

// EmailSender sends verification and test emails.
type EmailSender interface {
	SendRegisterCode(toEmail, code string, ttlSeconds int) error
	SendResetPasswordCode(toEmail, code string, ttlSeconds int) error
	SendTestEmail(toEmail, subject, body string) error
}

type SMTPConfig struct {
	Enabled           bool
	Provider          string
	Host              string
	Port              int
	Username          string
	Password          string
	FromName          string
	FromAddress       string
	ReplyTo           string
	ForceSSL          bool
	ConnectionTimeout int
}

type EmailTemplateSiteBasic struct {
	Name string
}

type EmailTemplateLogo struct {
	Normal string
}

type EmailTemplateCommonContext struct {
	SiteBasic EmailTemplateSiteBasic
	SiteUrl   string
	Logo      EmailTemplateLogo
}

type EmailTemplateRenderContext struct {
	CommonContext EmailTemplateCommonContext
	Url           string
	Code          string
	Email         string
	TTLSeconds    int
	TTLMinutes    int
	TTLHours      int
}

type smtpEmailSender struct {
	config           SMTPConfig
	resolver         func() (SMTPConfig, error)
	templateResolver func(templateKey string) (*ResolvedEmailTemplate, error)
	siteResolver     func() (EmailTemplateCommonContext, error)
}

// NewSMTPEmailSender creates a sender with a static config.
func NewSMTPEmailSender(config SMTPConfig) EmailSender {
	return &smtpEmailSender{config: config}
}

// NewSMTPEmailSenderWithResolver creates a sender that loads the latest config on every send.
func NewSMTPEmailSenderWithResolver(resolver func() (SMTPConfig, error)) EmailSender {
	return &smtpEmailSender{resolver: resolver}
}

// NewSMTPEmailSenderWithDependencies creates a sender with dynamic config and template support.
func NewSMTPEmailSenderWithDependencies(
	resolver func() (SMTPConfig, error),
	templateResolver func(templateKey string) (*ResolvedEmailTemplate, error),
	siteResolver func() (EmailTemplateCommonContext, error),
) EmailSender {
	return &smtpEmailSender{
		resolver:         resolver,
		templateResolver: templateResolver,
		siteResolver:     siteResolver,
	}
}

func (s *smtpEmailSender) SendRegisterCode(toEmail, code string, ttlSeconds int) error {
	return s.sendTemplatedCodeEmail(
		toEmail,
		code,
		ttlSeconds,
		"activation",
		"Xingyunpan account activation",
		"Click the link below to activate your account.",
	)
}

func (s *smtpEmailSender) SendResetPasswordCode(toEmail, code string, ttlSeconds int) error {
	return s.sendTemplatedCodeEmail(
		toEmail,
		code,
		ttlSeconds,
		"password-reset",
		"Xingyunpan password reset",
		"Click the link below to continue resetting your password.",
	)
}

func (s *smtpEmailSender) SendTestEmail(toEmail, subject, body string) error {
	config, err := s.getConfig()
	if err != nil {
		return err
	}
	if strings.TrimSpace(subject) == "" {
		subject = "星云盘测试邮件"
	}
	if strings.TrimSpace(body) == "" {
		body = "这是一封来自星云盘管理后台的测试邮件，用于验证 SMTP 发信配置是否已生效。"
	}

	return s.sendEmail(config, toEmail, subject, body, "text/plain; charset=UTF-8")
}

func (s *smtpEmailSender) sendTemplatedCodeEmail(toEmail, code string, ttlSeconds int, templateKey string, fallbackSubject string, fallbackIntro string) error {
	config, err := s.getConfig()
	if err != nil {
		return err
	}
	if strings.TrimSpace(toEmail) == "" {
		return fmt.Errorf("recipient email is required")
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return fmt.Errorf("recipient email format is invalid")
	}

	commonContext, err := s.getCommonContext()
	if err != nil {
		return err
	}

	renderContext := EmailTemplateRenderContext{
		CommonContext: commonContext,
		Url:           buildVerificationActionURL(commonContext.SiteUrl, templateKey, toEmail, code),
		Code:          code,
		Email:         toEmail,
		TTLSeconds:    ttlSeconds,
		TTLMinutes:    maxInt(1, ttlSeconds/60),
		TTLHours:      maxInt(1, ttlSeconds/3600),
	}

	if s.templateResolver != nil {
		resolvedTemplate, err := s.templateResolver(templateKey)
		if err != nil {
			return err
		}
		if resolvedTemplate != nil && (strings.TrimSpace(resolvedTemplate.Subject) != "" || strings.TrimSpace(resolvedTemplate.Content) != "") {
			subject, htmlBody, renderErr := renderEmailTemplate(resolvedTemplate, renderContext)
			if renderErr != nil {
				return renderErr
			}
			return s.sendEmail(config, toEmail, subject, htmlBody, "text/html; charset=UTF-8")
		}
	}

	fallbackBody := fmt.Sprintf("%s\n\n%s", fallbackIntro, renderContext.Url)
	return s.sendEmail(config, toEmail, fallbackSubject, fallbackBody, "text/plain; charset=UTF-8")
}

func (s *smtpEmailSender) getConfig() (SMTPConfig, error) {
	if s.resolver != nil {
		return resolveSMTPConfigFromSource(s.resolver)
	}
	return resolveSMTPConfig(s.config)
}

func (s *smtpEmailSender) getCommonContext() (EmailTemplateCommonContext, error) {
	if s.siteResolver != nil {
		context, err := s.siteResolver()
		if err != nil {
			return EmailTemplateCommonContext{}, err
		}
		return normalizeCommonContext(context), nil
	}

	return normalizeCommonContext(EmailTemplateCommonContext{
		SiteBasic: EmailTemplateSiteBasic{Name: "星云盘"},
		SiteUrl:   "http://127.0.0.1:4173",
		Logo:      EmailTemplateLogo{Normal: "http://127.0.0.1:4173/logo192.png"},
	}), nil
}

func normalizeCommonContext(context EmailTemplateCommonContext) EmailTemplateCommonContext {
	context.SiteBasic.Name = strings.TrimSpace(context.SiteBasic.Name)
	context.SiteUrl = strings.TrimRight(strings.TrimSpace(context.SiteUrl), "/")
	context.Logo.Normal = strings.TrimSpace(context.Logo.Normal)

	if context.SiteBasic.Name == "" {
		context.SiteBasic.Name = "星云盘"
	}
	if context.SiteUrl == "" {
		context.SiteUrl = "http://127.0.0.1:4173"
	}
	if context.Logo.Normal == "" {
		context.Logo.Normal = strings.TrimRight(context.SiteUrl, "/") + "/logo192.png"
	}

	return context
}

func resolveSMTPConfigFromSource(resolver func() (SMTPConfig, error)) (SMTPConfig, error) {
	config, err := resolver()
	if err != nil {
		return SMTPConfig{}, err
	}
	return resolveSMTPConfig(config)
}

func (s *smtpEmailSender) sendEmail(config SMTPConfig, toEmail, subject, body, contentType string) error {
	if !config.Enabled {
		return fmt.Errorf("email service is disabled")
	}
	if strings.TrimSpace(toEmail) == "" {
		return fmt.Errorf("recipient email is required")
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return fmt.Errorf("recipient email format is invalid")
	}

	message := buildEmailMessage(config.FromName, config.FromAddress, config.ReplyTo, toEmail, subject, body, contentType)
	if err := sendSMTPMessage(config, toEmail, message); err != nil {
		return fmt.Errorf("send email failed: %w", err)
	}
	return nil
}

func renderEmailTemplate(resolvedTemplate *ResolvedEmailTemplate, context EmailTemplateRenderContext) (string, string, error) {
	subjectSource := strings.TrimSpace(resolvedTemplate.Subject)
	if subjectSource == "" {
		subjectSource = resolvedTemplate.Name
	}
	bodySource := strings.TrimSpace(resolvedTemplate.Content)
	if bodySource == "" {
		return "", "", fmt.Errorf("template %q has empty content", resolvedTemplate.TemplateKey)
	}

	subjectTpl, err := texttemplate.New(resolvedTemplate.TemplateKey + "-subject").Parse(subjectSource)
	if err != nil {
		return "", "", fmt.Errorf("parse email subject template failed: %w", err)
	}

	var subjectBuilder bytes.Buffer
	if err := subjectTpl.Execute(&subjectBuilder, context); err != nil {
		return "", "", fmt.Errorf("render email subject template failed: %w", err)
	}

	bodyTpl, err := template.New(resolvedTemplate.TemplateKey + "-body").Parse(bodySource)
	if err != nil {
		return "", "", fmt.Errorf("parse email body template failed: %w", err)
	}

	var bodyBuilder bytes.Buffer
	if err := bodyTpl.Execute(&bodyBuilder, context); err != nil {
		return "", "", fmt.Errorf("render email body template failed: %w", err)
	}

	return strings.TrimSpace(subjectBuilder.String()), bodyBuilder.String(), nil
}

func buildVerificationActionURL(siteURL, templateKey, email, code string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(siteURL), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:4173"
	}

	path := "/register"
	if templateKey == "password-reset" {
		path = "/forgot-password"
	}

	values := url.Values{}
	values.Set("email", email)
	values.Set("email_code", code)
	return baseURL + path + "?" + values.Encode()
}

func resolveSMTPConfig(config SMTPConfig) (SMTPConfig, error) {
	resolved := config
	resolved.Provider = strings.ToLower(strings.TrimSpace(resolved.Provider))
	resolved.Host = strings.TrimSpace(resolved.Host)
	resolved.Username = strings.TrimSpace(resolved.Username)
	resolved.Password = strings.TrimSpace(resolved.Password)
	resolved.FromName = strings.TrimSpace(resolved.FromName)
	resolved.FromAddress = strings.TrimSpace(resolved.FromAddress)
	resolved.ReplyTo = strings.TrimSpace(resolved.ReplyTo)

	if resolved.FromAddress == "" {
		resolved.FromAddress = resolved.Username
	}
	if resolved.FromAddress == "" {
		return resolved, fmt.Errorf("from address is required")
	}
	if _, err := mail.ParseAddress(resolved.FromAddress); err != nil {
		return resolved, fmt.Errorf("from address format is invalid")
	}
	if resolved.ReplyTo != "" {
		if _, err := mail.ParseAddress(resolved.ReplyTo); err != nil {
			return resolved, fmt.Errorf("reply-to address format is invalid")
		}
	}

	if resolved.Provider == "" {
		resolved.Provider = detectSMTPProvider(resolved.FromAddress)
	}
	if preset, ok := smtpProviderPresets[resolved.Provider]; ok {
		if resolved.Host == "" {
			resolved.Host = preset.Host
		}
		if resolved.Port <= 0 {
			resolved.Port = preset.Port
		}
	}
	if resolved.ConnectionTimeout <= 0 {
		resolved.ConnectionTimeout = 30
	}
	if resolved.Host == "" || resolved.Port <= 0 || resolved.Username == "" || resolved.Password == "" {
		return resolved, fmt.Errorf("incomplete SMTP config")
	}

	return resolved, nil
}

func sendSMTPMessage(config SMTPConfig, toEmail, message string) error {
	if config.ForceSSL {
		return sendWithTLS(config, toEmail, message)
	}
	return sendWithOptionalStartTLS(config, toEmail, message)
}

func sendWithOptionalStartTLS(config SMTPConfig, toEmail, message string) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	dialer := &net.Dialer{Timeout: time.Duration(config.ConnectionTimeout) * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: config.Host}
		if err := client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}

	return sendWithClient(client, config, toEmail, message)
}

func sendWithTLS(config SMTPConfig, toEmail, message string) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	dialer := &net.Dialer{Timeout: time.Duration(config.ConnectionTimeout) * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: config.Host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	return sendWithClient(client, config, toEmail, message)
}

func sendWithClient(client *smtp.Client, config SMTPConfig, toEmail, message string) error {
	if ok, _ := client.Extension("AUTH"); ok {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(config.FromAddress); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

type smtpProviderPreset struct {
	Host string
	Port int
}

var smtpProviderPresets = map[string]smtpProviderPreset{
	"qq":      {Host: "smtp.qq.com", Port: 587},
	"163":     {Host: "smtp.163.com", Port: 587},
	"126":     {Host: "smtp.126.com", Port: 587},
	"gmail":   {Host: "smtp.gmail.com", Port: 587},
	"outlook": {Host: "smtp.office365.com", Port: 587},
	"hotmail": {Host: "smtp.office365.com", Port: 587},
}

func detectSMTPProvider(email string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")
	if len(parts) != 2 {
		return ""
	}

	domain := parts[1]
	switch {
	case strings.HasSuffix(domain, "qq.com"):
		return "qq"
	case strings.HasSuffix(domain, "163.com"):
		return "163"
	case strings.HasSuffix(domain, "126.com"):
		return "126"
	case strings.HasSuffix(domain, "gmail.com"):
		return "gmail"
	case strings.HasSuffix(domain, "outlook.com"), strings.HasSuffix(domain, "hotmail.com"), strings.HasSuffix(domain, "live.com"):
		return "outlook"
	default:
		return ""
	}
}

func buildEmailMessage(fromName, fromAddress, replyTo, toAddress, subject, body, contentType string) string {
	displayFrom := fromAddress
	if strings.TrimSpace(fromName) != "" {
		displayFrom = mime.QEncoding.Encode("utf-8", fromName) + " <" + fromAddress + ">"
	}

	headers := []string{
		fmt.Sprintf("From: %s", displayFrom),
		fmt.Sprintf("To: %s", toAddress),
		fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("utf-8", strings.TrimSpace(subject))),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: %s", contentType),
		"Content-Transfer-Encoding: 8bit",
	}
	if strings.TrimSpace(replyTo) != "" {
		headers = append(headers, fmt.Sprintf("Reply-To: %s", replyTo))
	}
	headers = append(headers, "", body)
	return strings.Join(headers, "\r\n")
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
