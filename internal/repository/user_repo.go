//go:generate mockgen -destination=mock/user_repo_mock.go -package=mock . UserRepo
package repository

import (
	"context"
	"database/sql"

	"github.com/ferdiebergado/gojeep/internal/model"
)

type UserRepo interface {
	CreateUser(ctx context.Context, params CreateUserParams) (*model.User, error)
	FindUserByEmail(ctx context.Context, email string) (*model.User, error)
	VerifyUser(ctx context.Context, email string) error
}

type userRepo struct {
	db *sql.DB
}

var _ UserRepo = (*userRepo)(nil)

func NewUserRepository(db *sql.DB) UserRepo {
	return &userRepo{db: db}
}

type CreateUserParams struct {
	Email        string
	PasswordHash string
}

const QueryUserCreate = `
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, created_at, updated_at
`

func (r *userRepo) CreateUser(ctx context.Context, params CreateUserParams) (*model.User, error) {
	var user model.User
	if err := r.db.QueryRowContext(ctx, QueryUserCreate, params.Email, params.PasswordHash).
		Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

const QueryUserFindByEmail = `
SELECT id, email, created_at, updated_at FROM users
WHERE email = $1
LIMIT 1
`

func (r *userRepo) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.QueryRowContext(ctx, QueryUserFindByEmail, email).
		Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

const QueryUserVerify = `
UPDATE users
SET verified_at = NOW()
WHERE email = $1
`

func (r *userRepo) VerifyUser(ctx context.Context, email string) error {
	_, err := r.db.ExecContext(ctx, QueryUserVerify, email)

	if err != nil {
		return err
	}

	return nil
}
