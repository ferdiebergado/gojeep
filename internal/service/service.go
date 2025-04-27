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
	Cfg    *config.Config
}

type Service struct {
	Base BaseService
	User AuthService
}

func NewService(deps *Dependencies) *Service {
	userSvcDeps := &AuthServiceDeps{
		Repo:   deps.Repo.User,
		Hasher: deps.Hasher,
		Signer: deps.Signer,
		Mailer: deps.Mailer,
		Cfg:    deps.Cfg,
	}
	return &Service{
		Base: NewBaseService(deps.Repo.Base),
		User: NewAuthService(userSvcDeps),
	}
}
