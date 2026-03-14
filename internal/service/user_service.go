package service

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
)

// UserService provides user management capabilities.
type UserService interface {
	List(ctx context.Context, filter repository.UserFilter) ([]model.User, error)
	Create(ctx context.Context, input CreateUserInput) (*model.User, error)
	Update(ctx context.Context, id uint, input UpdateUserInput) (*model.User, error)
	Delete(ctx context.Context, id uint) error
}

// CreateUserInput is the input of creating a user.
type CreateUserInput struct {
	Username string
	Nickname string
	Email    string
	Phone    string
	Status   string
	Password string
	RoleIDs  []uint
}

// UpdateUserInput is the input of updating a user.
type UpdateUserInput struct {
	Nickname string
	Email    string
	Phone    string
	Status   string
	Password string
	RoleIDs  []uint
}

type userService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

// NewUserService creates the user service.
func NewUserService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) UserService {
	return &userService{userRepo: userRepo, roleRepo: roleRepo}
}

func (s *userService) List(ctx context.Context, filter repository.UserFilter) ([]model.User, error) {
	return s.userRepo.List(ctx, filter)
}

func (s *userService) Create(ctx context.Context, input CreateUserInput) (*model.User, error) {
	password, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: input.Username,
		Nickname: input.Nickname,
		Email:    input.Email,
		Phone:    input.Phone,
		Status:   input.Status,
		Password: password,
	}
	if user.Status == "" {
		user.Status = "enabled"
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.userRepo.ReplaceRoles(ctx, user.ID, input.RoleIDs); err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, user.ID)
}

func (s *userService) Update(ctx context.Context, id uint, input UpdateUserInput) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Nickname = input.Nickname
	user.Email = input.Email
	user.Phone = input.Phone
	user.Status = input.Status

	if input.Password != "" {
		password, err := utils.HashPassword(input.Password)
		if err != nil {
			return nil, err
		}
		user.Password = password
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	if err := s.userRepo.ReplaceRoles(ctx, user.ID, input.RoleIDs); err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, user.ID)
}

func (s *userService) Delete(ctx context.Context, id uint) error {
	return s.userRepo.Delete(ctx, id)
}

