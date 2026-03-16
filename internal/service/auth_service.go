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

const usernameLoginFailureTTL = 30 * time.Minute

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
	EnableTwoFactor         bool `json:"enableTwoFactor"`
}

// LoginInput describes the public login payload.
type LoginInput struct {
	Account          string
	LoginType        string
	Password         string
	VerificationCode string
	CaptchaID        string
	CaptchaCode      string
	TwoFactorCode    string
}

// RegisterInput describes the public register payload.
type RegisterInput struct {
	Account          string
	RegisterType     string
	Nickname         string
	Password         string
	VerificationCode string
	CaptchaID        string
	CaptchaCode      string
	SMSCode          string
}

// SendTwoFactorCodeInput describes the two-factor send-code payload.
type SendTwoFactorCodeInput struct {
	Account   string
	LoginType string
}

// TwoFactorCodePayload describes the two-factor send-code response.
type TwoFactorCodePayload struct {
	Channel    string `json:"channel"`
	Target     string `json:"target"`
	Provider   string `json:"provider"`
	CooldownIn int64  `json:"cooldownIn"`
	DebugCode  string `json:"debugCode,omitempty"`
}

// UpdateProfileInput describes self profile updates.
type UpdateProfileInput struct {
	Avatar string
}

// AuthService provides auth capabilities.
type AuthService interface {
	Login(ctx context.Context, input LoginInput) (*TokenPayload, error)
	Register(ctx context.Context, input RegisterInput) (*TokenPayload, error)
	SendTwoFactorCode(ctx context.Context, input SendTwoFactorCodeInput) (TwoFactorCodePayload, error)
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
	smsService         SMSVerificationService
	emailService       EmailVerificationService
	captchaService     ImageCaptchaService
	avatarURLValidator func(string) bool
}

// NewAuthService creates the auth service.
func NewAuthService(
	cfg config.JWTConfig,
	defaults config.AuthConfig,
	authSettingRepo repository.AuthSettingRepository,
	userRepo repository.UserRepository,
	tokenRepo repository.TokenStore,
	cache repository.CacheStore,
	smsService SMSVerificationService,
	emailService EmailVerificationService,
	captchaService ImageCaptchaService,
	avatarURLValidator func(string) bool,
) AuthService {
	return &authService{
		cfg:                cfg,
		defaults:           defaults,
		authSettingRepo:    authSettingRepo,
		userRepo:           userRepo,
		tokenRepo:          tokenRepo,
		cache:              cache,
		smsService:         smsService,
		emailService:       emailService,
		captchaService:     captchaService,
		avatarURLValidator: avatarURLValidator,
	}
}

func (s *authService) Login(ctx context.Context, input LoginInput) (*TokenPayload, error) {
	options, err := s.Options(ctx)
	if err != nil {
		return nil, err
	}

	account := strings.TrimSpace(input.Account)
	if account == "" {
		return nil, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "account is required")
	}

	user, resolvedLoginType, err := s.findUserForLogin(ctx, account, input.LoginType, options)
	if err != nil {
		if resolvedLoginType == "username" {
			_ = s.recordUsernameLoginFailure(ctx, account)
		}
		if errors.Is(err, utils.ErrNotFound) {
			return nil, utils.ErrInvalidCredential
		}
		return nil, err
	}

	switch resolvedLoginType {
	case "username":
		if err := s.verifyUsernameLogin(ctx, user, input, options); err != nil {
			return nil, err
		}
	case "email":
		if err := s.verifyEmailLogin(ctx, user, input, options); err != nil {
			return nil, err
		}
	case "phone":
		if err := s.verifyPhoneLogin(ctx, user, input, options); err != nil {
			return nil, err
		}
	default:
		return nil, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "unsupported login type")
	}

	if user.Status != "enabled" {
		return nil, utils.ErrForbidden
	}

	if resolvedLoginType == "username" {
		_ = s.clearUsernameLoginFailures(ctx, account)
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
		if err := s.verifyImageCaptcha(ctx, input.CaptchaID, input.CaptchaCode); err != nil {
			return nil, err
		}
		if err := s.verifyEmailCode(ctx, email, "register", registerVerificationCode(input)); err != nil {
			return nil, err
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
		if err := s.verifyImageCaptcha(ctx, input.CaptchaID, input.CaptchaCode); err != nil {
			return nil, err
		}
		if err := s.verifyPhoneCode(ctx, phone, "register", registerVerificationCode(input)); err != nil {
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

func (s *authService) SendTwoFactorCode(ctx context.Context, input SendTwoFactorCodeInput) (TwoFactorCodePayload, error) {
	options, err := s.Options(ctx)
	if err != nil {
		return TwoFactorCodePayload{}, err
	}
	if !options.EnableTwoFactor {
		return TwoFactorCodePayload{}, utils.NewAppError(http.StatusForbidden, http.StatusForbidden, "two-factor authentication is disabled")
	}

	account := strings.TrimSpace(input.Account)
	if account == "" {
		return TwoFactorCodePayload{}, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "account is required")
	}

	loginType := strings.TrimSpace(input.LoginType)
	if loginType != "" && loginType != "username" {
		return TwoFactorCodePayload{}, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "two-factor code is only required for username login")
	}

	user, err := s.userRepo.GetByUsername(ctx, account)
	if err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			return TwoFactorCodePayload{}, utils.NewAppError(http.StatusNotFound, http.StatusNotFound, "user account does not exist")
		}
		return TwoFactorCodePayload{}, err
	}

	channel, target, payload, err := s.sendTwoFactorCodeForUser(ctx, user)
	if err != nil {
		return TwoFactorCodePayload{}, err
	}

	return TwoFactorCodePayload{
		Channel:    channel,
		Target:     target,
		Provider:   payload.Provider,
		CooldownIn: payload.CooldownIn,
		DebugCode:  payload.DebugCode,
	}, nil
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

func (s *authService) verifyUsernameLogin(ctx context.Context, user *model.User, input LoginInput, options AuthOptions) error {
	needsCaptcha, err := s.shouldRequireUsernameCaptcha(ctx, input.Account)
	if err != nil {
		return err
	}
	if needsCaptcha {
		if err := s.verifyImageCaptcha(ctx, input.CaptchaID, input.CaptchaCode); err != nil {
			return err
		}
	}

	if !utils.CheckPassword(input.Password, user.Password) {
		_ = s.recordUsernameLoginFailure(ctx, input.Account)
		return utils.ErrInvalidCredential
	}

	if options.EnableTwoFactor {
		if err := s.verifyUsernameTwoFactor(ctx, user, input.TwoFactorCode); err != nil {
			return err
		}
	}

	return nil
}

func (s *authService) verifyEmailLogin(ctx context.Context, user *model.User, input LoginInput, options AuthOptions) error {
	if err := s.verifyImageCaptcha(ctx, input.CaptchaID, input.CaptchaCode); err != nil {
		return err
	}
	if err := s.verifyEmailCode(ctx, user.Email, "login", input.VerificationCode); err != nil {
		return err
	}
	if options.EnableTwoFactor {
		if !utils.CheckPassword(input.Password, user.Password) {
			return utils.ErrInvalidCredential
		}
	}
	return nil
}

func (s *authService) verifyPhoneLogin(ctx context.Context, user *model.User, input LoginInput, options AuthOptions) error {
	if err := s.verifyImageCaptcha(ctx, input.CaptchaID, input.CaptchaCode); err != nil {
		return err
	}
	if err := s.verifyPhoneCode(ctx, user.Phone, "login", input.VerificationCode); err != nil {
		return err
	}
	if options.EnableTwoFactor {
		if !utils.CheckPassword(input.Password, user.Password) {
			return utils.ErrInvalidCredential
		}
	}
	return nil
}

func (s *authService) verifyUsernameTwoFactor(ctx context.Context, user *model.User, code string) error {
	if strings.TrimSpace(code) == "" {
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "two-factor verification code is required")
	}

	switch channel := preferredTwoFactorChannel(user); channel {
	case "phone":
		return s.verifyPhoneCode(ctx, user.Phone, "two_factor", code)
	case "email":
		return s.verifyEmailCode(ctx, user.Email, "two_factor", code)
	default:
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "user has no available second factor target")
	}
}

func (s *authService) sendTwoFactorCodeForUser(ctx context.Context, user *model.User) (string, string, SMSVerificationPayload, error) {
	switch channel := preferredTwoFactorChannel(user); channel {
	case "phone":
		payload, err := s.smsService.SendCode(ctx, SendSMSCodeInput{
			Phone:   user.Phone,
			Purpose: "two_factor",
		})
		return "phone", maskPhone(user.Phone), payload, err
	case "email":
		payload, err := s.emailService.SendCode(ctx, SendEmailCodeInput{
			Email:   user.Email,
			Purpose: "two_factor",
		})
		return "email", maskEmail(user.Email), payload, err
	default:
		return "", "", SMSVerificationPayload{}, utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "user has no available second factor target")
	}
}

func (s *authService) verifyImageCaptcha(ctx context.Context, captchaID string, captchaCode string) error {
	if s.captchaService == nil {
		return utils.NewAppError(http.StatusServiceUnavailable, http.StatusServiceUnavailable, "image captcha is not configured")
	}
	if strings.TrimSpace(captchaID) == "" || strings.TrimSpace(captchaCode) == "" {
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "image captcha is required")
	}
	return s.captchaService.Verify(ctx, captchaID, captchaCode)
}

func (s *authService) verifyPhoneCode(ctx context.Context, phone string, purpose string, code string) error {
	if s.smsService == nil {
		return utils.NewAppError(http.StatusServiceUnavailable, http.StatusServiceUnavailable, "phone verification is not configured")
	}
	if !isValidPhone(normalizePhone(phone)) {
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "invalid phone number")
	}
	if strings.TrimSpace(code) == "" {
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "verification code is required")
	}
	return s.smsService.VerifyCode(ctx, VerifySMSCodeInput{
		Phone:   phone,
		Purpose: purpose,
		Code:    code,
	})
}

func (s *authService) verifyEmailCode(ctx context.Context, email string, purpose string, code string) error {
	if s.emailService == nil {
		return utils.NewAppError(http.StatusServiceUnavailable, http.StatusServiceUnavailable, "email verification is not configured")
	}
	if !isValidEmail(normalizeEmail(email)) {
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "invalid email address")
	}
	if strings.TrimSpace(code) == "" {
		return utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "verification code is required")
	}
	return s.emailService.VerifyCode(ctx, VerifyEmailCodeInput{
		Email:   email,
		Purpose: purpose,
		Code:    code,
	})
}

func preferredTwoFactorChannel(user *model.User) string {
	if user == nil {
		return ""
	}
	if strings.TrimSpace(user.Phone) != "" {
		return "phone"
	}
	if strings.TrimSpace(user.Email) != "" {
		return "email"
	}
	return ""
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
	options.EnableTwoFactor = setting.EnableTwoFactor
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
		EnableTwoFactor:         defaults.EnableTwoFactor,
	})
}

func normalizeAuthOptions(options AuthOptions) AuthOptions {
	options.EnableUsernameLogin = true
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

func (s *authService) shouldRequireUsernameCaptcha(ctx context.Context, account string) (bool, error) {
	if s.cache == nil {
		return false, nil
	}

	var count int
	if err := s.cache.GetJSON(ctx, usernameLoginFailureKey(account), &count); err != nil {
		if err == repository.ErrCacheMiss {
			return false, nil
		}
		return false, err
	}
	return count >= 2, nil
}

func (s *authService) recordUsernameLoginFailure(ctx context.Context, account string) error {
	if s.cache == nil {
		return nil
	}

	var count int
	if err := s.cache.GetJSON(ctx, usernameLoginFailureKey(account), &count); err != nil && err != repository.ErrCacheMiss {
		return err
	}
	count++
	return s.cache.SetJSON(ctx, usernameLoginFailureKey(account), count, usernameLoginFailureTTL)
}

func (s *authService) clearUsernameLoginFailures(ctx context.Context, account string) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.Delete(ctx, usernameLoginFailureKey(account))
}

func usernameLoginFailureKey(account string) string {
	return "auth:username-failures:" + strings.ToLower(strings.TrimSpace(account))
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

func registerVerificationCode(input RegisterInput) string {
	if strings.TrimSpace(input.VerificationCode) != "" {
		return input.VerificationCode
	}
	return input.SMSCode
}
