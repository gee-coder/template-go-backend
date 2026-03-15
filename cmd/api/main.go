package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gee-coder/template-go-backend/internal/api"
	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/mysql"
	"github.com/gee-coder/template-go-backend/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewProduction()
	if cfg.App.Debug {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	db, err := repository.NewMySQL(cfg.Database)
	if err != nil {
		logger.Fatal("connect mysql", zap.Error(err))
	}

	tokenStore, err := repository.NewRedisTokenStore(cfg.Redis)
	if err != nil {
		logger.Fatal("connect redis", zap.Error(err))
	}
	cacheStore, err := repository.NewRedisCacheStore(cfg.Redis)
	if err != nil {
		logger.Fatal("connect redis cache", zap.Error(err))
	}

	if err := mysql.AutoMigrate(db); err != nil {
		logger.Fatal("auto migrate", zap.Error(err))
	}

	if err := os.MkdirAll(cfg.App.UploadPath(), 0o755); err != nil {
		logger.Fatal("create upload dir", zap.Error(err))
	}

	userRepo := mysql.NewUserRepository(db)
	roleRepo := mysql.NewRoleRepository(db)
	menuRepo := mysql.NewMenuRepository(db)
	authSettingRepo := mysql.NewAuthSettingRepository(db)
	brandingSettingRepo := mysql.NewBrandingSettingRepository(db)
	loginAuditRepo := mysql.NewLoginAuditRepository(db)
	contactRepo := mysql.NewContactSubmissionRepository(db)

	if err := mysql.SeedInitialData(context.Background(), userRepo, roleRepo, menuRepo); err != nil {
		logger.Fatal("seed data", zap.Error(err))
	}

	avatarAssetService := service.NewAvatarAssetService(cfg.App.UploadPath())
	authService := service.NewAuthService(cfg.JWT, cfg.Auth, authSettingRepo, userRepo, tokenStore, cacheStore)
	authSettingService := service.NewAuthSettingService(cfg.Auth, authSettingRepo, cacheStore)
	brandingSettingService := service.NewBrandingSettingService(brandingSettingRepo, cfg.App.UploadPath(), cacheStore)
	loginAuditService := service.NewLoginAuditService(loginAuditRepo)
	userService := service.NewUserService(userRepo, roleRepo, cacheStore)
	roleService := service.NewRoleService(roleRepo, menuRepo, cacheStore)
	menuService := service.NewMenuService(menuRepo, cacheStore)
	contactService := service.NewContactService(contactRepo)

	handlerSet := api.NewHandlerSet(authService, avatarAssetService, authSettingService, brandingSettingService, loginAuditService, userService, roleService, menuService, contactService)
	router := api.NewRouter(cfg, logger, handlerSet)

	server := &http.Server{
		Addr:         cfg.HTTP.Address(),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		logger.Info("http server listening", zap.String("addr", cfg.HTTP.Address()))
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Fatal("listen and serve", zap.Error(serveErr))
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown server", zap.Error(err))
	}
}
