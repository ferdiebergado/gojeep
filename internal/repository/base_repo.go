//go:generate mockgen -destination=mock/base_repo_mock.go -package=mock . BaseRepository
package repository

import (
	"context"
	"database/sql"
)

type BaseRepository interface {
	Ping(ctx context.Context) error
}

type repo struct {
	db *sql.DB
}

var _ BaseRepository = (*repo)(nil)

func NewBaseRepository(db *sql.DB) BaseRepository {
	return &repo{
		db: db,
	}
}

func (r *repo) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
