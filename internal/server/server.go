package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
)

type Server struct {
	http.Server
	cfg *config.Config
}

func New(cfg *config.Config, app http.Handler) *Server {
	return &Server{
		cfg: cfg,
		Server: http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.App.Port),
			Handler:      app,
			ReadTimeout:  time.Duration(cfg.Options.Server.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Options.Server.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.Options.Server.IdleTimeout) * time.Second,
		},
	}
}

func (s *Server) Start() chan error {
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Server started", "address", s.Addr, "env", s.cfg.App.Env, "log_level", s.cfg.App.LogLevel)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()
	return serverErr
}

func (s *Server) Shutdown() error {
	slog.Info("Shutting down server...")
	timeout := time.Duration(s.cfg.Options.Server.ShutdownTimeout) * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("Server gracefully shut down.")
	return nil
}
