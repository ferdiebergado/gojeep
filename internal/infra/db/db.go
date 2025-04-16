package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
)

func Connect(ctx context.Context, cfg *config.DBConfig) (*sql.DB, error) {
	slog.Info("Connecting to the database")

	const dbStr = "postgres://%s:%s@%s:%d/%s?sslmode=%s"
	dbOpts := cfg.Options
	dsn := fmt.Sprintf(dbStr, cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DB, cfg.SSLMode)
	db, err := sql.Open(dbOpts.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	pingTimeout := time.Duration(dbOpts.PingTimeout) * time.Second
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	db.SetMaxOpenConns(dbOpts.MaxOpenConns)
	db.SetMaxIdleConns(dbOpts.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(dbOpts.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(dbOpts.ConnMaxIdle) * time.Second)

	slog.Info("Connected to the database", "db", cfg.DB)
	return db, nil
}
