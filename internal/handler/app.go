package handler

import (
	"database/sql"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/go-playground/validator/v10"
)

type App struct {
	cfg       *config.Config
	db        *sql.DB
	router    *goexpress.Router
	validater *validator.Validate
	hasher    security.Hasher
}

type AppDependencies struct {
	Config    *config.Config
	DB        *sql.DB
	Router    *goexpress.Router
	Validator *validator.Validate
	Hasher    security.Hasher
}

func NewApp(deps *AppDependencies) *App {
	app := &App{
		cfg:       deps.Config,
		db:        deps.DB,
		router:    deps.Router,
		validater: deps.Validator,
		hasher:    deps.Hasher,
	}
	app.SetupMiddlewares()
	return app
}

func (a *App) Router() *goexpress.Router {
	return a.router
}

func (a *App) SetupMiddlewares() {
	a.router.Use(goexpress.RecoverFromPanic)
	a.router.Use(goexpress.LogRequest)
}

func (a *App) SetupRoutes() {
	repo := repository.NewRepository(a.db)
	svc := service.NewService(repo, a.hasher)

	apiHandler := NewHandler(*svc)
	mountAPIRoutes(a.router, apiHandler, a.validater)
}
