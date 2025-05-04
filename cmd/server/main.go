package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ferdiebergado/gojeep/internal/app"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()

	if err := app.Run(ctx); err != nil {
		slog.Error("fatal error", "reason", err)
		os.Exit(1)
	}
}
