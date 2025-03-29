package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/gopherkit/env"
	"github.com/ferdiebergado/goweb/internal/config"
	"github.com/ferdiebergado/goweb/internal/handler"
	"github.com/ferdiebergado/goweb/internal/infra/db"
	"github.com/ferdiebergado/goweb/internal/pkg/environment"
	"github.com/ferdiebergado/goweb/internal/pkg/logging"
	"github.com/ferdiebergado/goweb/internal/pkg/security"
	"github.com/ferdiebergado/goweb/internal/pkg/validation"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	envVar  = "ENV"
	envDev  = "development"
	envProd = "production"
	cfgFile = "config.json"
	fmtAddr = ":%d"
)

var validate *validator.Validate

func main() {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer func() {
		stop()
		slog.Info("Signal context cleanup complete.")
	}()

	if err := run(signalCtx); err != nil {
		slog.Error("fatal error", "reason", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	appEnv, err := setupEnvironment()
	if err != nil {
		return err
	}

	logging.SetLogger(os.Stdout, appEnv)

	cfg, err := loadConfiguration()
	if err != nil {
		return err
	}

	dbConn, err := db.Connect(ctx, &cfg.Db)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	deps, err := setupDependencies(cfg, dbConn)
	if err != nil {
		return err
	}

	app := handler.NewApp(deps)
	app.SetupRoutes()

	server := createServer(cfg, app.Router())

	serverErr := startServer(server, cfg)
	select {
	case <-ctx.Done():
		slog.Info("Shutdown signal received.")
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}

	return shutdownServer(server, cfg)
}

func setupEnvironment() (string, error) {
	appEnv := env.Get(envVar, envDev)
	if appEnv != envProd {
		if err := environment.LoadEnv(appEnv); err != nil {
			return "", fmt.Errorf("load env: %w", err)
		}
	}
	return appEnv, nil
}

func loadConfiguration() (*config.Config, error) {
	cf := flag.String("cfg", cfgFile, "Config file")
	flag.Parse()
	cfg, err := config.LoadConfig(*cf)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

func setupDependencies(cfg *config.Config, db *sql.DB) (*handler.AppDependencies, error) {
	router := goexpress.New()
	validate = validation.New()
	tmpl, err := handler.NewTemplate(cfg.Template)
	if err != nil {
		return nil, err
	}
	hasher := &security.Argon2Hasher{}

	deps := &handler.AppDependencies{
		Config:    cfg,
		DB:        db,
		Router:    router,
		Validator: validate,
		Template:  tmpl,
		Hasher:    hasher,
	}
	return deps, nil
}

func createServer(cfg *config.Config, router *goexpress.Router) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(fmtAddr, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
}

func startServer(server *http.Server, cfg *config.Config) chan error {
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Server started", "address", server.Addr, "env", cfg.App.Env, slog.Bool("debug", cfg.App.IsDebug))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()
	return serverErr
}

func shutdownServer(server *http.Server, cfg *config.Config) error {
	slog.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("Server gracefully shut down.")
	return nil
}
