package handler

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/service"
)

// AuthService is the auth handler dependency contract.
type AuthService interface {
	Login(ctx context.Context, account string, password string, loginType string) (*service.TokenPayload, error)
	Register(ctx context.Context, input service.RegisterInput) (*service.TokenPayload, error)
	Refresh(ctx context.Context, refreshToken string) (*service.TokenPayload, error)
	Logout(ctx context.Context, refreshToken string) error
	Profile(ctx context.Context, userID uint) (*service.ProfileUser, error)
	UpdateProfile(ctx context.Context, userID uint, input service.UpdateProfileInput) (*service.ProfileUser, error)
	ResolvePermissions(ctx context.Context, userID uint) ([]string, error)
	Options(ctx context.Context) (service.AuthOptions, error)
}

// AuthSettingService is the auth setting handler dependency contract.
type AuthSettingService interface {
	Get(ctx context.Context) (service.AuthOptions, error)
	Update(ctx context.Context, input service.UpdateAuthSettingInput) (service.AuthOptions, error)
}

// UserService is the user handler dependency contract.
type UserService interface {
	List(ctx context.Context, filter repository.UserFilter) ([]model.User, error)
	Create(ctx context.Context, input service.CreateUserInput) (*model.User, error)
	Update(ctx context.Context, id uint, input service.UpdateUserInput) (*model.User, error)
	Delete(ctx context.Context, id uint) error
}

// RoleService is the role handler dependency contract.
type RoleService interface {
	List(ctx context.Context, filter repository.RoleFilter) ([]model.Role, error)
	Create(ctx context.Context, input service.CreateRoleInput) (*model.Role, error)
	Update(ctx context.Context, id uint, input service.UpdateRoleInput) (*model.Role, error)
	Delete(ctx context.Context, id uint) error
}

// MenuService is the menu handler dependency contract.
type MenuService interface {
	List(ctx context.Context, filter repository.MenuFilter) ([]service.MenuNode, error)
	Create(ctx context.Context, input service.CreateMenuInput) (*model.Menu, error)
	Update(ctx context.Context, id uint, input service.UpdateMenuInput) (*model.Menu, error)
	Delete(ctx context.Context, id uint) error
}

// ContactService is the contact handler dependency contract.
type ContactService interface {
	Create(ctx context.Context, input service.CreateContactSubmissionInput) (*model.ContactSubmission, error)
}
