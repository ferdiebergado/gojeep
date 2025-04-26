package app

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/infra/db"
	"github.com/ferdiebergado/gojeep/internal/pkg/environment"
	"github.com/ferdiebergado/gojeep/internal/pkg/logging"
	"github.com/ferdiebergado/gojeep/internal/pkg/validation"
	"github.com/ferdiebergado/gojeep/internal/server"
)

func Run(ctx context.Context) error {
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

	validate := validation.New()
	deps, err := newDependencies(cfg, dbConn, validate)
	if err != nil {
		return err
	}

	application := New(deps)
	application.SetupRoutes()

	apiServer := server.New(cfg.Server, application.Router())
	apiServerErr := apiServer.Start()
	select {
	case <-ctx.Done():
		slog.Info("Shutdown signal received.")
	case err := <-apiServerErr:
		return fmt.Errorf("server error: %w", err)
	}

	return apiServer.Shutdown()
}
