package app

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/infra/db"
	"github.com/ferdiebergado/gojeep/internal/pkg/environment"
	"github.com/ferdiebergado/gojeep/internal/pkg/logging"
	"github.com/ferdiebergado/gojeep/internal/pkg/validation"
	"github.com/ferdiebergado/gojeep/internal/server"
)

func Run(ctx context.Context) error {
	slog.Info("Initializing...")

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

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

	dbConn, err := db.Connect(signalCtx, cfg.DB)
	if err != nil {
		return err
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			slog.Error("failed to close the database", "reason", err)
			return
		}
	}()

	validate := validation.New()
	deps, err := newDependencies(cfg, dbConn, validate)
	if err != nil {
		return err
	}

	application := New(deps)
	application.SetupRoutes()

	apiServer := server.New(signalCtx, cfg.Server, application.Router())
	apiServerErr := apiServer.Start()

	select {
	case <-signalCtx.Done():
		slog.Info("Shutdown signal received.")
	case err := <-apiServerErr:
		return fmt.Errorf("server error: %w", err)
	}

	return apiServer.Shutdown(ctx)
}
