package mysql

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"gorm.io/gorm"
)

type contactSubmissionRepository struct {
	db *gorm.DB
}

// NewContactSubmissionRepository creates a MySQL contact repository.
func NewContactSubmissionRepository(db *gorm.DB) repository.ContactSubmissionRepository {
	return &contactSubmissionRepository{db: db}
}

func (r *contactSubmissionRepository) Create(ctx context.Context, submission *model.ContactSubmission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

