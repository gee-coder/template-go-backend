package service

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
)

// UpdateAuthSettingInput describes auth setting changes from admin.
type UpdateAuthSettingInput struct {
	EnableEmailLogin        bool
	EnablePhoneLogin        bool
	EnableEmailRegistration bool
	EnablePhoneRegistration bool
}

// AuthSettingService provides admin-facing auth setting operations.
type AuthSettingService interface {
	Get(ctx context.Context) (AuthOptions, error)
	Update(ctx context.Context, input UpdateAuthSettingInput) (AuthOptions, error)
}

type authSettingService struct {
	defaults config.AuthConfig
	repo     repository.AuthSettingRepository
	cache    repository.CacheStore
}

// NewAuthSettingService creates an auth setting service.
func NewAuthSettingService(defaults config.AuthConfig, repo repository.AuthSettingRepository, cache repository.CacheStore) AuthSettingService {
	return &authSettingService{
		defaults: defaults,
		repo:     repo,
		cache:    cache,
	}
}

func (s *authSettingService) Get(ctx context.Context) (AuthOptions, error) {
	return loadAuthOptions(ctx, s.defaults, s.repo, s.cache)
}

func (s *authSettingService) Update(ctx context.Context, input UpdateAuthSettingInput) (AuthOptions, error) {
	setting, err := s.repo.Get(ctx)
	if err != nil {
		if err != utils.ErrNotFound {
			return AuthOptions{}, err
		}
		setting = &model.AuthSetting{}
	}

	setting.EnableEmailLogin = input.EnableEmailLogin
	setting.EnablePhoneLogin = input.EnablePhoneLogin
	setting.EnableEmailRegistration = input.EnableEmailRegistration
	setting.EnablePhoneRegistration = input.EnablePhoneRegistration

	options := normalizeAuthOptions(AuthOptions{
		EnableUsernameLogin:     true,
		EnableEmailLogin:        setting.EnableEmailLogin,
		EnablePhoneLogin:        setting.EnablePhoneLogin,
		EnableEmailRegistration: setting.EnableEmailRegistration,
		EnablePhoneRegistration: setting.EnablePhoneRegistration,
	})

	setting.EnableEmailLogin = options.EnableEmailLogin
	setting.EnablePhoneLogin = options.EnablePhoneLogin
	setting.EnableEmailRegistration = options.EnableEmailRegistration
	setting.EnablePhoneRegistration = options.EnablePhoneRegistration

	if err := s.repo.Save(ctx, setting); err != nil {
		return AuthOptions{}, err
	}

	if s.cache != nil {
		_ = s.cache.SetJSON(ctx, authOptionsCacheKey, options, authOptionsCacheTTL)
	}

	return options, nil
}
