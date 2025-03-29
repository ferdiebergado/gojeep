//go:generate mockgen -destination=mock/base_service_mock.go -package=mock . BaseService
package service

import (
	"context"
	"time"

	"github.com/ferdiebergado/goweb/internal/repository"
)

type BaseService interface {
	PingDB(ctx context.Context) error
}

type service struct {
	repo repository.BaseRepository
}

func NewBaseService(repo repository.BaseRepository) BaseService {
	return &service{
		repo: repo,
	}
}

func (s *service) PingDB(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return s.repo.Ping(pingCtx)
}
