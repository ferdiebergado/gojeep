package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ferdiebergado/gojeep/internal/app"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(signalCtx); err != nil {
		slog.Error("fatal error", "reason", err)
		stop()
		os.Exit(1)
	}
}
