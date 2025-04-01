package repository_test

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/ferdiebergado/gojeep/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestTokenServiceSaveToken(t *testing.T) {
	const (
		token = "sampletoken"
		email = "123@test.com"
		ttl   = "5m"
	)

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	duration, err := time.ParseDuration(ttl)
	if err != nil {
		t.Fatal("invalid duration string")
	}
	tz := time.Now().Add(duration)

	mock.ExpectExec(repository.QuerySaveToken).
		WithArgs(token, email, tz).WillReturnResult(sqlmock.NewResult(1, 1))

	repo := repository.NewTokenRepo(db)
	err = repo.SaveToken(context.Background(), token, email, tz)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
