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

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a MySQL user repository.
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Roles").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Roles").Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, filter repository.UserFilter) ([]model.User, error) {
	var users []model.User
	query := r.db.WithContext(ctx).Preload("Roles")

	if filter.Keyword != "" {
		keyword := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", keyword, keyword, keyword)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if err := query.Order("id DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return utils.ErrConflict
		}
		return err
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Updates(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) ReplaceRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return err
	}

	var roles []model.Role
	if len(roleIDs) > 0 {
		if err := r.db.WithContext(ctx).Find(&roles, roleIDs).Error; err != nil {
			return err
		}
	}

	return r.db.WithContext(ctx).Model(&user).Association("Roles").Replace(roles)
}

func (r *userRepository) GetPermissions(ctx context.Context, userID uint) ([]string, error) {
	type result struct {
		Permission string
	}

	var rows []result
	err := r.db.WithContext(ctx).
		Table("menus").
		Select("DISTINCT menus.permission AS permission").
		Joins("JOIN role_menus ON role_menus.menu_id = menus.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_menus.role_id").
		Where("user_roles.user_id = ? AND menus.permission <> ''", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	permissions := make([]string, 0, len(rows))
	for _, row := range rows {
		permissions = append(permissions, row.Permission)
	}
	return permissions, nil
}

