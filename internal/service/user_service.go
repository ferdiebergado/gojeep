//go:generate mockgen -destination=mock/user_service_mock.go -package=mock . UserService
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ferdiebergado/goweb/internal/model"
	"github.com/ferdiebergado/goweb/internal/pkg/security"
	"github.com/ferdiebergado/goweb/internal/repository"
)

type UserService interface {
	RegisterUser(ctx context.Context, params RegisterUserParams) (*model.User, error)
}

type userService struct {
	repo   repository.UserRepo
	hasher security.Hasher
}

var _ UserService = (*userService)(nil)
var ErrDuplicateUser = errors.New("duplicate user")

func NewUserService(repo repository.UserRepo, hasher security.Hasher) UserService {
	return &userService{
		repo:   repo,
		hasher: hasher,
	}
}

type RegisterUserParams struct {
	Email    string
	Password string
}

func (s *userService) RegisterUser(ctx context.Context, params RegisterUserParams) (*model.User, error) {
	existing, err := s.repo.FindUserByEmail(ctx, params.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists: %w", params.Email, ErrDuplicateUser)
	}

	hash, err := s.hasher.Hash(params.Password)

	if err != nil {
		return nil, fmt.Errorf("hasher hash: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{Email: params.Email, PasswordHash: hash})

	if err != nil {
		return nil, fmt.Errorf("create user %s: %w", params.Email, err)
	}

	return user, nil
}
