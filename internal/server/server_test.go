package server_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/server"
)

func TestServer_StartAndResponds(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "OK")
	})

	cfg := &config.ServerConfig{
		Port: 8081,
		Options: &config.ServerOptions{
			ReadTimeout:     1,
			WriteTimeout:    1,
			IdleTimeout:     1,
			ShutdownTimeout: 1,
		},
	}

	srv := server.New(cfg, handler)
	errCh := srv.Start()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost:8081")
	if err != nil {
		t.Fatalf("failed to GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	if err := srv.Shutdown(); err != nil {
		t.Errorf("shutdown error: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Errorf("unexpected server error: %v", err)
	}
}

func TestServer_CancelsRequestContextOnShutdown(t *testing.T) {
	ctxChan := make(chan context.Context, 1)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxChan <- r.Context()
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.ServerConfig{
		Port: 8082,
		Options: &config.ServerOptions{
			ReadTimeout:     1,
			WriteTimeout:    1,
			IdleTimeout:     1,
			ShutdownTimeout: 1,
		},
	}

	srv := server.New(cfg, handler)
	errCh := srv.Start()
	time.Sleep(100 * time.Millisecond)

	_, err := http.Get("http://localhost:8082")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	reqCtx := <-ctxChan

	if err := srv.Shutdown(); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	select {
	case <-reqCtx.Done():
		// expected
	case <-time.After(200 * time.Millisecond):
		t.Error("request context was not cancelled after shutdown")
	}

	if err := <-errCh; err != nil {
		t.Errorf("unexpected error from Start(): %v", err)
	}
}

func TestServer_StartFailsIfPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", ":8083")
	if err != nil {
		t.Skipf("could not bind to port 8083: %v", err)
	}
	defer ln.Close()

	cfg := &config.ServerConfig{
		Port: 8083,
		Options: &config.ServerOptions{
			ReadTimeout:     1,
			WriteTimeout:    1,
			IdleTimeout:     1,
			ShutdownTimeout: 1,
		},
	}

	srv := server.New(cfg, http.NewServeMux())
	errCh := srv.Start()

	err = <-errCh
	if err == nil {
		t.Fatal("expected error from Start(), got nil")
	}
}
