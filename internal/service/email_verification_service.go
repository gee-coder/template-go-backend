package service

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"go.uber.org/zap"
)

// SendEmailCodeInput describes the public email send-code payload.
type SendEmailCodeInput struct {
	Email   string
	Purpose string
}

// VerifyEmailCodeInput describes the email verify-code payload.
type VerifyEmailCodeInput struct {
	Email   string
	Purpose string
	Code    string
}

// EmailVerificationService provides configurable email verification flows.
type EmailVerificationService interface {
	SendCode(ctx context.Context, input SendEmailCodeInput) (SMSVerificationPayload, error)
	VerifyCode(ctx context.Context, input VerifyEmailCodeInput) error
}

type emailVerificationService struct {
	cfg      config.MailConfig
	cache    repository.CacheStore
	provider emailSender
	logger   *zap.Logger
}

type emailSender interface {
	Name() string
	Send(ctx context.Context, message emailMessage) error
}

type emailMessage struct {
	To         string
	Code       string
	Purpose    string
	TTLMinutes int
}

type mockEmailSender struct{}

type smtpEmailSender struct {
	cfg config.MailConfig
}

type emailVerificationRecord struct {
	Code      string    `json:"code"`
	Email     string    `json:"email"`
	Purpose   string    `json:"purpose"`
	SentAt    time.Time `json:"sentAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// NewEmailVerificationService creates the email verification service.
func NewEmailVerificationService(cfg config.MailConfig, cache repository.CacheStore, logger *zap.Logger) (EmailVerificationService, error) {
	provider, err := newEmailSender(cfg)
	if err != nil {
		return nil, err
	}
	return &emailVerificationService{
		cfg:      cfg,
		cache:    cache,
		provider: provider,
		logger:   logger,
	}, nil
}

func (s *emailVerificationService) SendCode(ctx context.Context, input SendEmailCodeInput) (SMSVerificationPayload, error) {
	if s.cache == nil {
		return SMSVerificationPayload{}, fmt.Errorf("email verification cache is not configured")
	}

	email := normalizeEmail(input.Email)
	if !isValidEmail(email) {
		return SMSVerificationPayload{}, utils.NewAppError(400, 400, "invalid email address")
	}

	purpose := normalizeVerificationPurpose(input.Purpose)
	if purpose == "" {
		return SMSVerificationPayload{}, utils.NewAppError(400, 400, "email verification purpose is required")
	}

	if err := s.ensureCooldown(ctx, email, purpose); err != nil {
		return SMSVerificationPayload{}, err
	}

	code, err := s.generateCode()
	if err != nil {
		return SMSVerificationPayload{}, err
	}

	ttl := s.cfg.ResolvedCodeTTL()
	record := emailVerificationRecord{
		Code:      code,
		Email:     email,
		Purpose:   purpose,
		SentAt:    time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	if err := s.provider.Send(ctx, emailMessage{
		To:         email,
		Code:       code,
		Purpose:    purpose,
		TTLMinutes: maxInt(1, int(ttl.Minutes())),
	}); err != nil {
		return SMSVerificationPayload{}, err
	}

	if err := s.cache.SetJSON(ctx, emailCodeCacheKey(email, purpose), record, ttl); err != nil {
		return SMSVerificationPayload{}, err
	}
	if err := s.cache.SetJSON(ctx, emailCooldownCacheKey(email, purpose), map[string]string{"email": email}, s.cfg.ResolvedCooldown()); err != nil {
		return SMSVerificationPayload{}, err
	}

	if s.logger != nil {
		s.logger.Info("email verification code sent",
			zap.String("provider", s.provider.Name()),
			zap.String("email", maskEmail(email)),
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

func (s *emailVerificationService) VerifyCode(ctx context.Context, input VerifyEmailCodeInput) error {
	if s.cache == nil {
		return fmt.Errorf("email verification cache is not configured")
	}

	email := normalizeEmail(input.Email)
	if !isValidEmail(email) {
		return utils.NewAppError(400, 400, "invalid email address")
	}

	purpose := normalizeVerificationPurpose(input.Purpose)
	if purpose == "" {
		return utils.NewAppError(400, 400, "email verification purpose is required")
	}

	code := strings.TrimSpace(input.Code)
	if len(code) < 4 {
		return utils.NewAppError(400, 400, "verification code is invalid")
	}

	var record emailVerificationRecord
	if err := s.cache.GetJSON(ctx, emailCodeCacheKey(email, purpose), &record); err != nil {
		if err == repository.ErrCacheMiss {
			return utils.NewAppError(400, 400, "verification code has expired")
		}
		return err
	}

	if record.Code != code {
		return utils.NewAppError(400, 400, "verification code is incorrect")
	}

	return s.cache.Delete(ctx, emailCodeCacheKey(email, purpose))
}

func (s *emailVerificationService) ensureCooldown(ctx context.Context, email string, purpose string) error {
	var marker map[string]string
	if err := s.cache.GetJSON(ctx, emailCooldownCacheKey(email, purpose), &marker); err == nil {
		return utils.NewAppError(429, 429, "verification code requested too frequently")
	} else if err != repository.ErrCacheMiss {
		return err
	}
	return nil
}

func (s *emailVerificationService) generateCode() (string, error) {
	if fixed := strings.TrimSpace(s.cfg.Mock.FixedCode); fixed != "" {
		return fixed, nil
	}

	buffer := make([]byte, smsVerificationCodeLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	for idx := range buffer {
		buffer[idx] = '0' + (buffer[idx] % 10)
	}
	return string(buffer), nil
}

func newEmailSender(cfg config.MailConfig) (emailSender, error) {
	switch cfg.ResolvedProvider() {
	case "mock":
		return mockEmailSender{}, nil
	case "smtp":
		if strings.TrimSpace(cfg.FromAddress) == "" {
			return nil, fmt.Errorf("mail fromAddress is required for smtp provider")
		}
		if strings.TrimSpace(cfg.SMTP.Host) == "" || cfg.SMTP.Port <= 0 {
			return nil, fmt.Errorf("smtp host and port are required")
		}
		return smtpEmailSender{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unsupported mail provider %q", cfg.ResolvedProvider())
	}
}

func (mockEmailSender) Name() string {
	return "mock"
}

func (mockEmailSender) Send(_ context.Context, _ emailMessage) error {
	return nil
}

func (s smtpEmailSender) Name() string {
	return "smtp"
}

func (s smtpEmailSender) Send(ctx context.Context, message emailMessage) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTP.Host, s.cfg.SMTP.Port)
	body := buildSMTPMessage(s.cfg, message)

	if s.cfg.SMTP.UseTLS {
		dialer := &net.Dialer{}
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName: strings.TrimSpace(s.cfg.SMTP.Host),
		})
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, strings.TrimSpace(s.cfg.SMTP.Host))
		if err != nil {
			return err
		}
		defer client.Quit()

		if err := applySMTPAuth(client, s.cfg); err != nil {
			return err
		}
		if err := client.Mail(strings.TrimSpace(s.cfg.FromAddress)); err != nil {
			return err
		}
		if err := client.Rcpt(message.To); err != nil {
			return err
		}
		writer, err := client.Data()
		if err != nil {
			return err
		}
		defer writer.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err = writer.Write([]byte(body))
			return err
		}
	}

	auth := smtp.PlainAuth("", s.cfg.SMTP.Username, s.cfg.SMTP.Password, s.cfg.SMTP.Host)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return smtp.SendMail(addr, auth, strings.TrimSpace(s.cfg.FromAddress), []string{message.To}, []byte(body))
	}
}

func applySMTPAuth(client *smtp.Client, cfg config.MailConfig) error {
	if strings.TrimSpace(cfg.SMTP.Username) == "" {
		return nil
	}
	auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)
	return client.Auth(auth)
}

func buildSMTPMessage(cfg config.MailConfig, message emailMessage) string {
	subject := "Nex verification code"
	if message.Purpose == "two_factor" {
		subject = "Nex two-factor verification code"
	}
	from := strings.TrimSpace(cfg.FromAddress)
	if name := strings.TrimSpace(cfg.FromName); name != "" {
		from = fmt.Sprintf("%s <%s>", name, from)
	}

	body := fmt.Sprintf(
		"Your verification code is %s. It expires in %d minutes.\r\nPurpose: %s\r\n",
		message.Code,
		message.TTLMinutes,
		message.Purpose,
	)
	return strings.Join([]string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", message.To),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
}

func emailCodeCacheKey(email string, purpose string) string {
	return fmt.Sprintf("email:code:%s:%s", purpose, email)
}

func emailCooldownCacheKey(email string, purpose string) string {
	return fmt.Sprintf("email:cooldown:%s:%s", purpose, email)
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) <= 2 {
		return email
	}
	return parts[0][:2] + "***@" + parts[1]
}
