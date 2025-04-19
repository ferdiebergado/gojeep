package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/infra/db"
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/environment"
	"github.com/ferdiebergado/gojeep/internal/pkg/logging"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/pkg/validation"
	"github.com/ferdiebergado/gojeep/internal/router"
	"github.com/ferdiebergado/gojeep/internal/server"
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

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		return err
	}

	dbConn, err := db.Connect(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	deps, err := setupDependencies(cfg, dbConn)
	if err != nil {
		return err
	}

	app := newApp(deps)
	app.SetupRoutes()

	apiServer := server.New(cfg.Server, app.Router())
	apiServerErr := apiServer.Start()
	select {
	case <-ctx.Done():
		slog.Info("Shutdown signal received.")
	case err := <-apiServerErr:
		return fmt.Errorf("server error: %w", err)
	}

	return apiServer.Shutdown()
}

func setupDependencies(cfg *config.Config, db *sql.DB) (*dependencies, error) {
	httpRouter := router.New()
	validate = validation.New()
	hasher := &security.Argon2Hasher{}
	mailer, err := email.New(cfg.Email)
	if err != nil {
		return nil, err
	}
	signer := security.NewSigner(cfg)

	deps := &dependencies{
		Config:    cfg,
		DB:        db,
		Router:    httpRouter,
		Validator: validate,
		Hasher:    hasher,
		Mailer:    mailer,
		Signer:    signer,
	}
	return deps, nil
}
