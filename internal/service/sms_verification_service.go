package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"go.uber.org/zap"
)

const (
	smsVerificationCodeLength = 6
)

// SendSMSCodeInput describes the public send-code payload.
type SendSMSCodeInput struct {
	Phone   string
	Purpose string
}

// VerifySMSCodeInput describes the public verify-code payload.
type VerifySMSCodeInput struct {
	Phone   string
	Purpose string
	Code    string
}

// SMSVerificationPayload describes the send-code response payload.
type SMSVerificationPayload struct {
	Provider   string `json:"provider"`
	ExpiresIn  int64  `json:"expiresIn"`
	CooldownIn int64  `json:"cooldownIn"`
	DebugCode  string `json:"debugCode,omitempty"`
}

// SMSVerificationService provides configurable SMS verification flows.
type SMSVerificationService interface {
	SendCode(ctx context.Context, input SendSMSCodeInput) (SMSVerificationPayload, error)
	VerifyCode(ctx context.Context, input VerifySMSCodeInput) error
}

type smsVerificationService struct {
	cfg      config.SMSConfig
	cache    repository.CacheStore
	provider smsSender
	logger   *zap.Logger
}

type smsSender interface {
	Name() string
	Send(ctx context.Context, message smsMessage) error
}

type smsMessage struct {
	Phone      string
	Code       string
	Purpose    string
	TTLMinutes int
}

type mockSMSSender struct{}

type smsVerificationRecord struct {
	Code      string    `json:"code"`
	Phone     string    `json:"phone"`
	Purpose   string    `json:"purpose"`
	SentAt    time.Time `json:"sentAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// NewSMSVerificationService creates the SMS verification service.
func NewSMSVerificationService(cfg config.SMSConfig, cache repository.CacheStore, logger *zap.Logger) (SMSVerificationService, error) {
	provider, err := newSMSSender(cfg)
	if err != nil {
		return nil, err
	}
	return &smsVerificationService{
		cfg:      cfg,
		cache:    cache,
		provider: provider,
		logger:   logger,
	}, nil
}

func (s *smsVerificationService) SendCode(ctx context.Context, input SendSMSCodeInput) (SMSVerificationPayload, error) {
	if s.cache == nil {
		return SMSVerificationPayload{}, fmt.Errorf("sms verification cache is not configured")
	}

	phone := normalizePhone(input.Phone)
	if !isValidPhone(phone) {
		return SMSVerificationPayload{}, utils.NewAppError(400, 400, "invalid phone number")
	}

	purpose := normalizeSMSPurpose(input.Purpose)
	if purpose == "" {
		return SMSVerificationPayload{}, utils.NewAppError(400, 400, "sms purpose is required")
	}

	if err := s.ensureCooldown(ctx, phone, purpose); err != nil {
		return SMSVerificationPayload{}, err
	}

	code, err := s.generateCode()
	if err != nil {
		return SMSVerificationPayload{}, err
	}

	ttl := s.cfg.ResolvedCodeTTL()
	record := smsVerificationRecord{
		Code:      code,
		Phone:     phone,
		Purpose:   purpose,
		SentAt:    time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	if err := s.provider.Send(ctx, smsMessage{
		Phone:      phone,
		Code:       code,
		Purpose:    purpose,
		TTLMinutes: maxInt(1, int(ttl.Minutes())),
	}); err != nil {
		return SMSVerificationPayload{}, err
	}

	if err := s.cache.SetJSON(ctx, smsCodeCacheKey(phone, purpose), record, ttl); err != nil {
		return SMSVerificationPayload{}, err
	}
	if err := s.cache.SetJSON(ctx, smsCooldownCacheKey(phone, purpose), map[string]string{"phone": phone}, s.cfg.ResolvedCooldown()); err != nil {
		return SMSVerificationPayload{}, err
	}

	if s.logger != nil {
		s.logger.Info("sms verification code sent",
			zap.String("provider", s.provider.Name()),
			zap.String("phone", maskPhone(phone)),
			zap.String("purpose", purpose),
		)
	}

	payload := SMSVerificationPayload{
		Provider:   s.provider.Name(),
		ExpiresIn:  int64(ttl.Seconds()),
		CooldownIn: int64(s.cfg.ResolvedCooldown().Seconds()),
	}
	if s.cfg.ResolvedProvider() == "mock" && s.cfg.Mock.RevealCode {
		payload.DebugCode = code
	}

	return payload, nil
}

func (s *smsVerificationService) VerifyCode(ctx context.Context, input VerifySMSCodeInput) error {
	if s.cache == nil {
		return fmt.Errorf("sms verification cache is not configured")
	}

	phone := normalizePhone(input.Phone)
	if !isValidPhone(phone) {
		return utils.NewAppError(400, 400, "invalid phone number")
	}

	purpose := normalizeSMSPurpose(input.Purpose)
	if purpose == "" {
		return utils.NewAppError(400, 400, "sms purpose is required")
	}

	code := strings.TrimSpace(input.Code)
	if len(code) < 4 {
		return utils.NewAppError(400, 400, "verification code is invalid")
	}

	var record smsVerificationRecord
	if err := s.cache.GetJSON(ctx, smsCodeCacheKey(phone, purpose), &record); err != nil {
		if err == repository.ErrCacheMiss {
			return utils.NewAppError(400, 400, "verification code has expired")
		}
		return err
	}

	if record.Code != code {
		return utils.NewAppError(400, 400, "verification code is incorrect")
	}

	if err := s.cache.Delete(ctx, smsCodeCacheKey(phone, purpose)); err != nil {
		return err
	}
	return nil
}

func (s *smsVerificationService) ensureCooldown(ctx context.Context, phone string, purpose string) error {
	if s.cache == nil {
		return fmt.Errorf("sms verification cache is not configured")
	}

	var marker map[string]string
	if err := s.cache.GetJSON(ctx, smsCooldownCacheKey(phone, purpose), &marker); err == nil {
		return utils.NewAppError(429, 429, "sms code requested too frequently")
	} else if err != repository.ErrCacheMiss {
		return err
	}
	return nil
}

func (s *smsVerificationService) generateCode() (string, error) {
	if fixed := strings.TrimSpace(s.cfg.Mock.FixedCode); fixed != "" {
		return fixed, nil
	}

	buffer := make([]byte, smsVerificationCodeLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	for i := range buffer {
		buffer[i] = '0' + (buffer[i] % 10)
	}
	return string(buffer), nil
}

func newSMSSender(cfg config.SMSConfig) (smsSender, error) {
	switch cfg.ResolvedProvider() {
	case "mock":
		return mockSMSSender{}, nil
	case "aliyun", "huawei":
		return nil, fmt.Errorf("sms provider %q is reserved in config but not enabled in the template yet", cfg.ResolvedProvider())
	default:
		return nil, fmt.Errorf("unsupported sms provider %q", cfg.ResolvedProvider())
	}
}

func (mockSMSSender) Name() string {
	return "mock"
}

func (mockSMSSender) Send(_ context.Context, _ smsMessage) error {
	return nil
}

func normalizeSMSPurpose(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "register", "login", "bind_phone", "reset_password":
		return value
	default:
		return ""
	}
}

func smsCodeCacheKey(phone string, purpose string) string {
	return fmt.Sprintf("sms:code:%s:%s", purpose, phone)
}

func smsCooldownCacheKey(phone string, purpose string) string {
	return fmt.Sprintf("sms:cooldown:%s:%s", purpose, phone)
}

func maskPhone(phone string) string {
	if len(phone) <= 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
