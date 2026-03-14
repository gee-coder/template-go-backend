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
				Nickname:  "管理员",
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
	}, userRepo, tokenStore)

	payload, err := svc.Login(context.Background(), "admin", "Admin123!")
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
	}, userRepo, tokenStore)

	_, err := svc.Login(context.Background(), "admin", "bad-password")
	if !errors.Is(err, utils.ErrInvalidCredential) {
		t.Fatalf("expected invalid credential, got %v", err)
	}
}

