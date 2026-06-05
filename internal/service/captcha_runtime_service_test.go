package service

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"xingyunpan-v2/internal/model"
	"xingyunpan-v2/internal/repository"
)

func TestCaptchaRuntimeImageEnabledWrongAndCorrectAnswer(t *testing.T) {
	runtime := newCaptchaRuntimeTestService(t, model.CaptchaSetting{
		LoginEnabled:     true,
		Provider:         "image",
		SecurityLevel:    "balanced",
		FailureThreshold: 3,
		CooldownSeconds:  60,
	})

	cfg, err := runtime.PublicConfig(CaptchaSceneLogin, "/api/v1/user/login", "alice", "127.0.0.1")
	if err != nil {
		t.Fatalf("public config: %v", err)
	}
	if !cfg.Required || cfg.Provider != "image" {
		t.Fatalf("config = required:%v provider:%q, want image required", cfg.Required, cfg.Provider)
	}

	challenge, err := runtime.CreateChallenge(CaptchaSceneLogin, "/api/v1/user/login", "alice", "127.0.0.1")
	if err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if challenge.CaptchaID == "" || challenge.ImageDataURL == "" {
		t.Fatalf("image challenge missing id or image: %#v", challenge)
	}
	if err := runtime.Verify(CaptchaInput{
		Scene:         CaptchaSceneLogin,
		Path:          "/api/v1/user/login",
		Identity:      "alice",
		RemoteIP:      "127.0.0.1",
		CaptchaID:     challenge.CaptchaID,
		CaptchaAnswer: "wrong",
	}); err == nil {
		t.Fatalf("wrong image captcha passed")
	}

	answer := runtime.(*captchaRuntimeService).challenges[challenge.CaptchaID].Answer
	if err := runtime.Verify(CaptchaInput{
		Scene:         CaptchaSceneLogin,
		Path:          "/api/v1/user/login",
		Identity:      "alice",
		RemoteIP:      "127.0.0.1",
		CaptchaID:     challenge.CaptchaID,
		CaptchaAnswer: answer,
	}); err != nil {
		t.Fatalf("correct image captcha failed: %v", err)
	}
}

func TestCaptchaRuntimeDisabledSceneCanBeForcedByFailureThresholdAndCooldown(t *testing.T) {
	runtime := newCaptchaRuntimeTestService(t, model.CaptchaSetting{
		LoginEnabled:     false,
		Provider:         "image",
		SecurityLevel:    "balanced",
		FailureThreshold: 2,
		CooldownSeconds:  30,
	})

	cfg, err := runtime.PublicConfig(CaptchaSceneLogin, "/api/v1/user/login", "bob", "127.0.0.1")
	if err != nil {
		t.Fatalf("public config: %v", err)
	}
	if cfg.Required {
		t.Fatalf("disabled login captcha required before failures")
	}

	_ = runtime.ReportFailure(CaptchaSceneLogin, "bob", "127.0.0.1")
	cfg, _ = runtime.PublicConfig(CaptchaSceneLogin, "/api/v1/user/login", "bob", "127.0.0.1")
	if cfg.Required {
		t.Fatalf("captcha required after one failure, want threshold of two")
	}

	_ = runtime.ReportFailure(CaptchaSceneLogin, "bob", "127.0.0.1")
	cfg, _ = runtime.PublicConfig(CaptchaSceneLogin, "/api/v1/user/login", "bob", "127.0.0.1")
	if !cfg.Required {
		t.Fatalf("captcha not forced after failure threshold")
	}
	if err := runtime.Verify(CaptchaInput{Scene: CaptchaSceneLogin, Path: "/api/v1/user/login", Identity: "bob", RemoteIP: "127.0.0.1"}); err == nil {
		t.Fatalf("cooldown did not block forced captcha verification")
	}

	_ = runtime.ReportSuccess(CaptchaSceneLogin, "bob", "127.0.0.1")
	cfg, _ = runtime.PublicConfig(CaptchaSceneLogin, "/api/v1/user/login", "bob", "127.0.0.1")
	if cfg.Required {
		t.Fatalf("captcha still required after success reset")
	}
}

func TestCaptchaRuntimeWhitelistSkipsCaptcha(t *testing.T) {
	paths, _ := json.Marshal([]string{"/api/v1/user/login"})
	runtime := newCaptchaRuntimeTestService(t, model.CaptchaSetting{
		LoginEnabled:       true,
		Provider:           "image",
		FailureThreshold:   1,
		WhitelistPathsJSON: string(paths),
	})
	cfg, err := runtime.PublicConfig(CaptchaSceneLogin, "/api/v1/user/login", "carol", "127.0.0.1")
	if err != nil {
		t.Fatalf("public config: %v", err)
	}
	if cfg.Required {
		t.Fatalf("whitelisted path required captcha")
	}
	if err := runtime.Verify(CaptchaInput{Scene: CaptchaSceneLogin, Path: "/api/v1/user/login", Identity: "carol"}); err != nil {
		t.Fatalf("whitelisted verify failed: %v", err)
	}
}

func TestCaptchaRuntimeSliderCorrectWrongAndExpired(t *testing.T) {
	runtime := newCaptchaRuntimeTestService(t, model.CaptchaSetting{
		LoginEnabled:     true,
		Provider:         "slider",
		SecurityLevel:    "strict",
		FailureThreshold: 3,
		CooldownSeconds:  0,
	})

	challenge, err := runtime.CreateChallenge(CaptchaSceneLogin, "/api/v1/user/login", "dave", "127.0.0.1")
	if err != nil {
		t.Fatalf("create slider challenge: %v", err)
	}
	target := runtime.(*captchaRuntimeService).challenges[challenge.CaptchaID].TargetX
	if err := runtime.Verify(CaptchaInput{
		Scene:         CaptchaSceneLogin,
		Path:          "/api/v1/user/login",
		Identity:      "dave",
		CaptchaID:     challenge.CaptchaID,
		CaptchaAnswer: "1",
		SliderTrack:   []int{0, 1, 1},
	}); err == nil {
		t.Fatalf("wrong slider captcha passed")
	}
	if err := runtime.Verify(CaptchaInput{
		Scene:         CaptchaSceneLogin,
		Path:          "/api/v1/user/login",
		Identity:      "dave",
		CaptchaID:     challenge.CaptchaID,
		CaptchaAnswer: stringInt(target),
		SliderTrack:   []int{0, target / 2, target},
	}); err != nil {
		t.Fatalf("correct slider captcha failed: %v", err)
	}

	expired, err := runtime.CreateChallenge(CaptchaSceneLogin, "/api/v1/user/login", "erin", "127.0.0.1")
	if err != nil {
		t.Fatalf("create expired slider challenge: %v", err)
	}
	concrete := runtime.(*captchaRuntimeService)
	state := concrete.challenges[expired.CaptchaID]
	state.ExpireAt = time.Now().Add(-time.Second)
	concrete.challenges[expired.CaptchaID] = state
	if err := runtime.Verify(CaptchaInput{
		Scene:         CaptchaSceneLogin,
		Path:          "/api/v1/user/login",
		Identity:      "erin",
		CaptchaID:     expired.CaptchaID,
		CaptchaAnswer: stringInt(state.TargetX),
	}); err == nil {
		t.Fatalf("expired slider captcha passed")
	}
}

func TestCaptchaRuntimeThirdPartyVerifierMock(t *testing.T) {
	verifier := &mockCaptchaVerifier{validToken: "ok-token"}
	runtime := newCaptchaRuntimeTestServiceWithVerifier(t, model.CaptchaSetting{
		LoginEnabled:     true,
		Provider:         "turnstile",
		SiteKey:          "site-key",
		SecretKey:        "secret-key",
		FailureThreshold: 3,
	}, verifier)

	cfg, err := runtime.PublicConfig(CaptchaSceneLogin, "/api/v1/user/login", "frank", "127.0.0.1")
	if err != nil {
		t.Fatalf("public config: %v", err)
	}
	if cfg.SiteKey != "site-key" {
		t.Fatalf("site key = %q, want public site key", cfg.SiteKey)
	}
	if err := runtime.Verify(CaptchaInput{
		Scene:        CaptchaSceneLogin,
		Path:         "/api/v1/user/login",
		Identity:     "frank",
		RemoteIP:     "127.0.0.1",
		CaptchaToken: "bad-token",
	}); err == nil {
		t.Fatalf("bad third-party token passed")
	}
	if err := runtime.Verify(CaptchaInput{
		Scene:        CaptchaSceneLogin,
		Path:         "/api/v1/user/login",
		Identity:     "frank",
		RemoteIP:     "127.0.0.1",
		CaptchaToken: "ok-token",
	}); err != nil {
		t.Fatalf("valid third-party token failed: %v", err)
	}
	if verifier.provider != "turnstile" || verifier.secret != "secret-key" {
		t.Fatalf("verifier called with provider:%q secret:%q", verifier.provider, verifier.secret)
	}
}

func newCaptchaRuntimeTestService(t *testing.T, setting model.CaptchaSetting) CaptchaRuntimeService {
	t.Helper()
	return newCaptchaRuntimeTestServiceWithVerifier(t, setting, nil)
}

func newCaptchaRuntimeTestServiceWithVerifier(t *testing.T, setting model.CaptchaSetting, verifier CaptchaVerifier) CaptchaRuntimeService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "captcha.db")), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.CaptchaSetting{}); err != nil {
		t.Fatalf("migrate captcha setting: %v", err)
	}
	if setting.Provider == "" {
		setting.Provider = "image"
	}
	if setting.SecurityLevel == "" {
		setting.SecurityLevel = "balanced"
	}
	loginEnabled := setting.LoginEnabled
	registerEnabled := setting.RegisterEnabled
	resetPasswordEnabled := setting.ResetPasswordEnabled
	if err := db.Select("*").Create(&setting).Error; err != nil {
		t.Fatalf("seed captcha setting: %v", err)
	}
	if err := db.Exec(
		"UPDATE captcha_settings SET login_enabled = ?, register_enabled = ?, reset_password_enabled = ? WHERE id = ?",
		loginEnabled,
		registerEnabled,
		resetPasswordEnabled,
		setting.ID,
	).Error; err != nil {
		t.Fatalf("force captcha scene booleans: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return NewCaptchaRuntimeService(repository.NewCaptchaSettingRepository(db), verifier)
}

type mockCaptchaVerifier struct {
	validToken string
	provider   string
	secret     string
}

func (v *mockCaptchaVerifier) Verify(_ context.Context, provider string, secret string, token string, _ string) (bool, error) {
	v.provider = provider
	v.secret = secret
	return token == v.validToken, nil
}

func stringInt(value int) string {
	return strconv.Itoa(value)
}
