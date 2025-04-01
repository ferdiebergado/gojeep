package service

import (
	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
)

type Service struct {
	Base  BaseService
	User  UserService
	Token TokenService
}

// TODO: refactor arguments into a struct
func NewService(repo *repository.Repository, hasher security.Hasher, mailer email.Mailer, signer security.Signer, cfg config.AppConfig) *Service {
	tokenService := NewTokenService(repo.Token)
	return &Service{
		Base:  NewBaseService(repo.Base),
		User:  NewUserService(repo.User, tokenService, hasher, mailer, signer, cfg),
		Token: tokenService,
	}
}
