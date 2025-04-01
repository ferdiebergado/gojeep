//go:generate mockgen -destination=mock/token_service_mock.go -package=mock . TokenService
package service

import (
	"context"
	"time"

	"github.com/ferdiebergado/gojeep/internal/repository"
)

type TokenService interface {
	SaveToken(ctx context.Context, id, email string, ttl time.Duration) error
}

type tokenService struct {
	repo repository.TokenRepo
}

var _ TokenService = (*tokenService)(nil)

func NewTokenService(repo repository.TokenRepo) TokenService {
	return &tokenService{
		repo: repo,
	}
}

func (t *tokenService) SaveToken(ctx context.Context, id string, email string, ttl time.Duration) error {
	return t.repo.SaveToken(ctx, id, email, time.Now().Add(ttl))
}
