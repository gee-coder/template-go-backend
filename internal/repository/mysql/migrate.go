package mysql

import (
	"context"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"gorm.io/gorm"
)

// AutoMigrate creates the required tables.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Menu{},
		&model.AuthSetting{},
		&model.UserRole{},
		&model.RoleMenu{},
		&model.ContactSubmission{},
	)
}

// SeedInitialData creates the default role, menus and admin user.
func SeedInitialData(ctx context.Context, userRepo repository.UserRepository, roleRepo repository.RoleRepository, menuRepo repository.MenuRepository) error {
	allMenus, err := menuRepo.List(ctx, repository.MenuFilter{})
	if err != nil {
		return err
	}

	findMenu := func(name string, parentID uint) *model.Menu {
		for idx := range allMenus {
			if allMenus[idx].Name == name && allMenus[idx].ParentID == parentID {
				return &allMenus[idx]
			}
		}
		return nil
	}

	ensureMenu := func(target *model.Menu) (*model.Menu, error) {
		existing := findMenu(target.Name, target.ParentID)
		if existing != nil {
			existing.Title = target.Title
			existing.Path = target.Path
			existing.Component = target.Component
			existing.Icon = target.Icon
			existing.Type = target.Type
			existing.Permission = target.Permission
			existing.Sort = target.Sort
			existing.Hidden = target.Hidden
			existing.Status = target.Status
			if err := menuRepo.Update(ctx, existing); err != nil {
				return nil, err
			}
			return existing, nil
		}

		if err := menuRepo.Create(ctx, target); err != nil {
			return nil, err
		}
		allMenus = append(allMenus, *target)
		return target, nil
	}

	dashboardMenu, err := ensureMenu(&model.Menu{
		Name:       "dashboard",
		Title:      "工作台",
		Path:       "/dashboard",
		Component:  "views/dashboard/index.vue",
		Type:       "menu",
		Icon:       "House",
		Permission: "dashboard:view",
		Sort:       1,
		Status:     "enabled",
	})
	if err != nil {
		return err
	}

	systemMenu, err := ensureMenu(&model.Menu{
		Name:      "system",
		Title:     "系统管理",
		Path:      "/system",
		Component: "Layout",
		Type:      "directory",
		Icon:      "Setting",
		Sort:      2,
		Status:    "enabled",
	})
	if err != nil {
		return err
	}

	seedMenus := []*model.Menu{
		{ParentID: systemMenu.ID, Name: "user", Title: "用户管理", Path: "/system/users", Component: "views/system/users/index.vue", Type: "menu", Icon: "User", Permission: "system:user:view", Sort: 1, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "user_write", Title: "用户写入", Type: "button", Permission: "system:user:write", Sort: 2, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "role", Title: "角色管理", Path: "/system/roles", Component: "views/system/roles/index.vue", Type: "menu", Icon: "Lock", Permission: "system:role:view", Sort: 3, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "role_write", Title: "角色写入", Type: "button", Permission: "system:role:write", Sort: 4, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "menu", Title: "菜单管理", Path: "/system/menus", Component: "views/system/menus/index.vue", Type: "menu", Icon: "Menu", Permission: "system:menu:view", Sort: 5, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "menu_write", Title: "菜单写入", Type: "button", Permission: "system:menu:write", Sort: 6, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "auth_setting", Title: "认证设置", Path: "/system/auth-settings", Component: "views/system/auth-settings/index.vue", Type: "menu", Icon: "Setting", Permission: "system:auth-setting:view", Sort: 7, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "auth_setting_write", Title: "认证设置写入", Type: "button", Permission: "system:auth-setting:write", Sort: 8, Status: "enabled"},
	}

	requiredMenus := []model.Menu{*dashboardMenu, *systemMenu}
	for _, item := range seedMenus {
		menu, ensureErr := ensureMenu(item)
		if ensureErr != nil {
			return ensureErr
		}
		requiredMenus = append(requiredMenus, *menu)
	}

	role, err := roleRepo.GetByCode(ctx, "super_admin")
	if err != nil {
		if err != utils.ErrNotFound {
			return err
		}
		role = &model.Role{
			Name:   "超级管理员",
			Code:   "super_admin",
			Status: "enabled",
			Remark: "模板初始化角色",
		}
		if err := roleRepo.Create(ctx, role); err != nil {
			return err
		}
		role, err = roleRepo.GetByCode(ctx, "super_admin")
		if err != nil {
			return err
		}
	}

	menuIDSet := map[uint]struct{}{}
	for _, item := range role.Menus {
		menuIDSet[item.ID] = struct{}{}
	}
	for _, item := range requiredMenus {
		menuIDSet[item.ID] = struct{}{}
	}
	menuIDs := make([]uint, 0, len(menuIDSet))
	for id := range menuIDSet {
		menuIDs = append(menuIDs, id)
	}
	if err := roleRepo.ReplaceMenus(ctx, role.ID, menuIDs); err != nil {
		return err
	}

	password, err := utils.HashPassword("Admin123!")
	if err != nil {
		return err
	}

	user, err := userRepo.GetByUsername(ctx, "admin")
	if err != nil {
		if err != utils.ErrNotFound {
			return err
		}
		user = &model.User{
			Username: "admin",
			Nickname: "系统管理员",
			Email:    "admin@example.com",
			Phone:    "18800000000",
			Status:   "enabled",
			Password: password,
		}
		if err := userRepo.Create(ctx, user); err != nil {
			return err
		}
		user, err = userRepo.GetByUsername(ctx, "admin")
		if err != nil {
			return err
		}
	}

	roleIDSet := map[uint]struct{}{}
	for _, item := range user.Roles {
		roleIDSet[item.ID] = struct{}{}
	}
	roleIDSet[role.ID] = struct{}{}
	roleIDs := make([]uint, 0, len(roleIDSet))
	for id := range roleIDSet {
		roleIDs = append(roleIDs, id)
	}

	return userRepo.ReplaceRoles(ctx, user.ID, roleIDs)
}
