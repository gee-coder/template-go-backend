package service

import (
	"context"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
)

// CreateLoginAuditInput describes a login audit event to persist.
type CreateLoginAuditInput struct {
	UserID    *uint
	Username  string
	Account   string
	LoginType string
	Status    string
	IP        string
	UserAgent string
	Message   string
}

// LoginAuditService provides login audit capabilities.
type LoginAuditService interface {
	Create(ctx context.Context, input CreateLoginAuditInput) error
	List(ctx context.Context, filter repository.LoginAuditFilter) ([]model.LoginAuditLog, error)
}

type loginAuditService struct {
	repo repository.LoginAuditRepository
}

// NewLoginAuditService creates a login audit service.
func NewLoginAuditService(repo repository.LoginAuditRepository) LoginAuditService {
	return &loginAuditService{repo: repo}
}

func (s *loginAuditService) Create(ctx context.Context, input CreateLoginAuditInput) error {
	item := &model.LoginAuditLog{
		UserID:    input.UserID,
		Username:  strings.TrimSpace(input.Username),
		Account:   strings.TrimSpace(input.Account),
		LoginType: normalizeLoginAuditType(input.LoginType),
		Status:    normalizeLoginAuditStatus(input.Status),
		IP:        strings.TrimSpace(input.IP),
		UserAgent: limitString(strings.TrimSpace(input.UserAgent), 255),
		Message:   limitString(strings.TrimSpace(input.Message), 255),
	}

	return s.repo.Create(ctx, item)
}

func (s *loginAuditService) List(ctx context.Context, filter repository.LoginAuditFilter) ([]model.LoginAuditLog, error) {
	return s.repo.List(ctx, filter)
}

func normalizeLoginAuditType(value string) string {
	switch strings.TrimSpace(value) {
	case "email", "phone", "username":
		return strings.TrimSpace(value)
	default:
		return "username"
	}
}

func normalizeLoginAuditStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "success", "failure":
		return strings.TrimSpace(value)
	default:
		return "failure"
	}
}

func limitString(value string, size int) string {
	if len(value) <= size {
		return value
	}
	return value[:size]
}
