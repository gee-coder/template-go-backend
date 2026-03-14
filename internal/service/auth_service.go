package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// TokenPayload describes the token payload returned by auth APIs.
type TokenPayload struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	ExpiresIn    int64        `json:"expiresIn"`
	TokenType    string       `json:"tokenType"`
	User         *ProfileUser `json:"user"`
}

// ProfileUser describes the current user profile response.
type ProfileUser struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Nickname    string    `json:"nickname"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Status      string    `json:"status"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AuthService provides auth capabilities.
type AuthService interface {
	Login(ctx context.Context, username string, password string) (*TokenPayload, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPayload, error)
	Logout(ctx context.Context, refreshToken string) error
	Profile(ctx context.Context, userID uint) (*ProfileUser, error)
	ResolvePermissions(ctx context.Context, userID uint) ([]string, error)
}

type authService struct {
	cfg       config.JWTConfig
	userRepo  repository.UserRepository
	tokenRepo repository.TokenStore
}

// NewAuthService creates the auth service.
func NewAuthService(cfg config.JWTConfig, userRepo repository.UserRepository, tokenRepo repository.TokenStore) AuthService {
	return &authService{
		cfg:       cfg,
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

func (s *authService) Login(ctx context.Context, username string, password string) (*TokenPayload, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return nil, utils.ErrInvalidCredential
		}
		return nil, err
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, utils.ErrInvalidCredential
	}
	if user.Status != "enabled" {
		return nil, utils.ErrForbidden
	}

	return s.issueTokens(ctx, user)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*TokenPayload, error) {
	userID, err := s.tokenRepo.Get(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, utils.ErrUnauthorized
		}
		return nil, err
	}

	if err := s.tokenRepo.Delete(ctx, refreshToken); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user)
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	if err := s.tokenRepo.Delete(ctx, refreshToken); err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	return nil
}

func (s *authService) Profile(ctx context.Context, userID uint) (*ProfileUser, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.userRepo.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return buildProfileUser(user, permissions), nil
}

func (s *authService) ResolvePermissions(ctx context.Context, userID uint) ([]string, error) {
	return s.userRepo.GetPermissions(ctx, userID)
}

func (s *authService) issueTokens(ctx context.Context, user *model.User) (*TokenPayload, error) {
	roleCodes := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roleCodes = append(roleCodes, role.Code)
	}

	accessClaims := utils.NewTokenClaims(user.ID, user.Username, roleCodes, s.cfg.Issuer, s.cfg.AccessTTL)
	accessToken, err := utils.BuildToken(s.cfg.Secret, accessClaims)
	if err != nil {
		return nil, fmt.Errorf("build access token: %w", err)
	}

	refreshToken := uuid.NewString()
	if err := s.tokenRepo.Save(ctx, refreshToken, user.ID, s.cfg.RefreshTTL); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	permissions, err := s.userRepo.GetPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &TokenPayload{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessTTL.Seconds()),
		TokenType:    "Bearer",
		User:         buildProfileUser(user, permissions),
	}, nil
}

func buildProfileUser(user *model.User, permissions []string) *ProfileUser {
	roleCodes := make([]string, 0, len(user.Roles))
	for _, role := range user.Roles {
		roleCodes = append(roleCodes, role.Code)
	}

	return &ProfileUser{
		ID:          user.ID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Email:       user.Email,
		Phone:       user.Phone,
		Status:      user.Status,
		Roles:       roleCodes,
		Permissions: permissions,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

