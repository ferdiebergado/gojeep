//go:generate mockgen -destination=mock/token_service_mock.go -package=mock . TokenService
package service

import (
	"context"
	"time"

	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
)

type TokenService interface {
	Sign(email string, audience []string, ttl time.Duration) (string, error)
	SaveToken(ctx context.Context, id, email string, ttl time.Duration) error
	Verify(tokenString string) (string, error)
}

type tokenService struct {
	repo   repository.TokenRepo
	signer security.Signer
}

var _ TokenService = (*tokenService)(nil)

func NewTokenService(repo repository.TokenRepo, signer security.Signer) TokenService {
	return &tokenService{
		repo:   repo,
		signer: signer,
	}
}

func (t *tokenService) Sign(email string, audience []string, ttl time.Duration) (string, error) {
	token, err := t.signer.Sign(email, audience, ttl)
	if err != nil {
		return "", nil
	}

	return token, nil
}

func (t *tokenService) SaveToken(ctx context.Context, id string, email string, ttl time.Duration) error {
	return t.repo.SaveToken(ctx, id, email, time.Now().Add(ttl))
}

func (t *tokenService) Verify(tokenString string) (string, error) {
	subject, err := t.signer.Verify(tokenString)
	if err != nil {
		return "", err
	}

	return subject, nil
}
