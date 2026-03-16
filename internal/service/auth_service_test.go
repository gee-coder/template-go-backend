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

func TestAuthServiceLogin(t *testing.T) {
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
	}, nil, userRepo, tokenStore, nil, nil, nil)

	payload, err := svc.Login(context.Background(), "admin", "Admin123!", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.AccessToken == "" || payload.RefreshToken == "" {
		t.Fatalf("expected tokens to be generated")
	}
}

func TestAuthServiceLoginInvalidPassword(t *testing.T) {
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
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnableEmailLogin:        true,
		EnablePhoneLogin:        true,
		EnableEmailRegistration: true,
		EnablePhoneRegistration: true,
	}, nil, userRepo, tokenStore, nil, nil, nil)

	_, err := svc.Login(context.Background(), "admin", "bad-password", "", "")
	if !errors.Is(err, utils.ErrInvalidCredential) {
		t.Fatalf("expected invalid credential, got %v", err)
	}
}

func TestAuthServiceLoginByEmail(t *testing.T) {
	password, _ := utils.HashPassword("Admin123!")
	userRepo := &fakeUserRepository{
		users: map[string]*model.User{
			"admin": {
				BaseModel: model.BaseModel{ID: 1},
				Username:  "admin",
				Email:     "admin@example.com",
				Password:  password,
				Status:    "enabled",
			},
		},
	}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnableEmailLogin: true,
	}, nil, userRepo, tokenStore, nil, nil, nil)

	payload, err := svc.Login(context.Background(), "admin@example.com", "Admin123!", "email", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.User.Email != "admin@example.com" {
		t.Fatalf("expected email login to resolve user")
	}
}

func TestAuthServiceLoginByPhoneRequiresSMS(t *testing.T) {
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
	smsVerifier := &fakeSMSVerificationService{}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnablePhoneLogin: true,
	}, nil, userRepo, tokenStore, nil, smsVerifier, nil)

	_, err := svc.Login(context.Background(), "18800003333", "Admin123!", "phone", "")
	if err == nil {
		t.Fatalf("expected sms code to be required")
	}

	_, err = svc.Login(context.Background(), "18800003333", "Admin123!", "phone", "654321")
	if err != nil {
		t.Fatalf("unexpected phone login error: %v", err)
	}
	if len(smsVerifier.verified) != 1 || smsVerifier.verified[0].Purpose != "login" {
		t.Fatalf("expected login sms verification to run, got %+v", smsVerifier.verified)
	}
}

func TestAuthServiceRegisterByPhone(t *testing.T) {
	userRepo := &fakeUserRepository{users: map[string]*model.User{}}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	smsVerifier := &fakeSMSVerificationService{}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnablePhoneLogin:        true,
		EnablePhoneRegistration: true,
	}, nil, userRepo, tokenStore, nil, smsVerifier, nil)

	payload, err := svc.Register(context.Background(), RegisterInput{
		Account:      "18800001111",
		RegisterType: "phone",
		Password:     "Admin123!",
		SMSCode:      "123456",
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
	if payload.AccessToken == "" {
		t.Fatalf("expected access token after register")
	}
	if len(smsVerifier.verified) != 1 || smsVerifier.verified[0].Purpose != "register" {
		t.Fatalf("expected register sms verification to run, got %+v", smsVerifier.verified)
	}
}

func TestAuthServiceRegisterByPhoneRequiresSMS(t *testing.T) {
	userRepo := &fakeUserRepository{users: map[string]*model.User{}}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnablePhoneRegistration: true,
	}, nil, userRepo, tokenStore, nil, &fakeSMSVerificationService{}, nil)

	_, err := svc.Register(context.Background(), RegisterInput{
		Account:      "18800004444",
		RegisterType: "phone",
		Password:     "Admin123!",
	})
	if err == nil {
		t.Fatalf("expected phone register to require sms code")
	}
}

func TestAuthServiceRegisterByEmailDisabled(t *testing.T) {
	userRepo := &fakeUserRepository{users: map[string]*model.User{}}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{
		EnableEmailLogin:        true,
		EnableEmailRegistration: false,
	}, nil, userRepo, tokenStore, nil, nil, nil)

	_, err := svc.Register(context.Background(), RegisterInput{
		Account:      "user@example.com",
		RegisterType: "email",
		Password:     "Admin123!",
	})
	if err == nil {
		t.Fatalf("expected email registration to be blocked")
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
	svc := NewAuthService(newTestJWTConfig(), config.AuthConfig{}, nil, userRepo, tokenStore, nil, nil, nil)

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
