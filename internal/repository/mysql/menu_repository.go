package mysql

import (
	"context"
	"errors"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"gorm.io/gorm"
)

type menuRepository struct {
	db *gorm.DB
}

// NewMenuRepository creates a MySQL menu repository.
func NewMenuRepository(db *gorm.DB) repository.MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) GetByID(ctx context.Context, id uint) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.WithContext(ctx).First(&menu, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &menu, nil
}

func (r *menuRepository) List(ctx context.Context, filter repository.MenuFilter) ([]model.Menu, error) {
	var menus []model.Menu
	query := r.db.WithContext(ctx)
	if filter.Keyword != "" {
		keyword := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("name LIKE ? OR title LIKE ? OR permission LIKE ?", keyword, keyword, keyword)
	}
	if err := query.Order("parent_id ASC, sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

func (r *menuRepository) Create(ctx context.Context, menu *model.Menu) error {
	if err := r.db.WithContext(ctx).Create(menu).Error; err != nil {
		return err
	}
	return nil
}

func (r *menuRepository) Update(ctx context.Context, menu *model.Menu) error {
	if err := r.db.WithContext(ctx).Updates(menu).Error; err != nil {
		return err
	}
	return nil
}

func (r *menuRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Menu{}, id).Error
}

