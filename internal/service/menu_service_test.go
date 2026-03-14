package service

import (
	"context"
	"testing"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
)

func TestBuildMenuTree(t *testing.T) {
	repo := &fakeMenuRepository{
		menus: []model.Menu{
			{BaseModel: model.BaseModel{ID: 1}, Name: "system", Title: "系统", ParentID: 0},
			{BaseModel: model.BaseModel{ID: 2}, Name: "user", Title: "用户", ParentID: 1},
		},
	}

	svc := NewMenuService(repo)
	menus, err := svc.List(context.Background(), repository.MenuFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(menus) != 1 {
		t.Fatalf("expected 1 root menu, got %d", len(menus))
	}
	if len(menus[0].Children) != 1 {
		t.Fatalf("expected 1 child menu, got %d", len(menus[0].Children))
	}
}

type fakeMenuRepository struct {
	menus []model.Menu
}

func (f *fakeMenuRepository) GetByID(ctx context.Context, id uint) (*model.Menu, error) {
	for _, menu := range f.menus {
		if menu.ID == id {
			return &menu, nil
		}
	}
	return nil, nil
}

func (f *fakeMenuRepository) List(ctx context.Context, filter repository.MenuFilter) ([]model.Menu, error) {
	return f.menus, nil
}

func (f *fakeMenuRepository) Create(ctx context.Context, menu *model.Menu) error { return nil }
func (f *fakeMenuRepository) Update(ctx context.Context, menu *model.Menu) error { return nil }
func (f *fakeMenuRepository) Delete(ctx context.Context, id uint) error          { return nil }
