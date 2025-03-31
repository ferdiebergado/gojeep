package service

import (
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
)

type Service struct {
	Base BaseService
	User UserService
}

func NewService(repo *repository.Repository, hasher security.Hasher, mailer email.Mailer) *Service {
	return &Service{
		Base: NewBaseService(repo.Base),
		User: NewUserService(repo.User, hasher, mailer),
	}
}
