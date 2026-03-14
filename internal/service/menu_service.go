package service

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
)

// MenuNode describes the menu tree node response.
type MenuNode struct {
	model.Menu
	Children []MenuNode `json:"children"`
}

// MenuService provides menu management capabilities.
type MenuService interface {
	List(ctx context.Context, filter repository.MenuFilter) ([]MenuNode, error)
	Create(ctx context.Context, input CreateMenuInput) (*model.Menu, error)
	Update(ctx context.Context, id uint, input UpdateMenuInput) (*model.Menu, error)
	Delete(ctx context.Context, id uint) error
}

// CreateMenuInput is the input of creating a menu.
type CreateMenuInput struct {
	ParentID   uint
	Name       string
	Title      string
	Path       string
	Component  string
	Icon       string
	Type       string
	Permission string
	Sort       int
	Hidden     bool
	Status     string
}

// UpdateMenuInput is the input of updating a menu.
type UpdateMenuInput = CreateMenuInput

type menuService struct {
	menuRepo repository.MenuRepository
}

// NewMenuService creates the menu service.
func NewMenuService(menuRepo repository.MenuRepository) MenuService {
	return &menuService{menuRepo: menuRepo}
}

func (s *menuService) List(ctx context.Context, filter repository.MenuFilter) ([]MenuNode, error) {
	menus, err := s.menuRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus), nil
}

func (s *menuService) Create(ctx context.Context, input CreateMenuInput) (*model.Menu, error) {
	menu := &model.Menu{
		ParentID:   input.ParentID,
		Name:       input.Name,
		Title:      input.Title,
		Path:       input.Path,
		Component:  input.Component,
		Icon:       input.Icon,
		Type:       input.Type,
		Permission: input.Permission,
		Sort:       input.Sort,
		Hidden:     input.Hidden,
		Status:     input.Status,
	}
	if menu.Type == "" {
		menu.Type = "menu"
	}
	if menu.Status == "" {
		menu.Status = "enabled"
	}
	if err := s.menuRepo.Create(ctx, menu); err != nil {
		return nil, err
	}
	return menu, nil
}

func (s *menuService) Update(ctx context.Context, id uint, input UpdateMenuInput) (*model.Menu, error) {
	menu, err := s.menuRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	menu.ParentID = input.ParentID
	menu.Name = input.Name
	menu.Title = input.Title
	menu.Path = input.Path
	menu.Component = input.Component
	menu.Icon = input.Icon
	menu.Type = input.Type
	menu.Permission = input.Permission
	menu.Sort = input.Sort
	menu.Hidden = input.Hidden
	menu.Status = input.Status

	if err := s.menuRepo.Update(ctx, menu); err != nil {
		return nil, err
	}
	return menu, nil
}

func (s *menuService) Delete(ctx context.Context, id uint) error {
	return s.menuRepo.Delete(ctx, id)
}

func buildMenuTree(menus []model.Menu) []MenuNode {
	nodes := make(map[uint]*MenuNode, len(menus))
	rootPointers := make([]*MenuNode, 0)

	for _, menu := range menus {
		item := MenuNode{
			Menu:     menu,
			Children: []MenuNode{},
		}
		nodes[menu.ID] = &item
	}

	for _, menu := range menus {
		node := nodes[menu.ID]
		if menu.ParentID == 0 {
			rootPointers = append(rootPointers, node)
			continue
		}

		parent, exists := nodes[menu.ParentID]
		if !exists {
			rootPointers = append(rootPointers, node)
			continue
		}
		parent.Children = append(parent.Children, *node)
	}

	roots := make([]MenuNode, 0, len(rootPointers))
	for _, item := range rootPointers {
		roots = append(roots, *item)
	}

	return roots
}
