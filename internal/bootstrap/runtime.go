package bootstrap

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
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Runtime wires config, logger, handlers, and shared infrastructure.
type Runtime struct {
	Config   *config.Config
	Logger   *zap.Logger
	Handlers *api.HandlerSet
}

// NewRuntime builds the shared runtime for an HTTP service.
func NewRuntime(runDatabaseBootstrap bool) (*Runtime, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger, err := newLogger(cfg.App.Debug)
	if err != nil {
		return nil, err
	}

	db, err := repository.NewMySQL(cfg.Database)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}

	tokenStore, err := repository.NewRedisTokenStore(cfg.Redis)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}
	cacheStore, err := repository.NewRedisCacheStore(cfg.Redis)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}

	if cfg.Storage.ResolvedProvider() == "local" {
		if err := os.MkdirAll(cfg.App.UploadPath(), 0o755); err != nil {
			_ = logger.Sync()
			return nil, err
		}
	}

	objectStorage, err := service.NewObjectStorage(cfg)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}
	smsVerificationService, err := service.NewSMSVerificationService(cfg.SMS, cacheStore, logger)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}
	emailVerificationService, err := service.NewEmailVerificationService(cfg.Mail, cacheStore, logger)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}
	imageCaptchaService := service.NewImageCaptchaService(cacheStore)

	userRepo := mysql.NewUserRepository(db)
	roleRepo := mysql.NewRoleRepository(db)
	menuRepo := mysql.NewMenuRepository(db)
	authSettingRepo := mysql.NewAuthSettingRepository(db)
	brandingSettingRepo := mysql.NewBrandingSettingRepository(db)
	loginAuditRepo := mysql.NewLoginAuditRepository(db)
	contactRepo := mysql.NewContactSubmissionRepository(db)

	if runDatabaseBootstrap {
		if err := mysql.AutoMigrate(db); err != nil {
			_ = logger.Sync()
			return nil, err
		}
		if err := mysql.SeedInitialData(context.Background(), userRepo, roleRepo, menuRepo); err != nil {
			_ = logger.Sync()
			return nil, err
		}
	}

	avatarAssetService := service.NewAvatarAssetService(objectStorage)
	authService := service.NewAuthService(cfg.JWT, cfg.Auth, authSettingRepo, userRepo, tokenStore, cacheStore, smsVerificationService, emailVerificationService, imageCaptchaService, objectStorage.SupportsPublicURL)
	authSettingService := service.NewAuthSettingService(cfg.Auth, authSettingRepo, cacheStore)
	brandingSettingService := service.NewBrandingSettingService(brandingSettingRepo, objectStorage, cacheStore)
	loginAuditService := service.NewLoginAuditService(loginAuditRepo)
	userService := service.NewUserService(userRepo, roleRepo, cacheStore)
	roleService := service.NewRoleService(roleRepo, menuRepo, cacheStore)
	menuService := service.NewMenuService(menuRepo, cacheStore)
	contactService := service.NewContactService(contactRepo)

	return &Runtime{
		Config: cfg,
		Logger: logger,
		Handlers: api.NewHandlerSet(
			authService,
			avatarAssetService,
			smsVerificationService,
			emailVerificationService,
			imageCaptchaService,
			authSettingService,
			brandingSettingService,
			loginAuditService,
			userService,
			roleService,
			menuService,
			contactService,
		),
	}, nil
}

// RunHTTP starts an HTTP service and blocks until shutdown.
func RunHTTP(runDatabaseBootstrap bool, buildHandler func(*config.Config, *zap.Logger, *api.HandlerSet) *gin.Engine) error {
	runtime, err := NewRuntime(runDatabaseBootstrap)
	if err != nil {
		return err
	}
	defer runtime.Close()

	return serveHTTP(runtime.Config, runtime.Logger, buildHandler(runtime.Config, runtime.Logger, runtime.Handlers))
}

// RunDatabaseBootstrap runs migrations and seed data once, then exits.
func RunDatabaseBootstrap() error {
	runtime, err := NewRuntime(true)
	if err != nil {
		return err
	}
	defer runtime.Close()

	runtime.Logger.Info("database bootstrap completed")
	return nil
}

// Close releases runtime resources.
func (r *Runtime) Close() {
	if r == nil || r.Logger == nil {
		return
	}
	_ = r.Logger.Sync()
}

func newLogger(debug bool) (*zap.Logger, error) {
	if debug {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

func serveHTTP(cfg *config.Config, logger *zap.Logger, handler http.Handler) error {
	server := &http.Server{
		Addr:         cfg.HTTP.Address(),
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	go func() {
		logger.Info("http server listening", zap.String("addr", cfg.HTTP.Address()))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("listen and serve", zap.Error(err))
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown server", zap.Error(err))
		return err
	}

	return nil
}
