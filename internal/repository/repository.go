package repository

import "database/sql"

type Repository struct {
	Base BaseRepository
	User UserRepo
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Base: NewBaseRepository(db),
		User: NewUserRepository(db),
	}
}
