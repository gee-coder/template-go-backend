package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
)

func TestAuthServiceUsernameLoginSuccess(t *testing.T) {
	password, _ := utils.HashPassword("Admin123!")
	userRepo := &fakeUserRepository{
		users: map[string]*model.User{
			"admin": {
				BaseModel: model.BaseModel{ID: 1},
				Username:  "admin",
				Nickname:  "Admin",
				Password:  password,
				Status:    "enabled",
				Roles:     []model.Role{{Code: "super_admin"}},
			},
		},
		permissions: []string{"system:user:view"},
	}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnableEmailLogin:        true,
		EnablePhoneLogin:        true,
		EnableEmailRegistration: true,
		EnablePhoneRegistration: true,
	}, nil, userRepo, tokenStore, nil, &fakeSMSVerificationService{}, &fakeEmailVerificationService{}, &fakeImageCaptchaService{}, nil)

	payload, err := svc.Login(context.Background(), LoginInput{
		Account:  "admin",
		Password: "Admin123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.AccessToken == "" || payload.RefreshToken == "" {
		t.Fatalf("expected tokens to be generated")
	}
}

func TestAuthServiceUsernameLoginRequiresCaptchaAfterTwoFailures(t *testing.T) {
	password, _ := utils.HashPassword("Admin123!")
	userRepo := &fakeUserRepository{
		users: map[string]*model.User{
			"admin": {
				BaseModel: model.BaseModel{ID: 1},
				Username:  "admin",
				Password:  password,
				Status:    "enabled",
			},
		},
	}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	cache := &fakeCacheStore{}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{}, nil, userRepo, tokenStore, cache, &fakeSMSVerificationService{}, &fakeEmailVerificationService{}, &fakeImageCaptchaService{}, nil)

	for attempt := 0; attempt < 2; attempt++ {
		_, err := svc.Login(context.Background(), LoginInput{
			Account:  "admin",
			Password: "wrong-pass",
		})
		if !errors.Is(err, utils.ErrInvalidCredential) {
			t.Fatalf("expected invalid credential on attempt %d, got %v", attempt+1, err)
		}
	}

	_, err := svc.Login(context.Background(), LoginInput{
		Account:  "admin",
		Password: "Admin123!",
	})
	if err == nil || err.Error() != "image captcha is required" {
		t.Fatalf("expected image captcha requirement, got %v", err)
	}

	payload, err := svc.Login(context.Background(), LoginInput{
		Account:     "admin",
		Password:    "Admin123!",
		CaptchaID:   "captcha_1",
		CaptchaCode: "ABCD5",
	})
	if err != nil {
		t.Fatalf("expected login success after captcha, got %v", err)
	}
	if payload.AccessToken == "" {
		t.Fatalf("expected token payload after captcha success")
	}
}

func TestAuthServiceEmailLoginUsesEmailCodeAndCaptcha(t *testing.T) {
	userRepo := &fakeUserRepository{
		users: map[string]*model.User{
			"admin": {
				BaseModel: model.BaseModel{ID: 1},
				Username:  "admin",
				Email:     "admin@example.com",
				Password:  "$2a$10$placeholder",
				Status:    "enabled",
			},
		},
	}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	emailService := &fakeEmailVerificationService{}
	captchaService := &fakeImageCaptchaService{}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnableEmailLogin: true,
	}, nil, userRepo, tokenStore, nil, &fakeSMSVerificationService{}, emailService, captchaService, nil)

	payload, err := svc.Login(context.Background(), LoginInput{
		Account:          "admin@example.com",
		LoginType:        "email",
		VerificationCode: "654321",
		CaptchaID:        "captcha_2",
		CaptchaCode:      "PQRS8",
	})
	if err != nil {
		t.Fatalf("unexpected email login error: %v", err)
	}
	if payload.User.Email != "admin@example.com" {
		t.Fatalf("expected email login to resolve user")
	}
	if len(emailService.verified) != 1 || emailService.verified[0].Purpose != "login" {
		t.Fatalf("expected email verification to run, got %+v", emailService.verified)
	}
	if len(captchaService.verified) != 1 {
		t.Fatalf("expected captcha verification to run")
	}
}

func TestAuthServicePhoneLoginWithTwoFactorRequiresPassword(t *testing.T) {
	password, _ := utils.HashPassword("Admin123!")
	userRepo := &fakeUserRepository{
		users: map[string]*model.User{
			"phone_user": {
				BaseModel: model.BaseModel{ID: 1},
				Username:  "phone_user",
				Phone:     "18800003333",
				Password:  password,
				Status:    "enabled",
			},
		},
	}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	smsService := &fakeSMSVerificationService{}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnablePhoneLogin: true,
		EnableTwoFactor:  true,
	}, nil, userRepo, tokenStore, nil, smsService, &fakeEmailVerificationService{}, &fakeImageCaptchaService{}, nil)

	_, err := svc.Login(context.Background(), LoginInput{
		Account:          "18800003333",
		LoginType:        "phone",
		VerificationCode: "123456",
		CaptchaID:        "captcha_3",
		CaptchaCode:      "CODE1",
	})
	if err == nil {
		t.Fatalf("expected password to be required as second factor")
	}

	_, err = svc.Login(context.Background(), LoginInput{
		Account:          "18800003333",
		LoginType:        "phone",
		VerificationCode: "123456",
		CaptchaID:        "captcha_3",
		CaptchaCode:      "CODE1",
		Password:         "Admin123!",
	})
	if err != nil {
		t.Fatalf("unexpected phone login error: %v", err)
	}
	if len(smsService.verified) != 2 {
		t.Fatalf("expected sms verification to run on each phone login attempt")
	}
}

func TestAuthServiceSendTwoFactorCodeUsesPreferredTarget(t *testing.T) {
	password, _ := utils.HashPassword("Admin123!")
	userRepo := &fakeUserRepository{
		users: map[string]*model.User{
			"admin": {
				BaseModel: model.BaseModel{ID: 1},
				Username:  "admin",
				Phone:     "18800001111",
				Email:     "admin@example.com",
				Password:  password,
				Status:    "enabled",
			},
		},
	}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	smsService := &fakeSMSVerificationService{}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnableTwoFactor: true,
	}, nil, userRepo, tokenStore, nil, smsService, &fakeEmailVerificationService{}, &fakeImageCaptchaService{}, nil)

	payload, err := svc.SendTwoFactorCode(context.Background(), SendTwoFactorCodeInput{
		Account: "admin",
	})
	if err != nil {
		t.Fatalf("unexpected send 2fa error: %v", err)
	}
	if payload.Channel != "phone" {
		t.Fatalf("expected phone to be preferred second factor target, got %s", payload.Channel)
	}
	if len(smsService.sent) != 1 || smsService.sent[0].Purpose != "two_factor" {
		t.Fatalf("expected sms send to run, got %+v", smsService.sent)
	}
}

func TestAuthServiceRegisterByPhone(t *testing.T) {
	userRepo := &fakeUserRepository{users: map[string]*model.User{}}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	smsService := &fakeSMSVerificationService{}
	captchaService := &fakeImageCaptchaService{}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnablePhoneRegistration: true,
	}, nil, userRepo, tokenStore, nil, smsService, &fakeEmailVerificationService{}, captchaService, nil)

	payload, err := svc.Register(context.Background(), RegisterInput{
		Account:          "18800001111",
		RegisterType:     "phone",
		Password:         "Admin123!",
		VerificationCode: "123456",
		CaptchaID:        "captcha_4",
		CaptchaCode:      "PHONE",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.User.Phone != "18800001111" {
		t.Fatalf("expected phone to be persisted, got %s", payload.User.Phone)
	}
	if !isSupportedAvatarKey(payload.User.Avatar) {
		t.Fatalf("expected default avatar to be assigned, got %s", payload.User.Avatar)
	}
	if len(smsService.verified) != 1 || smsService.verified[0].Purpose != "register" {
		t.Fatalf("expected phone register verification to run, got %+v", smsService.verified)
	}
	if len(captchaService.verified) != 1 {
		t.Fatalf("expected register captcha verification to run")
	}
}

func TestAuthServiceRegisterByEmailRequiresCodeAndCaptcha(t *testing.T) {
	userRepo := &fakeUserRepository{users: map[string]*model.User{}}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	emailService := &fakeEmailVerificationService{}
	captchaService := &fakeImageCaptchaService{}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnableEmailRegistration: true,
	}, nil, userRepo, tokenStore, nil, &fakeSMSVerificationService{}, emailService, captchaService, nil)

	payload, err := svc.Register(context.Background(), RegisterInput{
		Account:          "newuser@example.com",
		RegisterType:     "email",
		Password:         "Admin123!",
		VerificationCode: "654321",
		CaptchaID:        "captcha_5",
		CaptchaCode:      "EMAIL",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.User.Email != "newuser@example.com" {
		t.Fatalf("expected email to be persisted, got %s", payload.User.Email)
	}
	if len(emailService.verified) != 1 || emailService.verified[0].Purpose != "register" {
		t.Fatalf("expected email register verification to run, got %+v", emailService.verified)
	}
	if len(captchaService.verified) != 1 {
		t.Fatalf("expected register captcha verification to run")
	}
}

func TestAuthServiceUpdateProfileAvatar(t *testing.T) {
	password, _ := utils.HashPassword("Admin123!")
	userRepo := &fakeUserRepository{
		users: map[string]*model.User{
			"admin": {
				BaseModel: model.BaseModel{ID: 1},
				Username:  "admin",
				Nickname:  "Admin",
				Password:  password,
				Status:    "enabled",
				Avatar:    "default-01",
			},
		},
		usersByID:   map[uint]*model.User{},
		permissions: []string{"dashboard:view"},
	}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{}, nil, userRepo, tokenStore, nil, &fakeSMSVerificationService{}, &fakeEmailVerificationService{}, &fakeImageCaptchaService{}, nil)

	profile, err := svc.UpdateProfile(context.Background(), 1, UpdateProfileInput{Avatar: "default-05"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Avatar != "default-05" {
		t.Fatalf("expected avatar to be updated, got %s", profile.Avatar)
	}
}

func newTestJWTConfig() config.JWTConfig {
	return config.JWTConfig{
		Issuer:     "test",
		Secret:     "secret",
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}
}
