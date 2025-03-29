package service

import (
	"github.com/ferdiebergado/goweb/internal/pkg/security"
	"github.com/ferdiebergado/goweb/internal/repository"
)

type Service struct {
	Base BaseService
	User UserService
}

func NewService(repo *repository.Repository, hasher security.Hasher) *Service {
	return &Service{
		Base: NewBaseService(repo.Base),
		User: NewUserService(repo.User, hasher),
	}
}
