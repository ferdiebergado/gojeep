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

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/infra/db"
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/environment"
	"github.com/ferdiebergado/gojeep/internal/pkg/logging"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/pkg/validation"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var validate *validator.Validate

func main() {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(signalCtx); err != nil {
		slog.Error("fatal error", "reason", err)
		stop()
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfgFile := flag.String("cfg", "config.json", "Config file")
	logLevel := flag.String("loglevel", "", "Log level (info/warn/error/debug)")
	flag.Parse()

	appEnv, err := environment.Setup()
	if err != nil {
		return err
	}

	logging.SetLogger(os.Stdout, appEnv, *logLevel)

	handler.StartPProf()

	cfg, err := config.New(*cfgFile)
	if err != nil {
		return err
	}

	dbConn, err := db.Connect(ctx, cfg)
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

func setupDependencies(cfg *config.Config, db *sql.DB) (*handler.AppDependencies, error) {
	router := handler.NewRouter()
	validate = validation.New()
	hasher := &security.Argon2Hasher{}
	mailer, err := email.New(cfg)
	if err != nil {
		return nil, err
	}
	signer := security.NewSigner(cfg)

	deps := &handler.AppDependencies{
		Config:    cfg,
		DB:        db,
		Router:    router,
		Validator: validate,
		Hasher:    hasher,
		Mailer:    mailer,
		Signer:    signer,
	}
	return deps, nil
}

func createServer(cfg *config.Config, router http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Options.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Options.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Options.Server.IdleTimeout) * time.Second,
	}
}

func startServer(server *http.Server, cfg *config.Config) chan error {
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Server started", "address", server.Addr, "env", cfg.App.Env, "log_level", cfg.App.LogLevel)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()
	return serverErr
}

func shutdownServer(server *http.Server, cfg *config.Config) error {
	slog.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Options.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("Server gracefully shut down.")
	return nil
}
