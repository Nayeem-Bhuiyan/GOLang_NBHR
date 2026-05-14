package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"nbhr/config"
	"nbhr/internal/modules/auth"
	"nbhr/internal/modules/permission"
	"nbhr/internal/modules/role"
	"nbhr/internal/modules/user"
	"nbhr/internal/router"
	"nbhr/internal/shared/jwt"
	"nbhr/internal/shared/logger"
)

// App encapsulates the entire application.
type App struct {
	cfg    *config.Config
	db     *gorm.DB
	log    *zap.Logger
	server *http.Server
}

// New bootstraps the application with all dependencies wired.
func New(cfg *config.Config) (*App, error) {
	log := logger.New(cfg.App.Env)

	db, err := initDatabase(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("database init failed: %w", err)
	}

	// JWT manager
	jwtManager := jwt.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)

	// Repositories
	userRepo := user.NewRepository(db)
	roleRepo := role.NewRepository(db)
	permRepo := permission.NewRepository(db)
	tokenRepo := auth.NewTokenRepository(db)

	// Services
	authSvc := auth.NewService(userRepo, tokenRepo, jwtManager)
	userSvc := user.NewService(userRepo, roleRepo)
	roleSvc := role.NewService(roleRepo, permRepo)
	permSvc := permission.NewService(permRepo)

	// Handlers
	authHandler := auth.NewHandler(authSvc)
	userHandler := user.NewHandler(userSvc)
	roleHandler := role.NewHandler(roleSvc)
	permHandler := permission.NewHandler(permSvc)

	// Router
	engine := router.Setup(&router.Dependencies{
		AuthHandler:       authHandler,
		UserHandler:       userHandler,
		RoleHandler:       roleHandler,
		PermissionHandler: permHandler,
		JWTManager:        jwtManager,
		Config:            cfg,
		Logger:            log,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		cfg:    cfg,
		db:     db,
		log:    log,
		server: srv,
	}, nil
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
func (a *App) Run() error {
	errCh := make(chan error, 1)

	go func() {
		a.log.Info("server starting",
			zap.String("address", a.server.Addr),
			zap.String("env", a.cfg.App.Env),
			zap.String("version", a.cfg.App.Version),
		)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		a.log.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	return a.shutdown()
}

// shutdown gracefully stops the server and closes DB connections.
func (a *App) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	a.log.Info("shutting down server gracefully...")

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("server forced shutdown", zap.Error(err))
		return err
	}

	sqlDB, err := a.db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			a.log.Error("failed to close database", zap.Error(err))
		}
	}

	a.log.Info("server shutdown complete")
	return nil
}

func initDatabase(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	logLevel := gormlogger.Silent
	if cfg.App.Debug {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger:                 gormlogger.Default.LogMode(logLevel),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Info("database connected successfully",
		zap.String("host", cfg.Database.Host),
		zap.String("name", cfg.Database.Name),
	)

	return db, nil
}