package service

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/crypto"
	"xingyunpan-v2/pkg/token"
)

type DavAccountService interface {
	List(ctx context.Context, userID uint) ([]DavAccountPayload, error)
	Create(ctx context.Context, userID uint, req DavAccountUpsertRequest) (*DavAccountPayload, string, error)
	Update(ctx context.Context, userID uint, id uint, req DavAccountUpsertRequest) (*DavAccountPayload, error)
	Delete(ctx context.Context, userID uint, id uint) error
	ResetSecret(ctx context.Context, userID uint, id uint) (*DavAccountPayload, string, error)
	ResolveByToken(ctx context.Context, accountToken string, secret string, clientIP string) (*DavAccountPayload, error)
}

type davAccountService struct {
	repo    repository.DavAccountRepository
	baseURL string
}

func NewDavAccountService(repo repository.DavAccountRepository, baseURL string) DavAccountService {
	return &davAccountService{
		repo:    repo,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

type DavAccountUpsertRequest struct {
	Name         string `json:"name"`
	RootPath     string `json:"root_path"`
	Permission   string `json:"permission"`
	ReverseProxy bool   `json:"reverse_proxy"`
	Status       string `json:"status"`
	Description  string `json:"description"`
}

type DavAccountPayload struct {
	ID           uint       `json:"id"`
	AccountToken string     `json:"account_token"`
	Name         string     `json:"name"`
	RootPath     string     `json:"root_path"`
	Permission   string     `json:"permission"`
	ReverseProxy bool       `json:"reverse_proxy"`
	Status       string     `json:"status"`
	Endpoint     string     `json:"endpoint"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	LastUsedIP   string     `json:"last_used_ip"`
	Description  string     `json:"description"`
}

func (s *davAccountService) List(ctx context.Context, userID uint) ([]DavAccountPayload, error) {
	accounts, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	payloads := make([]DavAccountPayload, 0, len(accounts))
	for _, account := range accounts {
		payloads = append(payloads, s.toPayload(&account))
	}
	return payloads, nil
}

func (s *davAccountService) Create(ctx context.Context, userID uint, req DavAccountUpsertRequest) (*DavAccountPayload, string, error) {
	normalized, err := normalizeDavAccountRequest(req, true)
	if err != nil {
		return nil, "", err
	}

	accountToken, err := token.GenerateShareTokenWithLength(18)
	if err != nil {
		return nil, "", err
	}
	accountToken = strings.TrimRight(accountToken, "=")

	secret, hash, err := generateDavSecret()
	if err != nil {
		return nil, "", err
	}

	account := &model.DavAccount{
		UserID:       userID,
		AccountToken: accountToken,
		Name:         normalized.Name,
		RootPath:     normalized.RootPath,
		Permission:   normalized.Permission,
		ReverseProxy: normalized.ReverseProxy,
		Status:       normalized.Status,
		SecretHash:   hash,
		Description:  normalized.Description,
	}
	if err := s.repo.Create(ctx, account); err != nil {
		return nil, "", err
	}
	payload := s.toPayload(account)
	return &payload, secret, nil
}

func (s *davAccountService) Update(ctx context.Context, userID uint, id uint, req DavAccountUpsertRequest) (*DavAccountPayload, error) {
	account, err := s.repo.GetByIDForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeDavAccountRequest(req, false)
	if err != nil {
		return nil, err
	}

	account.Name = normalized.Name
	account.RootPath = normalized.RootPath
	account.Permission = normalized.Permission
	account.ReverseProxy = normalized.ReverseProxy
	account.Status = normalized.Status
	account.Description = normalized.Description
	if err := s.repo.Save(ctx, account); err != nil {
		return nil, err
	}
	payload := s.toPayload(account)
	return &payload, nil
}

func (s *davAccountService) Delete(ctx context.Context, userID uint, id uint) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *davAccountService) ResetSecret(ctx context.Context, userID uint, id uint) (*DavAccountPayload, string, error) {
	account, err := s.repo.GetByIDForUser(ctx, userID, id)
	if err != nil {
		return nil, "", err
	}
	secret, hash, err := generateDavSecret()
	if err != nil {
		return nil, "", err
	}
	account.SecretHash = hash
	if err := s.repo.Save(ctx, account); err != nil {
		return nil, "", err
	}
	payload := s.toPayload(account)
	return &payload, secret, nil
}

func (s *davAccountService) ResolveByToken(ctx context.Context, accountToken string, secret string, clientIP string) (*DavAccountPayload, error) {
	account, err := s.repo.GetByToken(ctx, accountToken)
	if err != nil {
		return nil, err
	}
	if account.Status != model.DavAccountStatusActive {
		return nil, fmt.Errorf("WebDAV 账号已停用")
	}
	if err := crypto.VerifyPassword(account.SecretHash, secret); err != nil {
		return nil, fmt.Errorf("WebDAV 密钥错误")
	}
	now := time.Now()
	account.LastUsedAt = &now
	account.LastUsedIP = strings.TrimSpace(clientIP)
	_ = s.repo.Save(ctx, account)
	payload := s.toPayload(account)
	return &payload, nil
}

func (s *davAccountService) toPayload(account *model.DavAccount) DavAccountPayload {
	return DavAccountPayload{
		ID:           account.ID,
		AccountToken: account.AccountToken,
		Name:         account.Name,
		RootPath:     account.RootPath,
		Permission:   account.Permission,
		ReverseProxy: account.ReverseProxy,
		Status:       account.Status,
		Endpoint:     s.endpointFor(account.AccountToken),
		CreatedAt:    account.CreatedAt,
		UpdatedAt:    account.UpdatedAt,
		LastUsedAt:   account.LastUsedAt,
		LastUsedIP:   account.LastUsedIP,
		Description:  account.Description,
	}
}

func (s *davAccountService) endpointFor(accountToken string) string {
	baseURL := s.baseURL
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}
	return baseURL + "/dav/" + accountToken
}

func normalizeDavAccountRequest(req DavAccountUpsertRequest, create bool) (DavAccountUpsertRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		if create {
			req.Name = "WebDAV 账号"
		} else {
			return req, fmt.Errorf("账号名称不能为空")
		}
	}
	if len([]rune(req.Name)) > 120 {
		return req, fmt.Errorf("账号名称不能超过 120 个字符")
	}

	req.RootPath = normalizeRootPath(req.RootPath)
	if req.Permission == "" {
		req.Permission = model.DavAccountPermissionWrite
	}
	if req.Permission != model.DavAccountPermissionRead && req.Permission != model.DavAccountPermissionWrite {
		return req, fmt.Errorf("权限必须为 read 或 write")
	}
	if req.Status == "" {
		req.Status = model.DavAccountStatusActive
	}
	if req.Status != model.DavAccountStatusActive && req.Status != model.DavAccountStatusDisabled {
		return req, fmt.Errorf("状态必须为 active 或 disabled")
	}
	req.Description = strings.TrimSpace(req.Description)
	return req, nil
}

func normalizeRootPath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "." {
		return "/"
	}
	value = strings.ReplaceAll(value, "\\", "/")
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	cleaned := path.Clean(value)
	if cleaned == "." {
		return "/"
	}
	return cleaned
}

func generateDavSecret() (string, string, error) {
	raw, err := token.GenerateShareTokenWithLength(24)
	if err != nil {
		return "", "", err
	}
	secret := "dav_" + strings.TrimRight(raw, "=")
	hash, err := crypto.HashPassword(secret)
	if err != nil {
		return "", "", err
	}
	return secret, hash, nil
}

func ParseDavAccountID(value string) (uint, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("无效的 WebDAV 账号 ID")
	}
	return uint(parsed), nil
}
