package mysql

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"gorm.io/gorm"
)

type brandingSettingRepository struct {
	db *gorm.DB
}

// NewBrandingSettingRepository creates a branding settings repository.
func NewBrandingSettingRepository(db *gorm.DB) repository.BrandingSettingRepository {
	return &brandingSettingRepository{db: db}
}

func (r *brandingSettingRepository) Get(ctx context.Context) (*model.BrandingSetting, error) {
	var setting model.BrandingSetting
	if err := r.db.WithContext(ctx).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrNotFound
		}
		return nil, err
	}
	return &setting, nil
}

func (r *brandingSettingRepository) Save(ctx context.Context, setting *model.BrandingSetting) error {
	if setting.ID == 0 {
		return r.db.WithContext(ctx).Create(setting).Error
	}
	return r.db.WithContext(ctx).Save(setting).Error
}
