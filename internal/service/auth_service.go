package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	emailPattern           = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	phonePattern           = regexp.MustCompile(`^\+?[0-9]{6,20}$`)
	generatedNameSanitizer = regexp.MustCompile(`[^a-z0-9_]+`)
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
	Avatar      string    `json:"avatar"`
	Status      string    `json:"status"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AuthOptions describes the auth channels exposed to clients.
type AuthOptions struct {
	EnableUsernameLogin     bool `json:"enableUsernameLogin"`
	EnableEmailLogin        bool `json:"enableEmailLogin"`
	EnablePhoneLogin        bool `json:"enablePhoneLogin"`
	EnableEmailRegistration bool `json:"enableEmailRegistration"`
	EnablePhoneRegistration bool `json:"enablePhoneRegistration"`
}

// RegisterInput describes the public register payload.
type RegisterInput struct {
	Account      string
	RegisterType string
	Nickname     string
	Password     string
	SMSCode      string
}

// UpdateProfileInput describes self profile updates.
type UpdateProfileInput struct {
	Avatar string
}

// AuthService provides auth capabilities.
type AuthService interface {
	Login(ctx context.Context, account string, password string, loginType string, smsCode string) (*TokenPayload, error)
	Register(ctx context.Context, input RegisterInput) (*TokenPayload, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPayload, error)
	Logout(ctx context.Context, refreshToken string) error
	Profile(ctx context.Context, userID uint) (*ProfileUser, error)
	UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) (*ProfileUser, error)
	ResolvePermissions(ctx context.Context, userID uint) ([]string, error)
	Options(ctx context.Context) (AuthOptions, error)
}

type authService struct {
	cfg                config.JWTConfig
	defaults           config.AuthConfig
	authSettingRepo    repository.AuthSettingRepository
	userRepo           repository.UserRepository
	tokenRepo          repository.TokenStore
	cache              repository.CacheStore
	avatarURLValidator func(string) bool
	smsVerifier        smsCodeVerifier
}

type smsCodeVerifier interface {
	VerifyCode(ctx context.Context, input VerifySMSCodeInput) error
}

// NewAuthService creates the auth service.
func NewAuthService(
	cfg config.JWTConfig,
	defaults config.AuthConfig,
	authSettingRepo repository.AuthSettingRepository,
	userRepo repository.UserRepository,
	tokenRepo repository.TokenStore,
	cache repository.CacheStore,
	smsVerifier smsCodeVerifier,
	avatarURLValidator func(string) bool,
) AuthService {
	return &authService{
		cfg:                cfg,
		defaults:           defaults,
		authSettingRepo:    authSettingRepo,
		userRepo:           userRepo,
		tokenRepo:          tokenRepo,
		cache:              cache,
		smsVerifier:        smsVerifier,
		avatarURLValidator: avatarURLValidator,
	}
}

func (s *authService) Login(ctx context.Context, account string, password string, loginType string, smsCode string) (*TokenPayload, error) {
	options, err := s.Options(ctx)
	if err != nil {
		return nil, err
	}

	user, resolvedLoginType, err := s.findUserForLogin(ctx, account, loginType, options)
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
	if resolvedLoginType == "phone" {
		if err := s.verifyPhoneCode(ctx, user.Phone, "login", smsCode); err != nil {
			return nil, err
		}
	}

	return s.issueTokens(ctx, user)
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (*TokenPayload, error) {
	options, err := s.Options(ctx)
	if err != nil {
		return nil, err
	}

	password, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	registerType := detectRegisterType(input.RegisterType, input.Account)
	if registerType == "" {
		return nil, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "account must be a valid email or phone number")
	}

	user := &model.User{
		Nickname: defaultNickname(registerType, input.Account, input.Nickname),
		Status:   "enabled",
		Password: password,
	}

	switch registerType {
	case "email":
		if !options.EnableEmailRegistration {
			return nil, utils.NewAppError(http.StatusForbidden, http.StatusForbidden, "email registration is disabled")
		}
		email := normalizeEmail(input.Account)
		if !isValidEmail(email) {
			return nil, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "invalid email address")
		}
		if err := ensureEmailAvailable(ctx, s.userRepo, email); err != nil {
			return nil, err
		}
		user.Username = buildGeneratedUsername("email", email)
		user.Email = email
	case "phone":
		if !options.EnablePhoneRegistration {
			return nil, utils.NewAppError(http.StatusForbidden, http.StatusForbidden, "phone registration is disabled")
		}
		phone := normalizePhone(input.Account)
		if !isValidPhone(phone) {
			return nil, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "invalid phone number")
		}
		if err := s.verifyPhoneCode(ctx, phone, "register", input.SMSCode); err != nil {
			return nil, err
		}
		if err := ensurePhoneAvailable(ctx, s.userRepo, phone); err != nil {
			return nil, err
		}
		user.Username = buildGeneratedUsername("phone", phone)
		user.Phone = phone
	default:
		return nil, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "unsupported register type")
	}

	avatar, err := normalizeAvatarChoice("", s.avatarURLValidator, user.Username, user.Email, user.Phone, user.Nickname)
	if err != nil {
		return nil, err
	}
	user.Avatar = avatar

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	user, err = s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, err
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

	permissions, err := s.loadPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return buildProfileUser(user, permissions), nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID uint, input UpdateProfileInput) (*ProfileUser, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	avatar, err := normalizeAvatarChoice(input.Avatar, s.avatarURLValidator, user.Username, user.Email, user.Phone, user.Nickname)
	if err != nil {
		return nil, err
	}
	user.Avatar = avatar

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	permissions, err := s.loadPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return buildProfileUser(user, permissions), nil
}

func (s *authService) ResolvePermissions(ctx context.Context, userID uint) ([]string, error) {
	return s.loadPermissions(ctx, userID)
}

func (s *authService) Options(ctx context.Context) (AuthOptions, error) {
	return loadAuthOptions(ctx, s.defaults, s.authSettingRepo, s.cache)
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

	permissions, err := s.loadPermissions(ctx, user.ID)
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

func (s *authService) loadPermissions(ctx context.Context, userID uint) ([]string, error) {
	if s.cache != nil {
		var cached []string
		if err := s.cache.GetJSON(ctx, permissionCacheKey(userID), &cached); err == nil {
			return cached, nil
		} else if err != nil && err != repository.ErrCacheMiss {
			return nil, err
		}
	}

	permissions, err := s.userRepo.GetPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.SetJSON(ctx, permissionCacheKey(userID), permissions, permissionCacheTTL)
	}

	return permissions, nil
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
		Avatar:      resolveUserAvatar(user),
		Status:      user.Status,
		Roles:       roleCodes,
		Permissions: permissions,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func loadAuthOptions(ctx context.Context, defaults config.AuthConfig, repo repository.AuthSettingRepository, cache repository.CacheStore) (AuthOptions, error) {
	if cache != nil {
		var cached AuthOptions
		if err := cache.GetJSON(ctx, authOptionsCacheKey, &cached); err == nil {
			return cached, nil
		} else if err != nil && err != repository.ErrCacheMiss {
			return AuthOptions{}, err
		}
	}

	options := authOptionsFromDefaults(defaults)
	if repo == nil {
		return options, nil
	}

	setting, err := repo.Get(ctx)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			if cache != nil {
				_ = cache.SetJSON(ctx, authOptionsCacheKey, options, authOptionsCacheTTL)
			}
			return options, nil
		}
		return AuthOptions{}, err
	}

	options.EnableEmailLogin = setting.EnableEmailLogin
	options.EnablePhoneLogin = setting.EnablePhoneLogin
	options.EnableEmailRegistration = setting.EnableEmailRegistration
	options.EnablePhoneRegistration = setting.EnablePhoneRegistration
	options = normalizeAuthOptions(options)

	if cache != nil {
		_ = cache.SetJSON(ctx, authOptionsCacheKey, options, authOptionsCacheTTL)
	}

	return options, nil
}

func authOptionsFromDefaults(defaults config.AuthConfig) AuthOptions {
	return normalizeAuthOptions(AuthOptions{
		EnableUsernameLogin:     true,
		EnableEmailLogin:        defaults.EnableEmailLogin,
		EnablePhoneLogin:        defaults.EnablePhoneLogin,
		EnableEmailRegistration: defaults.EnableEmailRegistration,
		EnablePhoneRegistration: defaults.EnablePhoneRegistration,
	})
}

func normalizeAuthOptions(options AuthOptions) AuthOptions {
	options.EnableUsernameLogin = true
	if !options.EnableEmailLogin {
		options.EnableEmailRegistration = false
	}
	if !options.EnablePhoneLogin {
		options.EnablePhoneRegistration = false
	}
	return options
}

func (s *authService) findUserForLogin(ctx context.Context, account string, loginType string, options AuthOptions) (*model.User, string, error) {
	account = strings.TrimSpace(account)
	if account == "" {
		return nil, "", utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "account is required")
	}

	switch strings.TrimSpace(loginType) {
	case "username":
		user, err := s.userRepo.GetByUsername(ctx, account)
		return user, "username", err
	case "email":
		if !options.EnableEmailLogin {
			return nil, "", utils.NewAppError(http.StatusForbidden, http.StatusForbidden, "email login is disabled")
		}
		email := normalizeEmail(account)
		if !isValidEmail(email) {
			return nil, "", utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "invalid email address")
		}
		user, err := s.userRepo.GetByEmail(ctx, email)
		return user, "email", err
	case "phone":
		if !options.EnablePhoneLogin {
			return nil, "", utils.NewAppError(http.StatusForbidden, http.StatusForbidden, "phone login is disabled")
		}
		phone := normalizePhone(account)
		if !isValidPhone(phone) {
			return nil, "", utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "invalid phone number")
		}
		user, err := s.userRepo.GetByPhone(ctx, phone)
		return user, "phone", err
	case "":
		return s.findUserByFallback(ctx, account, options)
	default:
		return nil, "", utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "unsupported login type")
	}
}

func (s *authService) findUserByFallback(ctx context.Context, account string, options AuthOptions) (*model.User, string, error) {
	if options.EnableEmailLogin {
		email := normalizeEmail(account)
		if isValidEmail(email) {
			user, err := s.userRepo.GetByEmail(ctx, email)
			if err == nil || !errors.Is(err, utils.ErrNotFound) {
				return user, "email", err
			}
		}
	}

	if options.EnablePhoneLogin {
		phone := normalizePhone(account)
		if isValidPhone(phone) {
			user, err := s.userRepo.GetByPhone(ctx, phone)
			if err == nil || !errors.Is(err, utils.ErrNotFound) {
				return user, "phone", err
			}
		}
	}

	user, err := s.userRepo.GetByUsername(ctx, account)
	return user, "username", err
}

func (s *authService) verifyPhoneCode(ctx context.Context, phone string, purpose string, smsCode string) error {
	if s.smsVerifier == nil {
		return utils.NewAppError(http.StatusServiceUnavailable, http.StatusServiceUnavailable, "phone verification is not configured")
	}
	if !isValidPhone(normalizePhone(phone)) {
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "invalid phone number")
	}
	if strings.TrimSpace(smsCode) == "" {
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "sms verification code is required")
	}
	return s.smsVerifier.VerifyCode(ctx, VerifySMSCodeInput{
		Phone:   phone,
		Purpose: purpose,
		Code:    smsCode,
	})
}

func detectRegisterType(registerType string, account string) string {
	switch strings.TrimSpace(registerType) {
	case "email", "phone":
		return strings.TrimSpace(registerType)
	}

	account = strings.TrimSpace(account)
	switch {
	case isValidEmail(normalizeEmail(account)):
		return "email"
	case isValidPhone(normalizePhone(account)):
		return "phone"
	default:
		return ""
	}
}

func ensureEmailAvailable(ctx context.Context, userRepo repository.UserRepository, email string) error {
	_, err := userRepo.GetByEmail(ctx, email)
	if err == nil {
		return utils.NewAppError(http.StatusConflict, http.StatusConflict, "email has already been registered")
	}
	if errors.Is(err, utils.ErrNotFound) {
		return nil
	}
	return err
}

func ensurePhoneAvailable(ctx context.Context, userRepo repository.UserRepository, phone string) error {
	_, err := userRepo.GetByPhone(ctx, phone)
	if err == nil {
		return utils.NewAppError(http.StatusConflict, http.StatusConflict, "phone has already been registered")
	}
	if errors.Is(err, utils.ErrNotFound) {
		return nil
	}
	return err
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizePhone(value string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	return replacer.Replace(strings.TrimSpace(value))
}

func isValidEmail(value string) bool {
	return emailPattern.MatchString(value)
}

func isValidPhone(value string) bool {
	return phonePattern.MatchString(value)
}

func buildGeneratedUsername(prefix string, account string) string {
	clean := strings.ToLower(strings.TrimSpace(account))
	clean = strings.NewReplacer("@", "_", ".", "_", "+", "_", "-", "_").Replace(clean)
	clean = generatedNameSanitizer.ReplaceAllString(clean, "")
	if clean == "" {
		clean = prefix
	}
	if len(clean) > 24 {
		clean = clean[:24]
	}
	return fmt.Sprintf("%s_%s", clean, uuid.NewString()[:8])
}

func defaultNickname(registerType string, account string, nickname string) string {
	nickname = strings.TrimSpace(nickname)
	if nickname != "" {
		return nickname
	}

	switch registerType {
	case "email":
		local := strings.SplitN(normalizeEmail(account), "@", 2)[0]
		if local != "" {
			return local
		}
	case "phone":
		phone := normalizePhone(account)
		if len(phone) > 4 {
			return "user_" + phone[len(phone)-4:]
		}
		if phone != "" {
			return "user_" + phone
		}
	}

	return "new_user"
}
