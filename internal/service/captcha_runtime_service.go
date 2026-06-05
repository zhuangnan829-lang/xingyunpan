package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

type CaptchaScene string

const (
	CaptchaSceneLogin         CaptchaScene = "login"
	CaptchaSceneRegister      CaptchaScene = "register"
	CaptchaSceneResetPassword CaptchaScene = "reset_password"
)

type CaptchaInput struct {
	Scene         CaptchaScene `json:"scene"`
	Path          string       `json:"path"`
	Identity      string       `json:"identity"`
	RemoteIP      string       `json:"remote_ip"`
	CaptchaToken  string       `json:"captcha_token"`
	CaptchaID     string       `json:"captcha_id"`
	CaptchaAnswer string       `json:"captcha_answer"`
	SliderTrack   []int        `json:"slider_track"`
}

type CaptchaPublicConfig struct {
	Enabled          bool     `json:"enabled"`
	Required         bool     `json:"required"`
	Provider         string   `json:"provider"`
	SecurityLevel    string   `json:"security_level"`
	SiteKey          string   `json:"site_key"`
	FailureThreshold int      `json:"failure_threshold"`
	CooldownSeconds  int      `json:"cooldown_seconds"`
	WhitelistPaths   []string `json:"whitelist_paths"`
}

type CaptchaChallenge struct {
	CaptchaID    string `json:"captcha_id"`
	Provider     string `json:"provider"`
	ImageDataURL string `json:"image_data_url,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

type CaptchaVerifier interface {
	Verify(ctx context.Context, provider string, secret string, token string, remoteIP string) (bool, error)
}

type CaptchaRuntimeService interface {
	PublicConfig(scene CaptchaScene, path, identity, remoteIP string) (*CaptchaPublicConfig, error)
	CreateChallenge(scene CaptchaScene, path, identity, remoteIP string) (*CaptchaChallenge, error)
	Verify(input CaptchaInput) error
	ReportFailure(scene CaptchaScene, identity, remoteIP string) error
	ReportSuccess(scene CaptchaScene, identity, remoteIP string) error
}

type captchaRuntimeService struct {
	repo     repository.CaptchaSettingRepository
	verifier CaptchaVerifier
	now      func() time.Time

	mu         sync.Mutex
	challenges map[string]captchaChallengeState
	failures   map[string]captchaFailureState
}

type captchaChallengeState struct {
	Provider string
	Scene    CaptchaScene
	Answer   string
	TargetX  int
	ExpireAt time.Time
	Used     bool
}

type captchaFailureState struct {
	Count       int
	CooldownTil time.Time
}

func NewCaptchaRuntimeService(repo repository.CaptchaSettingRepository, verifier CaptchaVerifier) CaptchaRuntimeService {
	if verifier == nil {
		verifier = HTTPTokenCaptchaVerifier{Client: http.DefaultClient}
	}
	return &captchaRuntimeService{
		repo:       repo,
		verifier:   verifier,
		now:        time.Now,
		challenges: make(map[string]captchaChallengeState),
		failures:   make(map[string]captchaFailureState),
	}
}

func (s *captchaRuntimeService) PublicConfig(scene CaptchaScene, requestPath, identity, remoteIP string) (*CaptchaPublicConfig, error) {
	setting, err := s.resolveSetting()
	if err != nil {
		return nil, err
	}
	whitelist := captchaWhitelist(setting)
	enabled := captchaSceneEnabled(setting, scene)
	required := s.isRequired(setting, scene, requestPath, identity, remoteIP)
	return &CaptchaPublicConfig{
		Enabled:          enabled,
		Required:         required,
		Provider:         setting.Provider,
		SecurityLevel:    setting.SecurityLevel,
		SiteKey:          setting.SiteKey,
		FailureThreshold: setting.FailureThreshold,
		CooldownSeconds:  setting.CooldownSeconds,
		WhitelistPaths:   whitelist,
	}, nil
}

func (s *captchaRuntimeService) CreateChallenge(scene CaptchaScene, requestPath, identity, remoteIP string) (*CaptchaChallenge, error) {
	setting, err := s.resolveSetting()
	if err != nil {
		return nil, err
	}
	if !s.isRequired(setting, scene, requestPath, identity, remoteIP) {
		return &CaptchaChallenge{Provider: setting.Provider, ExpiresIn: 0}, nil
	}
	if setting.Provider == "turnstile" || setting.Provider == "recaptcha" {
		return nil, fmt.Errorf("third-party captcha does not use local challenges")
	}

	id, err := captchaRandomHex(16)
	if err != nil {
		return nil, err
	}
	expireAt := s.now().Add(2 * time.Minute)
	challenge := captchaChallengeState{Provider: setting.Provider, Scene: scene, ExpireAt: expireAt}
	response := &CaptchaChallenge{CaptchaID: id, Provider: setting.Provider, ExpiresIn: 120}

	if setting.Provider == "slider" {
		target, err := randomInt(42, 238)
		if err != nil {
			return nil, err
		}
		challenge.TargetX = target
		response.Width = 280
		response.Height = 44
	} else {
		answer, err := randomCaptchaAnswer(5)
		if err != nil {
			return nil, err
		}
		challenge.Answer = answer
		response.ImageDataURL = captchaSVGDataURL(answer)
	}

	s.mu.Lock()
	s.cleanupLocked()
	s.challenges[id] = challenge
	s.mu.Unlock()
	return response, nil
}

func (s *captchaRuntimeService) Verify(input CaptchaInput) error {
	setting, err := s.resolveSetting()
	if err != nil {
		return err
	}
	if !s.isRequired(setting, input.Scene, input.Path, input.Identity, input.RemoteIP) {
		return nil
	}
	if s.cooldownActive(setting, input.Scene, input.Identity, input.RemoteIP) {
		return fmt.Errorf("captcha cooldown is active, please try again later")
	}

	var verifyErr error
	switch setting.Provider {
	case "image":
		verifyErr = s.verifyImage(input)
	case "slider":
		verifyErr = s.verifySlider(input)
	case "turnstile", "recaptcha":
		verifyErr = s.verifyRemote(setting, input)
	default:
		verifyErr = fmt.Errorf("unsupported captcha provider")
	}
	if verifyErr != nil {
		_ = s.ReportFailure(input.Scene, input.Identity, input.RemoteIP)
		return verifyErr
	}
	return nil
}

func (s *captchaRuntimeService) ReportFailure(scene CaptchaScene, identity, remoteIP string) error {
	setting, err := s.resolveSetting()
	if err != nil {
		return err
	}
	key := captchaFailureKey(scene, identity, remoteIP)
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.failures[key]
	if state.CooldownTil.After(now) {
		return nil
	}
	state.Count++
	if setting.FailureThreshold > 0 && state.Count >= setting.FailureThreshold && setting.CooldownSeconds > 0 {
		state.CooldownTil = now.Add(time.Duration(setting.CooldownSeconds) * time.Second)
	}
	s.failures[key] = state
	return nil
}

func (s *captchaRuntimeService) ReportSuccess(scene CaptchaScene, identity, remoteIP string) error {
	s.mu.Lock()
	delete(s.failures, captchaFailureKey(scene, identity, remoteIP))
	s.mu.Unlock()
	return nil
}

func (s *captchaRuntimeService) verifyImage(input CaptchaInput) error {
	if strings.TrimSpace(input.CaptchaID) == "" || strings.TrimSpace(input.CaptchaAnswer) == "" {
		return fmt.Errorf("captcha is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.challenges[input.CaptchaID]
	if !ok || state.Provider != "image" || state.Scene != input.Scene || state.Used {
		return fmt.Errorf("captcha is invalid or expired")
	}
	if !state.ExpireAt.After(s.now()) {
		delete(s.challenges, input.CaptchaID)
		return fmt.Errorf("captcha is invalid or expired")
	}
	if !strings.EqualFold(strings.TrimSpace(input.CaptchaAnswer), state.Answer) {
		return fmt.Errorf("captcha answer is incorrect")
	}
	state.Used = true
	s.challenges[input.CaptchaID] = state
	return nil
}

func (s *captchaRuntimeService) verifySlider(input CaptchaInput) error {
	if strings.TrimSpace(input.CaptchaID) == "" {
		return fmt.Errorf("captcha is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.challenges[input.CaptchaID]
	if !ok || state.Provider != "slider" || state.Scene != input.Scene || state.Used {
		return fmt.Errorf("captcha is invalid or expired")
	}
	if !state.ExpireAt.After(s.now()) {
		delete(s.challenges, input.CaptchaID)
		return fmt.Errorf("captcha is invalid or expired")
	}
	answer, err := parseSliderAnswer(input.CaptchaAnswer)
	if err != nil {
		return err
	}
	if math.Abs(float64(answer-state.TargetX)) > 6 {
		return fmt.Errorf("slider captcha answer is incorrect")
	}
	if len(input.SliderTrack) > 0 {
		if len(input.SliderTrack) < 3 {
			return fmt.Errorf("slider track is too short")
		}
		last := input.SliderTrack[len(input.SliderTrack)-1]
		if math.Abs(float64(last-answer)) > 8 {
			return fmt.Errorf("slider track does not match answer")
		}
	}
	state.Used = true
	s.challenges[input.CaptchaID] = state
	return nil
}

func (s *captchaRuntimeService) verifyRemote(setting *model.CaptchaSetting, input CaptchaInput) error {
	token := strings.TrimSpace(input.CaptchaToken)
	if token == "" {
		return fmt.Errorf("captcha token is required")
	}
	ok, err := s.verifier.Verify(context.Background(), setting.Provider, setting.SecretKey, token, input.RemoteIP)
	if err != nil {
		return fmt.Errorf("captcha verification failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("captcha token is invalid")
	}
	return nil
}

func (s *captchaRuntimeService) isRequired(setting *model.CaptchaSetting, scene CaptchaScene, requestPath, identity, remoteIP string) bool {
	if captchaWhitelisted(setting, requestPath) {
		return false
	}
	if captchaSceneEnabled(setting, scene) {
		return true
	}
	key := captchaFailureKey(scene, identity, remoteIP)
	s.mu.Lock()
	state := s.failures[key]
	s.mu.Unlock()
	return setting.FailureThreshold > 0 && state.Count >= setting.FailureThreshold
}

func (s *captchaRuntimeService) cooldownActive(setting *model.CaptchaSetting, scene CaptchaScene, identity, remoteIP string) bool {
	if setting.CooldownSeconds <= 0 {
		return false
	}
	key := captchaFailureKey(scene, identity, remoteIP)
	s.mu.Lock()
	state := s.failures[key]
	s.mu.Unlock()
	return state.CooldownTil.After(s.now())
}

func (s *captchaRuntimeService) resolveSetting() (*model.CaptchaSetting, error) {
	if s.repo == nil {
		return defaultCaptchaRuntimeSetting(), nil
	}
	setting, err := s.repo.Get()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return defaultCaptchaRuntimeSetting(), nil
	}
	if setting.Provider == "" {
		setting.Provider = "image"
	}
	if setting.SecurityLevel == "" {
		setting.SecurityLevel = "balanced"
	}
	if setting.FailureThreshold <= 0 {
		setting.FailureThreshold = 3
	}
	return setting, nil
}

func (s *captchaRuntimeService) cleanupLocked() {
	now := s.now()
	for id, state := range s.challenges {
		if !state.ExpireAt.After(now) || state.Used {
			delete(s.challenges, id)
		}
	}
}

func defaultCaptchaRuntimeSetting() *model.CaptchaSetting {
	return &model.CaptchaSetting{
		LoginEnabled:         true,
		RegisterEnabled:      true,
		ResetPasswordEnabled: true,
		Provider:             "image",
		SecurityLevel:        "balanced",
		FailureThreshold:     3,
		CooldownSeconds:      60,
	}
}

func captchaSceneEnabled(setting *model.CaptchaSetting, scene CaptchaScene) bool {
	if setting == nil {
		return false
	}
	switch scene {
	case CaptchaSceneLogin:
		return setting.LoginEnabled
	case CaptchaSceneRegister:
		return setting.RegisterEnabled
	case CaptchaSceneResetPassword:
		return setting.ResetPasswordEnabled
	default:
		return false
	}
}

func captchaWhitelisted(setting *model.CaptchaSetting, requestPath string) bool {
	requestPath = strings.TrimSpace(requestPath)
	if requestPath == "" {
		return false
	}
	for _, item := range captchaWhitelist(setting) {
		if item == requestPath || strings.HasPrefix(requestPath, strings.TrimRight(item, "/")+"/") {
			return true
		}
	}
	return false
}

func captchaWhitelist(setting *model.CaptchaSetting) []string {
	paths := make([]string, 0)
	if setting == nil || strings.TrimSpace(setting.WhitelistPathsJSON) == "" {
		return paths
	}
	_ = json.Unmarshal([]byte(setting.WhitelistPathsJSON), &paths)
	return paths
}

func captchaFailureKey(scene CaptchaScene, identity, remoteIP string) string {
	identity = strings.ToLower(strings.TrimSpace(identity))
	if identity == "" {
		identity = strings.TrimSpace(remoteIP)
	}
	return string(scene) + ":" + identity
}

func captchaRandomHex(bytesLen int) (string, error) {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate captcha id failed: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func randomInt(min, max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}

func randomCaptchaAnswer(length int) (string, error) {
	const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	var builder strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		builder.WriteByte(alphabet[n.Int64()])
	}
	return builder.String(), nil
}

func captchaSVGDataURL(answer string) string {
	escaped := strings.ToUpper(answer)
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="148" height="54" viewBox="0 0 148 54"><rect width="148" height="54" rx="12" fill="#f8fbff"/><path d="M8 34 C28 12, 44 42, 64 20 S102 14, 138 34" stroke="rgba(37,99,235,.24)" stroke-width="1.5" fill="none"/><text x="20" y="36" font-size="25" font-family="Arial,sans-serif" font-weight="700" fill="#1f3f75" letter-spacing="5">%s</text></svg>`, escaped)
	return "data:image/svg+xml;charset=UTF-8," + url.QueryEscape(svg)
}

func parseSliderAnswer(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("slider captcha answer is required")
	}
	if strings.HasPrefix(value, "{") {
		var payload struct {
			X int `json:"x"`
		}
		if err := json.Unmarshal([]byte(value), &payload); err != nil {
			return 0, fmt.Errorf("slider captcha answer is invalid")
		}
		return payload.X, nil
	}
	answer, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("slider captcha answer is invalid")
	}
	return answer, nil
}

type HTTPTokenCaptchaVerifier struct {
	Client *http.Client
}

func (v HTTPTokenCaptchaVerifier) Verify(ctx context.Context, provider string, secret string, token string, remoteIP string) (bool, error) {
	endpoint := "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	if provider == "recaptcha" {
		endpoint = "https://www.google.com/recaptcha/api/siteverify"
	}
	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if strings.TrimSpace(remoteIP) != "" {
		form.Set("remoteip", remoteIP)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := v.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return false, err
	}
	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}
	return result.Success, nil
}
