package mysql

import (
	"context"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"gorm.io/gorm"
)

type loginAuditRepository struct {
	db *gorm.DB
}

// NewLoginAuditRepository creates a MySQL login audit repository.
func NewLoginAuditRepository(db *gorm.DB) repository.LoginAuditRepository {
	return &loginAuditRepository{db: db}
}

func (r *loginAuditRepository) Create(ctx context.Context, item *model.LoginAuditLog) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *loginAuditRepository) List(ctx context.Context, filter repository.LoginAuditFilter) ([]model.LoginAuditLog, error) {
	var items []model.LoginAuditLog
	query := r.db.WithContext(ctx).Model(&model.LoginAuditLog{})

	if filter.Keyword != "" {
		keyword := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("account LIKE ? OR username LIKE ? OR ip LIKE ?", keyword, keyword, keyword)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.LoginType != "" {
		query = query.Where("login_type = ?", filter.LoginType)
	}

	if err := query.Order("id DESC").Limit(200).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
