package main

import (
	"database/sql"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
	"github.com/ferdiebergado/gojeep/internal/router"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/go-playground/validator/v10"
)

type app struct {
	cfg       *config.Config
	db        *sql.DB
	handler   router.Router
	validater *validator.Validate
	hasher    security.Hasher
	mailer    email.Mailer
	signer    security.Signer
}

type dependencies struct {
	Config    *config.Config
	DB        *sql.DB
	Router    router.Router
	Validator *validator.Validate
	Hasher    security.Hasher
	Mailer    email.Mailer
	Signer    security.Signer
}

func newApp(deps *dependencies) *app {
	app := &app{
		cfg:       deps.Config,
		db:        deps.DB,
		handler:   deps.Router,
		validater: deps.Validator,
		hasher:    deps.Hasher,
		mailer:    deps.Mailer,
		signer:    deps.Signer,
	}
	app.SetupMiddlewares()
	return app
}

func (a *app) Router() router.Router {
	return a.handler
}

func (a *app) SetupMiddlewares() {
	a.handler.Use(goexpress.RecoverFromPanic)
	a.handler.Use(goexpress.LogRequest)
}

func (a *app) SetupRoutes() {
	repo := repository.NewRepository(a.db)
	deps := &service.Dependencies{
		Repo:   *repo,
		Hasher: a.hasher,
		Signer: a.signer,
		Mailer: a.mailer,
		Cfg:    a.cfg,
	}
	svc := service.NewService(deps)

	apiHandler := handler.New(*svc)
	handler.MountRoutes(a.handler, apiHandler, a.validater)
}
