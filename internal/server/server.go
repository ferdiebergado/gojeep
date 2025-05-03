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
	http.Server
	stop            context.CancelFunc
	shutdownTimeout time.Duration
}

func New(cfg *config.ServerConfig, handler http.Handler) *Server {
	serverCtx, stopServer := context.WithCancel(context.Background())
	opts := cfg.Options
	return &Server{
		Server: http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      handler,
			ReadTimeout:  time.Duration(opts.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(opts.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(opts.IdleTimeout) * time.Second,
			BaseContext: func(_ net.Listener) context.Context {
				return serverCtx
			},
		},
		stop:            stopServer,
		shutdownTimeout: time.Duration(opts.ShutdownTimeout) * time.Second,
	}
}

func (s *Server) Start() chan error {
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

func (s *Server) Shutdown() error {
	slog.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.Server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	s.stop()

	slog.Info("Server gracefully shut down.")
	return nil
}
