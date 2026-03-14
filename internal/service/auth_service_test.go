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
	svc := NewAuthService(config.JWTConfig{
		Issuer:     "test",
		Secret:     "secret",
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}, config.AuthConfig{
		EnableEmailLogin:        true,
		EnablePhoneLogin:        true,
		EnableEmailRegistration: true,
		EnablePhoneRegistration: true,
	}, userRepo, tokenStore)

	payload, err := svc.Login(context.Background(), "admin", "Admin123!", "")
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
	svc := NewAuthService(config.JWTConfig{
		Issuer:     "test",
		Secret:     "secret",
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}, config.AuthConfig{
		EnableEmailLogin:        true,
		EnablePhoneLogin:        true,
		EnableEmailRegistration: true,
		EnablePhoneRegistration: true,
	}, userRepo, tokenStore)

	_, err := svc.Login(context.Background(), "admin", "bad-password", "")
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
	svc := NewAuthService(config.JWTConfig{
		Issuer:     "test",
		Secret:     "secret",
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}, config.AuthConfig{
		EnableEmailLogin: true,
	}, userRepo, tokenStore)

	payload, err := svc.Login(context.Background(), "admin@example.com", "Admin123!", "email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.User.Email != "admin@example.com" {
		t.Fatalf("expected email login to resolve user")
	}
}

func TestAuthServiceRegisterByPhone(t *testing.T) {
	userRepo := &fakeUserRepository{users: map[string]*model.User{}}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	svc := NewAuthService(config.JWTConfig{
		Issuer:     "test",
		Secret:     "secret",
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}, config.AuthConfig{
		EnablePhoneLogin:        true,
		EnablePhoneRegistration: true,
	}, userRepo, tokenStore)

	payload, err := svc.Register(context.Background(), RegisterInput{
		Account:      "18800001111",
		RegisterType: "phone",
		Password:     "Admin123!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.User.Phone != "18800001111" {
		t.Fatalf("expected phone to be persisted, got %s", payload.User.Phone)
	}
	if payload.AccessToken == "" {
		t.Fatalf("expected access token after register")
	}
}

func TestAuthServiceRegisterByEmailDisabled(t *testing.T) {
	userRepo := &fakeUserRepository{users: map[string]*model.User{}}
	tokenStore := &fakeTokenStore{tokens: map[string]uint{}}
	svc := NewAuthService(config.JWTConfig{
		Issuer:     "test",
		Secret:     "secret",
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
	}, config.AuthConfig{
		EnableEmailLogin:        true,
		EnableEmailRegistration: false,
	}, userRepo, tokenStore)

	_, err := svc.Register(context.Background(), RegisterInput{
		Account:      "user@example.com",
		RegisterType: "email",
		Password:     "Admin123!",
	})
	if err == nil {
		t.Fatalf("expected email registration to be blocked")
	}
}
