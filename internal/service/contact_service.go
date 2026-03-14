package service

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
)

// ContactService provides contact submission capabilities.
type ContactService interface {
	Create(ctx context.Context, input CreateContactSubmissionInput) (*model.ContactSubmission, error)
}

// CreateContactSubmissionInput is the input of creating a contact submission.
type CreateContactSubmissionInput struct {
	Name    string
	Email   string
	Phone   string
	Company string
	Message string
	Source  string
}

type contactService struct {
	repo repository.ContactSubmissionRepository
}

// NewContactService creates the contact service.
func NewContactService(repo repository.ContactSubmissionRepository) ContactService {
	return &contactService{repo: repo}
}

func (s *contactService) Create(ctx context.Context, input CreateContactSubmissionInput) (*model.ContactSubmission, error) {
	submission := &model.ContactSubmission{
		Name:    input.Name,
		Email:   input.Email,
		Phone:   input.Phone,
		Company: input.Company,
		Message: input.Message,
		Source:  input.Source,
	}
	if submission.Source == "" {
		submission.Source = "website"
	}

	if err := s.repo.Create(ctx, submission); err != nil {
		return nil, err
	}
	return submission, nil
}

