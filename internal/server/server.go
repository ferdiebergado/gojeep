package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
)

type Server struct {
	*http.Server
	stop            context.CancelFunc
	shutdownTimeout time.Duration
}

func New(ctx context.Context, cfg *config.ServerConfig, handler http.Handler) *Server {
	serverCtx, stopServer := context.WithCancel(ctx)
	opts := cfg.Options
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(opts.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(opts.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(opts.IdleTimeout) * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return serverCtx
		}}

	return &Server{
		Server:          srv,
		stop:            stopServer,
		shutdownTimeout: time.Duration(opts.ShutdownTimeout) * time.Second,
	}
}

func (s *Server) Start() chan error {
	slog.Info("Starting server...")
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Server started", "address", s.Addr)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()
	return serverErr
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down server...")
	s.stop()

	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()

	if err := s.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("Shutdown complete.")
	return nil
}
