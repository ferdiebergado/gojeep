package handler

import (
	"database/sql"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/go-playground/validator/v10"
)

type App struct {
	cfg       *config.Config
	db        *sql.DB
	router    Router
	validater *validator.Validate
	hasher    security.Hasher
	mailer    email.Mailer
	signer    security.Signer
}

type AppDependencies struct {
	Config    *config.Config
	DB        *sql.DB
	Router    Router
	Validator *validator.Validate
	Hasher    security.Hasher
	Mailer    email.Mailer
	Signer    security.Signer
}

func NewApp(deps *AppDependencies) *App {
	app := &App{
		cfg:       deps.Config,
		db:        deps.DB,
		router:    deps.Router,
		validater: deps.Validator,
		hasher:    deps.Hasher,
		mailer:    deps.Mailer,
		signer:    deps.Signer,
	}
	app.SetupMiddlewares()
	return app
}

func (a *App) Router() Router {
	return a.router
}

func (a *App) SetupMiddlewares() {
	a.router.Use(goexpress.RecoverFromPanic)
	a.router.Use(goexpress.LogRequest)
}

func (a *App) SetupRoutes() {
	repo := repository.NewRepository(a.db)
	deps := &service.Dependencies{
		Repo:   *repo,
		Hasher: a.hasher,
		Signer: a.signer,
		Mailer: a.mailer,
		Cfg:    a.cfg,
	}
	svc := service.NewService(deps)

	apiHandler := NewHandler(*svc)
	mountAPIRoutes(a.router, apiHandler, a.validater)
}
