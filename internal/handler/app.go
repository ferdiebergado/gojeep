package handler

import (
	"database/sql"
	"net/http"

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
	template  *Template
	hasher    security.Hasher
}

type AppDependencies struct {
	Config    *config.Config
	DB        *sql.DB
	Router    *goexpress.Router
	Validator *validator.Validate
	Template  *Template
	Hasher    security.Hasher
}

func NewApp(deps *AppDependencies) *App {
	app := &App{
		cfg:       deps.Config,
		db:        deps.DB,
		router:    deps.Router,
		validater: deps.Validator,
		template:  deps.Template,
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
	if a.cfg.App.Env == "development" {
		const prefix = "/assets/"
		a.router.Handle("GET "+prefix, http.StripPrefix(prefix, http.FileServer(http.Dir("web/assets/"))))
	}

	repo := repository.NewRepository(a.db)
	svc := service.NewService(repo, a.hasher)

	htmlHandler := NewHandler(a.template)
	apiHandler := NewAPIHandler(*svc)

	mountRoutes(a.router, htmlHandler)
	mountAPIRoutes(a.router, apiHandler, a.validater)
}
