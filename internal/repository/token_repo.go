//go:generate mockgen -destination=mock/tokenrepo_mock.go -package=mock . TokenRepo
package repository

import (
	"context"
	"database/sql"
	"time"
)

type TokenRepo interface {
	SaveToken(ctx context.Context, id, email string, ttl time.Time) error
}

type tokenRepo struct {
	db *sql.DB
}

var _ TokenRepo = (*tokenRepo)(nil)

func NewTokenRepo(db *sql.DB) TokenRepo {
	return &tokenRepo{
		db: db,
	}
}

const QuerySaveToken = `
INSERT INTO tokens (id, email, ttl)
VALUES ($1, $2, $3)
`

func (r *tokenRepo) SaveToken(ctx context.Context, id, email string, ttl time.Time) error {
	if _, err := r.db.ExecContext(ctx, QuerySaveToken, id, email, ttl); err != nil {
		return err
	}

	return nil
}
