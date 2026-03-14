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

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a MySQL role repository.
func NewRoleRepository(db *gorm.DB) repository.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Preload("Menus").First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Preload("Menus").Where("code = ?", code).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context, filter repository.RoleFilter) ([]model.Role, error) {
	var roles []model.Role
	query := r.db.WithContext(ctx).Preload("Menus")
	if filter.Keyword != "" {
		keyword := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", keyword, keyword)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if err := query.Order("id DESC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return utils.ErrConflict
		}
		return err
	}
	return nil
}

func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	if err := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Updates(role).Error; err != nil {
		return err
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

func (r *roleRepository) ReplaceMenus(ctx context.Context, roleID uint, menuIDs []uint) error {
	var role model.Role
	if err := r.db.WithContext(ctx).First(&role, roleID).Error; err != nil {
		return err
	}

	var menus []model.Menu
	if len(menuIDs) > 0 {
		if err := r.db.WithContext(ctx).Find(&menus, menuIDs).Error; err != nil {
			return err
		}
	}

	return r.db.WithContext(ctx).Model(&role).Association("Menus").Replace(menus)
}

