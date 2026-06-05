package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
	"xingyunpan-v2/pkg/cache"
	"xingyunpan-v2/pkg/crypto"
	"xingyunpan-v2/pkg/jwt"
)

type UserService interface {
	SendRegisterEmailCode(email string) error
	SendResetPasswordEmailCode(email string) error
	Register(username, password, email, emailCode string) (*model.User, error)
	ResetPasswordByEmailCode(email, emailCode, newPassword string) error
	Login(username, password string) (*LoginResponse, error)
	GetUserInfo(userID uint) (*model.User, error)
	UpdateProfile(userID uint, username string) (*model.User, error)
	UpdateAvatar(userID uint, avatarURL string) (*model.User, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	GetPreferences(userID uint) (*model.UserPreference, error)
	UpdatePreferences(userID uint, payload UserPreferencePayload) (*model.UserPreference, error)
}

type userService struct {
	userRepo        repository.UserRepository
	userGroupRepo   repository.UserGroupRepository
	siteSettingRepo repository.SiteSettingRepository
	jwtSecret       string
	jwtExpire       int64
	jwtRefresh      int64
	userCapacity    int64
	cache           *cache.CacheService
	redisClient     *redis.Client
	emailSender     EmailSender
	codeTTL         time.Duration
	sendInterval    time.Duration
}

func NewUserService(
	userRepo repository.UserRepository,
	userGroupRepo repository.UserGroupRepository,
	siteSettingRepo repository.SiteSettingRepository,
	jwtSecret string,
	jwtExpire int64,
	jwtRefresh int64,
	userCapacity int64,
	cache *cache.CacheService,
	redisClient *redis.Client,
	emailSender EmailSender,
	codeTTLSeconds int,
	sendIntervalSeconds int,
) UserService {
	codeTTL := 5 * time.Minute
	if codeTTLSeconds > 0 {
		codeTTL = time.Duration(codeTTLSeconds) * time.Second
	}
	sendInterval := 60 * time.Second
	if sendIntervalSeconds > 0 {
		sendInterval = time.Duration(sendIntervalSeconds) * time.Second
	}

	return &userService{
		userRepo:        userRepo,
		userGroupRepo:   userGroupRepo,
		siteSettingRepo: siteSettingRepo,
		jwtSecret:       jwtSecret,
		jwtExpire:       jwtExpire,
		jwtRefresh:      jwtRefresh,
		userCapacity:    userCapacity,
		cache:           cache,
		redisClient:     redisClient,
		emailSender:     emailSender,
		codeTTL:         codeTTL,
		sendInterval:    sendInterval,
	}
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	User         *model.User `json:"user"`
}

type UserPreferencePayload struct {
	Language          string `json:"language"`
	Timezone          string `json:"timezone"`
	Mode              string `json:"mode"`
	Theme             string `json:"theme"`
	KeepVersions      bool   `json:"keep_versions"`
	VersionExtensions string `json:"version_extensions"`
	MaxVersions       int    `json:"max_versions"`
	ViewSync          string `json:"view_sync"`
	ExpandTree        bool   `json:"expand_tree"`
	FolderAction      string `json:"folder_action"`
	HomeVisibility    string `json:"home_visibility"`
}

func registerCodeKey(email string) string {
	return fmt.Sprintf("register:email_code:%s", strings.ToLower(strings.TrimSpace(email)))
}

func registerSendCooldownKey(email string) string {
	return fmt.Sprintf("register:email_code:cooldown:%s", strings.ToLower(strings.TrimSpace(email)))
}

func resetPasswordCodeKey(email string) string {
	return fmt.Sprintf("reset_password:email_code:%s", strings.ToLower(strings.TrimSpace(email)))
}

func resetPasswordSendCooldownKey(email string) string {
	return fmt.Sprintf("reset_password:email_code:cooldown:%s", strings.ToLower(strings.TrimSpace(email)))
}

func generateSixDigitCode() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate verification code failed: %w", err)
	}
	v := int(binary.BigEndian.Uint64(b[:]) % 1000000)
	return fmt.Sprintf("%06d", v), nil
}

func (s *userService) SendRegisterEmailCode(email string) error {
	return s.sendEmailCode(email, true)
}

func (s *userService) SendResetPasswordEmailCode(email string) error {
	return s.sendEmailCode(email, false)
}

func (s *userService) sendEmailCode(email string, register bool) error {
	if s.redisClient == nil {
		return fmt.Errorf("verification code service is unavailable")
	}
	if s.emailSender == nil {
		return fmt.Errorf("email service is unavailable")
	}
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}

	_, lookupErr := s.userRepo.GetByEmail(email)
	if register && lookupErr == nil {
		return fmt.Errorf("email is already registered")
	}
	if !register && lookupErr != nil {
		return fmt.Errorf("email is not registered")
	}

	ctx := context.Background()
	var codeKey, cooldownKey string
	if register {
		codeKey = registerCodeKey(email)
		cooldownKey = registerSendCooldownKey(email)
	} else {
		codeKey = resetPasswordCodeKey(email)
		cooldownKey = resetPasswordSendCooldownKey(email)
	}
	exists, err := s.redisClient.Exists(ctx, cooldownKey).Result()
	if err != nil {
		return fmt.Errorf("verification code service error")
	}
	if exists > 0 {
		return fmt.Errorf("verification code was sent too frequently")
	}

	code, err := generateSixDigitCode()
	if err != nil {
		return err
	}
	if err := s.redisClient.Set(ctx, codeKey, code, s.codeTTL).Err(); err != nil {
		return fmt.Errorf("save verification code failed")
	}
	if err := s.redisClient.Set(ctx, cooldownKey, "1", s.sendInterval).Err(); err != nil {
		return fmt.Errorf("save verification cooldown failed")
	}

	if register {
		err = s.emailSender.SendRegisterCode(email, code, int(s.codeTTL.Seconds()))
	} else {
		err = s.emailSender.SendResetPasswordCode(email, code, int(s.codeTTL.Seconds()))
	}
	if err != nil {
		_ = s.redisClient.Del(ctx, codeKey, cooldownKey).Err()
		return err
	}
	return nil
}

func (s *userService) Register(username, password, email, emailCode string) (*model.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	emailCode = strings.TrimSpace(emailCode)
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if emailCode == "" {
		return nil, fmt.Errorf("email code cannot be empty")
	}
	if s.redisClient == nil {
		return nil, fmt.Errorf("verification code service is unavailable")
	}

	ctx := context.Background()
	codeKey := registerCodeKey(email)
	savedCode, err := s.redisClient.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("verification code expired")
	}
	if err != nil {
		return nil, fmt.Errorf("verify email code failed")
	}
	if emailCode != strings.TrimSpace(savedCode) {
		return nil, fmt.Errorf("verification code is incorrect")
	}
	if _, err := s.userRepo.GetByUsername(username); err == nil {
		return nil, fmt.Errorf("username already exists")
	}
	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, fmt.Errorf("email already registered")
	}

	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}
	defaultGroupID, capacity, err := s.resolveDefaultUserGroupQuota()
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:    username,
		Password:    hashedPassword,
		Email:       email,
		Role:        "user",
		Enabled:     true,
		UserGroupID: defaultGroupID,
		Capacity:    capacity,
		UsedSize:    0,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}
	_ = s.redisClient.Del(ctx, codeKey, registerSendCooldownKey(email)).Err()
	return user, nil
}

func (s *userService) resolveDefaultUserGroupID() (uint, error) {
	groupID, _, err := s.resolveDefaultUserGroupQuota()
	return groupID, err
}

func (s *userService) resolveDefaultUserGroupQuota() (uint, int64, error) {
	if s.userGroupRepo == nil {
		return 0, s.userCapacity, nil
	}

	preferredName := "User"
	if s.siteSettingRepo != nil {
		setting, err := s.siteSettingRepo.Get()
		if err != nil {
			return 0, 0, fmt.Errorf("resolve default user group failed: %w", err)
		}
		if setting != nil && strings.TrimSpace(setting.DefaultGroup) != "" {
			preferredName = strings.TrimSpace(setting.DefaultGroup)
		}
	}

	group, err := s.userGroupRepo.GetByName(preferredName)
	if err != nil {
		return 0, 0, fmt.Errorf("resolve default user group failed: %w", err)
	}
	if group != nil {
		return group.ID, group.MaxCapacity, nil
	}

	items, err := s.userGroupRepo.List()
	if err != nil {
		return 0, 0, fmt.Errorf("resolve default user group failed: %w", err)
	}
	if len(items) == 0 {
		return 0, s.userCapacity, nil
	}
	return items[0].ID, s.userCapacity, nil
}

func (s *userService) ResetPasswordByEmailCode(email, emailCode, newPassword string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	emailCode = strings.TrimSpace(emailCode)
	newPassword = strings.TrimSpace(newPassword)
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	if emailCode == "" {
		return fmt.Errorf("email code cannot be empty")
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if s.redisClient == nil {
		return fmt.Errorf("verification code service is unavailable")
	}

	ctx := context.Background()
	codeKey := resetPasswordCodeKey(email)
	savedCode, err := s.redisClient.Get(ctx, codeKey).Result()
	if err == redis.Nil {
		return fmt.Errorf("verification code expired")
	}
	if err != nil {
		return fmt.Errorf("verify email code failed")
	}
	if emailCode != strings.TrimSpace(savedCode) {
		return fmt.Errorf("verification code is incorrect")
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("email is not registered")
	}
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}
	user.Password = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("reset password failed: %w", err)
	}
	_ = s.redisClient.Del(ctx, codeKey, resetPasswordSendCooldownKey(email)).Err()
	return nil
}

func (s *userService) Login(username, password string) (*LoginResponse, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" {
		return nil, fmt.Errorf("username or email cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	var user *model.User
	var err error
	if strings.Contains(username, "@") {
		user, err = s.userRepo.GetByEmail(strings.ToLower(username))
	} else {
		user, err = s.userRepo.GetByUsername(username)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}
	if err := crypto.VerifyPassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}
	if !user.Enabled {
		return nil, fmt.Errorf("account is disabled")
	}

	tokenPair, err := jwt.GenerateTokenPair(user.ID, user.Username, s.jwtSecret, int(s.jwtExpire), int(s.jwtRefresh))
	if err != nil {
		return nil, fmt.Errorf("generate token failed: %w", err)
	}
	return &LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         s.withRuntimeAvatar(user),
	}, nil
}

func (s *userService) GetUserInfo(userID uint) (*model.User, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	ctx := context.Background()
	if s.cache != nil {
		var cached model.User
		if err := s.cache.GetUserProfile(ctx, userID, &cached); err == nil {
			return s.withRuntimeAvatar(&cached), nil
		}
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("get user info failed: %w", err)
	}
	if s.cache != nil {
		_ = s.cache.CacheUserProfile(ctx, userID, user)
	}
	return s.withRuntimeAvatar(user), nil
}

func (s *userService) UpdateProfile(userID uint, username string) (*model.User, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("nickname cannot be empty")
	}
	if len([]rune(username)) > 50 {
		return nil, fmt.Errorf("nickname cannot exceed 50 characters")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user.Username != username {
		if existing, err := s.userRepo.GetByUsername(username); err == nil && existing != nil && existing.ID != userID {
			return nil, fmt.Errorf("nickname is already in use")
		}
		user.Username = username
		if err := s.userRepo.Update(user); err != nil {
			return nil, err
		}
	}
	if s.cache != nil {
		_ = s.cache.InvalidateUserProfile(context.Background(), userID)
	}
	return s.withRuntimeAvatar(user), nil
}

func (s *userService) UpdateAvatar(userID uint, avatarURL string) (*model.User, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	avatarURL = strings.TrimSpace(avatarURL)
	if avatarURL == "" {
		return nil, fmt.Errorf("avatar URL cannot be empty")
	}
	if len([]rune(avatarURL)) > 1000 {
		return nil, fmt.Errorf("avatar URL is too long")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	user.AvatarURL = avatarURL
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.InvalidateUserProfile(context.Background(), userID)
	}
	return user, nil
}

func (s *userService) withRuntimeAvatar(user *model.User) *model.User {
	if user == nil || strings.TrimSpace(user.AvatarURL) != "" {
		return user
	}

	server := "https://www.gravatar.com/"
	dimension := 200
	if s.siteSettingRepo != nil {
		setting, err := s.siteSettingRepo.Get()
		if err == nil && setting != nil {
			if strings.TrimSpace(setting.GravatarServer) != "" {
				server = strings.TrimSpace(setting.GravatarServer)
			}
			if setting.AvatarDimension > 0 {
				dimension = setting.AvatarDimension
			}
		}
	}
	copyUser := *user
	copyUser.AvatarURL = GravatarURL(server, user.Email, dimension)
	return &copyUser
}

func (s *userService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	if userID == 0 {
		return fmt.Errorf("user ID cannot be empty")
	}
	oldPassword = strings.TrimSpace(oldPassword)
	newPassword = strings.TrimSpace(newPassword)
	if oldPassword == "" {
		return fmt.Errorf("old password cannot be empty")
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("new password must be at least 6 characters")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if err := crypto.VerifyPassword(user.Password, oldPassword); err != nil {
		return fmt.Errorf("old password is incorrect")
	}
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password failed: %w", err)
	}
	user.Password = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.InvalidateUserProfile(context.Background(), userID)
	}
	return nil
}

func (s *userService) GetPreferences(userID uint) (*model.UserPreference, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if preference, err := s.userRepo.GetPreference(userID); err == nil {
		return preference, nil
	}
	preference := defaultUserPreference(userID)
	if err := s.userRepo.SavePreference(preference); err != nil {
		return nil, err
	}
	return preference, nil
}

func (s *userService) UpdatePreferences(userID uint, payload UserPreferencePayload) (*model.UserPreference, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	preference := defaultUserPreference(userID)
	preference.Language = normalizeChoice(payload.Language, "zh-CN", map[string]bool{"zh-CN": true, "en-US": true})
	preference.Timezone = normalizeText(payload.Timezone, "Asia/Shanghai", 64)
	preference.Mode = normalizeChoice(payload.Mode, "light", map[string]bool{"light": true, "system": true, "dark": true})
	preference.Theme = normalizeChoice(payload.Theme, "sky", map[string]bool{"sky": true, "violet": true, "mint": true})
	preference.KeepVersions = payload.KeepVersions
	preference.VersionExtensions = normalizeText(payload.VersionExtensions, "", 512)
	preference.MaxVersions = payload.MaxVersions
	if preference.MaxVersions < 0 {
		return nil, fmt.Errorf("max versions cannot be negative")
	}
	if preference.MaxVersions > 1000 {
		return nil, fmt.Errorf("max versions cannot exceed 1000")
	}
	preference.ViewSync = normalizeChoice(payload.ViewSync, "server", map[string]bool{"server": true, "local": true})
	preference.ExpandTree = payload.ExpandTree
	preference.FolderAction = normalizeChoice(payload.FolderAction, "open", map[string]bool{"open": true, "select": true})
	preference.HomeVisibility = normalizeChoice(payload.HomeVisibility, "passwordless", map[string]bool{
		"passwordless": true,
		"all":          true,
		"hidden":       true,
	})
	if err := s.userRepo.SavePreference(preference); err != nil {
		return nil, err
	}
	return preference, nil
}

func defaultUserPreference(userID uint) *model.UserPreference {
	return &model.UserPreference{
		UserID:            userID,
		Language:          "zh-CN",
		Timezone:          "Asia/Shanghai",
		Mode:              "light",
		Theme:             "sky",
		KeepVersions:      true,
		VersionExtensions: "",
		MaxVersions:       10,
		ViewSync:          "server",
		ExpandTree:        true,
		FolderAction:      "open",
		HomeVisibility:    "passwordless",
	}
}

func normalizeChoice(value string, fallback string, allowed map[string]bool) string {
	value = strings.TrimSpace(value)
	if allowed[value] {
		return value
	}
	return fallback
}

func normalizeText(value string, fallback string, maxRunes int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	runes := []rune(value)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes])
	}
	return value
}
