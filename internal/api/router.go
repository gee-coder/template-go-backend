package api

import (
	"net/http"

	"github.com/gee-coder/template-go-backend/internal/api/handler"
	"github.com/gee-coder/template-go-backend/internal/api/middleware"
	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HandlerSet groups all route handlers.
type HandlerSet struct {
	Health          *handler.HealthHandler
	Auth            *handler.AuthHandler
	AuthSetting     *handler.AuthSettingHandler
	BrandingSetting *handler.BrandingSettingHandler
	LoginAudit      *handler.LoginAuditHandler
	User            *handler.UserHandler
	Role            *handler.RoleHandler
	Menu            *handler.MenuHandler
	Contact         *handler.ContactHandler
}

// NewHandlerSet creates all handlers.
func NewHandlerSet(
	auth handler.AuthService,
	avatarAsset handler.AvatarAssetService,
	authSetting handler.AuthSettingService,
	brandingSetting handler.BrandingSettingService,
	loginAudit handler.LoginAuditService,
	user handler.UserService,
	role handler.RoleService,
	menu handler.MenuService,
	contact handler.ContactService,
) *HandlerSet {
	return &HandlerSet{
		Health:          handler.NewHealthHandler(),
		Auth:            handler.NewAuthHandler(auth, loginAudit, avatarAsset),
		AuthSetting:     handler.NewAuthSettingHandler(authSetting),
		BrandingSetting: handler.NewBrandingSettingHandler(brandingSetting),
		LoginAudit:      handler.NewLoginAuditHandler(loginAudit),
		User:            handler.NewUserHandler(user),
		Role:            handler.NewRoleHandler(role),
		Menu:            handler.NewMenuHandler(menu),
		Contact:         handler.NewContactHandler(contact),
	}
}

// NewRouter creates the gin router.
func NewRouter(cfg *config.Config, logger *zap.Logger, handlers *HandlerSet) *gin.Engine {
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.AccessLog(logger))
	router.Use(middleware.CORS(cfg.HTTP.AllowedOrigins))

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "后端服务运行中"})
	})
	router.Static("/docs", "./docs")
	router.Static("/uploads", cfg.App.UploadPath())

	apiV1 := router.Group("/api/v1")
	apiV1.GET("/healthz", handlers.Health.Check)
	apiV1.POST("/public/contact-submissions", handlers.Contact.Create)
	apiV1.GET("/auth/options", handlers.Auth.Options)
	apiV1.GET("/branding/settings", handlers.BrandingSetting.GetPublic)
	apiV1.POST("/auth/login", handlers.Auth.Login)
	apiV1.POST("/auth/register", handlers.Auth.Register)
	apiV1.POST("/auth/refresh", handlers.Auth.Refresh)

	authenticated := apiV1.Group("")
	authenticated.Use(middleware.JWTAuth(cfg.JWT.Secret))
	authenticated.GET("/auth/profile", handlers.Auth.Profile)
	authenticated.PUT("/auth/profile", handlers.Auth.UpdateProfile)
	authenticated.POST("/auth/avatar-assets", handlers.Auth.UploadAvatarAsset)
	authenticated.POST("/auth/logout", handlers.Auth.Logout)

	system := authenticated.Group("/system")
	system.GET("/users", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:user:view"), handlers.User.List)
	system.POST("/users", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:user:write"), handlers.User.Create)
	system.PUT("/users/:id", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:user:write"), handlers.User.Update)
	system.DELETE("/users/:id", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:user:write"), handlers.User.Delete)

	system.GET("/roles", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:role:view"), handlers.Role.List)
	system.POST("/roles", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:role:write"), handlers.Role.Create)
	system.PUT("/roles/:id", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:role:write"), handlers.Role.Update)
	system.DELETE("/roles/:id", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:role:write"), handlers.Role.Delete)

	system.GET("/menus", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:menu:view"), handlers.Menu.List)
	system.POST("/menus", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:menu:write"), handlers.Menu.Create)
	system.PUT("/menus/:id", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:menu:write"), handlers.Menu.Update)
	system.DELETE("/menus/:id", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:menu:write"), handlers.Menu.Delete)

	system.GET("/auth-settings", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:auth-setting:view"), handlers.AuthSetting.Get)
	system.PUT("/auth-settings", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:auth-setting:write"), handlers.AuthSetting.Update)
	system.GET("/branding-settings", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:branding-setting:view"), handlers.BrandingSetting.Get)
	system.PUT("/branding-settings", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:branding-setting:write"), handlers.BrandingSetting.Update)
	system.POST("/branding-settings/assets", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:branding-setting:write"), handlers.BrandingSetting.UploadAsset)
	system.GET("/login-audits", middleware.PermissionGuard(handlers.Auth.ResolvePermissions, "system:login-audit:view"), handlers.LoginAudit.List)

	return router
}
