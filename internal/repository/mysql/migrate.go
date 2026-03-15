package mysql

import (
	"context"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
	"gorm.io/gorm"
)

const (
	seedDashboardTitle = "Dashboard"
	seedSystemTitle    = "System"

	seedRoleName   = "Super Admin"
	seedRoleRemark = "Template default role"

	seedAdminUsername = "admin"
	seedAdminNickname = "System Admin"
	seedAdminEmail    = "admin@example.com"
	seedAdminPhone    = "18800000000"
	seedAdminAvatar   = "default-07"
)

var (
	legacySuperAdminNames   = []string{seedRoleName, "\u8d85\u7ea7\u7ba1\u7406\u5458"}
	legacySuperAdminRemarks = []string{seedRoleRemark, "\u6a21\u677f\u521d\u59cb\u5316\u89d2\u8272"}
	legacyAdminNicknames    = []string{seedAdminNickname, "\u7cfb\u7edf\u7ba1\u7406\u5458"}
)

// AutoMigrate creates the required tables.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Menu{},
		&model.AuthSetting{},
		&model.BrandingSetting{},
		&model.LoginAuditLog{},
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
		Title:      seedDashboardTitle,
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
		Title:     seedSystemTitle,
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
		{ParentID: systemMenu.ID, Name: "user", Title: "Users", Path: "/system/users", Component: "views/system/users/index.vue", Type: "menu", Icon: "User", Permission: "system:user:view", Sort: 1, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "user_write", Title: "User Write", Type: "button", Permission: "system:user:write", Sort: 2, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "role", Title: "Roles", Path: "/system/roles", Component: "views/system/roles/index.vue", Type: "menu", Icon: "Lock", Permission: "system:role:view", Sort: 3, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "role_write", Title: "Role Write", Type: "button", Permission: "system:role:write", Sort: 4, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "menu", Title: "Menus", Path: "/system/menus", Component: "views/system/menus/index.vue", Type: "menu", Icon: "Menu", Permission: "system:menu:view", Sort: 5, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "menu_write", Title: "Menu Write", Type: "button", Permission: "system:menu:write", Sort: 6, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "auth_setting", Title: "Auth Settings", Path: "/system/auth-settings", Component: "views/system/auth-settings/index.vue", Type: "menu", Icon: "Setting", Permission: "system:auth-setting:view", Sort: 7, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "auth_setting_write", Title: "Auth Settings Write", Type: "button", Permission: "system:auth-setting:write", Sort: 8, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "branding_setting", Title: "Branding", Path: "/system/branding-settings", Component: "views/system/branding-settings/index.vue", Type: "menu", Icon: "Grid", Permission: "system:branding-setting:view", Sort: 9, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "branding_setting_write", Title: "Branding Write", Type: "button", Permission: "system:branding-setting:write", Sort: 10, Status: "enabled"},
		{ParentID: systemMenu.ID, Name: "login_audit", Title: "Login Audits", Path: "/system/login-audits", Component: "views/system/login-audits/index.vue", Type: "menu", Icon: "Document", Permission: "system:login-audit:view", Sort: 11, Status: "enabled"},
	}

	requiredMenus := []model.Menu{*dashboardMenu, *systemMenu}
	for _, item := range seedMenus {
		menu, ensureErr := ensureMenu(item)
		if ensureErr != nil {
			return ensureErr
		}
		requiredMenus = append(requiredMenus, *menu)
	}

	role, err := ensureSeedRole(ctx, roleRepo)
	if err != nil {
		return err
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

	user, err := ensureSeedAdminUser(ctx, userRepo, password)
	if err != nil {
		return err
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

func ensureSeedRole(ctx context.Context, roleRepo repository.RoleRepository) (*model.Role, error) {
	role, err := roleRepo.GetByCode(ctx, "super_admin")
	if err != nil {
		if err != utils.ErrNotFound {
			return nil, err
		}

		role = &model.Role{
			Name:   seedRoleName,
			Code:   "super_admin",
			Status: "enabled",
			Remark: seedRoleRemark,
		}
		if err := roleRepo.Create(ctx, role); err != nil {
			return nil, err
		}
		return roleRepo.GetByCode(ctx, "super_admin")
	}

	roleUpdate := *role
	roleUpdate.Menus = nil
	changed := false

	if strings.TrimSpace(roleUpdate.Name) == "" || matchesTemplateValue(roleUpdate.Name, legacySuperAdminNames...) {
		roleUpdate.Name = seedRoleName
		changed = true
	}
	if strings.TrimSpace(roleUpdate.Remark) == "" || matchesTemplateValue(roleUpdate.Remark, legacySuperAdminRemarks...) {
		roleUpdate.Remark = seedRoleRemark
		changed = true
	}
	if strings.TrimSpace(roleUpdate.Status) == "" {
		roleUpdate.Status = "enabled"
		changed = true
	}

	if changed {
		if err := roleRepo.Update(ctx, &roleUpdate); err != nil {
			return nil, err
		}
		return roleRepo.GetByCode(ctx, "super_admin")
	}

	return role, nil
}

func ensureSeedAdminUser(ctx context.Context, userRepo repository.UserRepository, password string) (*model.User, error) {
	user, err := userRepo.GetByUsername(ctx, seedAdminUsername)
	if err != nil {
		if err != utils.ErrNotFound {
			return nil, err
		}

		user = &model.User{
			Username: seedAdminUsername,
			Nickname: seedAdminNickname,
			Email:    seedAdminEmail,
			Phone:    seedAdminPhone,
			Avatar:   seedAdminAvatar,
			Status:   "enabled",
			Password: password,
		}
		if err := userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
		return userRepo.GetByUsername(ctx, seedAdminUsername)
	}

	userUpdate := *user
	userUpdate.Roles = nil
	changed := false

	if strings.TrimSpace(userUpdate.Nickname) == "" || matchesTemplateValue(userUpdate.Nickname, legacyAdminNicknames...) {
		userUpdate.Nickname = seedAdminNickname
		changed = true
	}
	if strings.TrimSpace(userUpdate.Email) == "" {
		userUpdate.Email = seedAdminEmail
		changed = true
	}
	if strings.TrimSpace(userUpdate.Phone) == "" {
		userUpdate.Phone = seedAdminPhone
		changed = true
	}
	if strings.TrimSpace(userUpdate.Avatar) == "" {
		userUpdate.Avatar = seedAdminAvatar
		changed = true
	}
	if strings.TrimSpace(userUpdate.Status) == "" {
		userUpdate.Status = "enabled"
		changed = true
	}

	if changed {
		if err := userRepo.Update(ctx, &userUpdate); err != nil {
			return nil, err
		}
		return userRepo.GetByUsername(ctx, seedAdminUsername)
	}

	return user, nil
}

func matchesTemplateValue(value string, candidates ...string) bool {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return false
	}

	for _, candidate := range candidates {
		if normalized == strings.TrimSpace(candidate) {
			return true
		}
	}
	return false
}
