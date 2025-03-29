package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/ferdiebergado/goweb/internal/config"
)

func Connect(ctx context.Context, cfg *config.DBConfig) (*sql.DB, error) {
	const dbStr = "postgres://%s:%s@%s:%d/%s?sslmode=%s"
	slog.Info("Connecting to the database")
	dsn := fmt.Sprintf(dbStr, cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DB, cfg.SSLMode)
	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.PingTimeout)*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdle) * time.Second)

	slog.Info("Connected to the database", "db", cfg.DB)
	return db, nil
}
