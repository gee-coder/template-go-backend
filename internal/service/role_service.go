package service

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
)

// RoleService provides role management capabilities.
type RoleService interface {
	List(ctx context.Context, filter repository.RoleFilter) ([]model.Role, error)
	Create(ctx context.Context, input CreateRoleInput) (*model.Role, error)
	Update(ctx context.Context, id uint, input UpdateRoleInput) (*model.Role, error)
	Delete(ctx context.Context, id uint) error
}

// CreateRoleInput is the input of creating a role.
type CreateRoleInput struct {
	Name    string
	Code    string
	Status  string
	Remark  string
	MenuIDs []uint
}

// UpdateRoleInput is the input of updating a role.
type UpdateRoleInput struct {
	Name    string
	Status  string
	Remark  string
	MenuIDs []uint
}

type roleService struct {
	roleRepo repository.RoleRepository
	menuRepo repository.MenuRepository
}

// NewRoleService creates the role service.
func NewRoleService(roleRepo repository.RoleRepository, menuRepo repository.MenuRepository) RoleService {
	return &roleService{roleRepo: roleRepo, menuRepo: menuRepo}
}

func (s *roleService) List(ctx context.Context, filter repository.RoleFilter) ([]model.Role, error) {
	return s.roleRepo.List(ctx, filter)
}

func (s *roleService) Create(ctx context.Context, input CreateRoleInput) (*model.Role, error) {
	role := &model.Role{
		Name:   input.Name,
		Code:   input.Code,
		Status: input.Status,
		Remark: input.Remark,
	}
	if role.Status == "" {
		role.Status = "enabled"
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}
	if err := s.roleRepo.ReplaceMenus(ctx, role.ID, input.MenuIDs); err != nil {
		return nil, err
	}
	return s.roleRepo.GetByID(ctx, role.ID)
}

func (s *roleService) Update(ctx context.Context, id uint, input UpdateRoleInput) (*model.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	role.Name = input.Name
	role.Status = input.Status
	role.Remark = input.Remark

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}
	if err := s.roleRepo.ReplaceMenus(ctx, role.ID, input.MenuIDs); err != nil {
		return nil, err
	}
	return s.roleRepo.GetByID(ctx, role.ID)
}

func (s *roleService) Delete(ctx context.Context, id uint) error {
	return s.roleRepo.Delete(ctx, id)
}

