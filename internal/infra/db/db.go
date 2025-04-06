package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
)

func Connect(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	slog.Info("Connecting to the database")

	const dbStr = "postgres://%s:%s@%s:%d/%s?sslmode=%s"
	dbCfg := cfg.DB
	dbOpts := cfg.Options.DB
	dsn := fmt.Sprintf(dbStr, dbCfg.User, dbCfg.Pass, dbCfg.Host, dbCfg.Port, dbCfg.DB, dbCfg.SSLMode)
	db, err := sql.Open(cfg.Options.DB.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, time.Duration(dbOpts.PingTimeout)*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	db.SetMaxOpenConns(dbOpts.MaxOpenConns)
	db.SetMaxIdleConns(dbOpts.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(dbOpts.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(dbOpts.ConnMaxIdle) * time.Second)

	slog.Info("Connected to the database", "db", dbCfg.DB)
	return db, nil
}
