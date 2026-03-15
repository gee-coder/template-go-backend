package mysql

import (
	"context"
	"errors"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"gorm.io/gorm"
)

type authSettingRepository struct {
	db *gorm.DB
}

// NewAuthSettingRepository creates an auth settings repository.
func NewAuthSettingRepository(db *gorm.DB) repository.AuthSettingRepository {
	return &authSettingRepository{db: db}
}

func (r *authSettingRepository) Get(ctx context.Context) (*model.AuthSetting, error) {
	var setting model.AuthSetting
	if err := r.db.WithContext(ctx).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &setting, nil
}

func (r *authSettingRepository) Save(ctx context.Context, setting *model.AuthSetting) error {
	if setting.ID == 0 {
		return r.db.WithContext(ctx).Create(setting).Error
	}
	return r.db.WithContext(ctx).Save(setting).Error
}
