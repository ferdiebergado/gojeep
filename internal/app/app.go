package app

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

// Dependencies holds all external dependencies required by the application.
type dependencies struct {
	Config    *config.Config
	DB        *sql.DB
	Router    router.Router
	Validator *validator.Validate
	Hasher    security.Hasher
	Mailer    email.Mailer
	Signer    security.Signer
}

// NewDependencies creates and initializes all dependencies.
func newDependencies(cfg *config.Config, db *sql.DB, validate *validator.Validate) (*dependencies, error) {
	mailer, err := email.New(cfg.Email)
	if err != nil {
		return nil, err
	}

	deps := &dependencies{
		Config:    cfg,
		DB:        db,
		Router:    router.New(),
		Validator: validate,
		Hasher:    security.NewArgon2Hasher(cfg.Hash, cfg.Server.Key),
		Mailer:    mailer,
		Signer:    security.NewSigner(cfg),
	}
	return deps, nil
}

// Application represents the core application with all dependencies wired.
type application struct {
	cfg       *config.Config
	db        *sql.DB
	handler   router.Router
	validater *validator.Validate
	hasher    security.Hasher
	mailer    email.Mailer
	signer    security.Signer
}

// New creates a new Application instance with the provided dependencies.
func New(deps *dependencies) *application {
	app := &application{
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

func (a *application) Router() router.Router {
	return a.handler
}

func (a *application) SetupMiddlewares() {
	a.handler.Use(goexpress.RecoverFromPanic)
	a.handler.Use(goexpress.LogRequest)
}

func (a *application) SetupRoutes() {
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
