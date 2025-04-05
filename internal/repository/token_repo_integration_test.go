package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/infra/db"
	"github.com/ferdiebergado/gojeep/internal/repository"
	"github.com/ferdiebergado/gopherkit/env"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

const (
	path       = "../../"
	envFile    = ".env.testing"
	configFile = "config.json"
)

func TestTokenRepoIntegrationSaveToken(t *testing.T) {
	const (
		token = "sampletoken"
		email = "123@test.com"
		ttl   = "5m"
	)

	duration, err := time.ParseDuration(ttl)
	if err != nil {
		t.Fatal("invalid duration string")
	}
	tz := time.Now().Add(duration)

	if err := env.Load(path + envFile); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.New(path + configFile)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	db, err := db.Connect(ctx, &cfg.Db)
	if err != nil {
		t.Fatal(err)
	}

	repo := repository.NewTokenRepo(db)
	err = repo.SaveToken(context.Background(), token, email, tz)
	assert.NoError(t, err)

	t.Cleanup(func() {
		_, err := db.ExecContext(ctx, "TRUNCATE tokens")
		if err != nil {
			t.Error(err)
		}
	})
}
