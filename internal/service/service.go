package service

import (
	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
)

type Dependencies struct {
	Repo   repository.Repository
	Hasher security.Hasher
	Signer security.Signer
	Mailer email.Mailer
	Cfg    config.AppConfig
}

type Service struct {
	Base BaseService
	User UserService
}

func NewService(deps *Dependencies) *Service {
	userSvcDeps := &UserServiceDeps{
		Repo:   deps.Repo.User,
		Hasher: deps.Hasher,
		Signer: deps.Signer,
		Mailer: deps.Mailer,
		Cfg:    deps.Cfg,
	}
	return &Service{
		Base: NewBaseService(deps.Repo.Base),
		User: NewUserService(userSvcDeps),
	}
}
