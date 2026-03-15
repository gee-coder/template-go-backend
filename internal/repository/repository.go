package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// UserFilter defines user list filters.
type UserFilter struct {
	Keyword string
	Status  string
}

// RoleFilter defines role list filters.
type RoleFilter struct {
	Keyword string
	Status  string
}

// MenuFilter defines menu list filters.
type MenuFilter struct {
	Keyword string
}

// UserRepository defines data access of users.
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	List(ctx context.Context, filter UserFilter) ([]model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	ReplaceRoles(ctx context.Context, userID uint, roleIDs []uint) error
	GetPermissions(ctx context.Context, userID uint) ([]string, error)
}

// AuthSettingRepository defines data access of auth settings.
type AuthSettingRepository interface {
	Get(ctx context.Context) (*model.AuthSetting, error)
	Save(ctx context.Context, setting *model.AuthSetting) error
}

// RoleRepository defines data access of roles.
type RoleRepository interface {
	GetByID(ctx context.Context, id uint) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
	List(ctx context.Context, filter RoleFilter) ([]model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uint) error
	ReplaceMenus(ctx context.Context, roleID uint, menuIDs []uint) error
}

// MenuRepository defines data access of menus.
type MenuRepository interface {
	GetByID(ctx context.Context, id uint) (*model.Menu, error)
	List(ctx context.Context, filter MenuFilter) ([]model.Menu, error)
	Create(ctx context.Context, menu *model.Menu) error
	Update(ctx context.Context, menu *model.Menu) error
	Delete(ctx context.Context, id uint) error
}

// ContactSubmissionRepository defines data access of contacts.
type ContactSubmissionRepository interface {
	Create(ctx context.Context, submission *model.ContactSubmission) error
}

// TokenStore stores refresh tokens.
type TokenStore interface {
	Save(ctx context.Context, refreshToken string, userID uint, ttl time.Duration) error
	Get(ctx context.Context, refreshToken string) (uint, error)
	Delete(ctx context.Context, refreshToken string) error
}

// NewMySQL creates a GORM DB instance.
func NewMySQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.Default.LogMode(glogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("mysql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	return db, nil
}

type redisTokenStore struct {
	client *redis.Client
	prefix string
}

// NewRedisTokenStore creates a token store backed by Redis.
func NewRedisTokenStore(cfg config.RedisConfig) (TokenStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &redisTokenStore{
		client: client,
		prefix: cfg.KeyPrefix,
	}, nil
}

func (s *redisTokenStore) Save(ctx context.Context, refreshToken string, userID uint, ttl time.Duration) error {
	return s.client.Set(ctx, s.key(refreshToken), userID, ttl).Err()
}

func (s *redisTokenStore) Get(ctx context.Context, refreshToken string) (uint, error) {
	value, err := s.client.Get(ctx, s.key(refreshToken)).Uint64()
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func (s *redisTokenStore) Delete(ctx context.Context, refreshToken string) error {
	return s.client.Del(ctx, s.key(refreshToken)).Err()
}

func (s *redisTokenStore) key(refreshToken string) string {
	return s.prefix + "refresh:" + refreshToken
}
