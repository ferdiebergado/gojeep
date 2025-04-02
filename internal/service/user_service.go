//go:generate mockgen -destination=mock/user_service_mock.go -package=mock . UserService
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/model"
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
)

type UserService interface {
	RegisterUser(ctx context.Context, params RegisterUserParams) (*model.User, error)
}

type userService struct {
	repo     repository.UserRepo
	tokenSvc TokenService
	hasher   security.Hasher
	mailer   email.Mailer
	cfg      config.AppConfig
}

var _ UserService = (*userService)(nil)
var ErrDuplicateUser = errors.New("duplicate user")

// TODO: move arguments into a struct
func NewUserService(repo repository.UserRepo, tokenSvc TokenService, hasher security.Hasher, mailer email.Mailer, cfg config.AppConfig) UserService {
	return &userService{
		repo:     repo,
		tokenSvc: tokenSvc,
		hasher:   hasher,
		mailer:   mailer,
		cfg:      cfg,
	}
}

type RegisterUserParams struct {
	Email    string
	Password string
}

func (s *userService) RegisterUser(ctx context.Context, params RegisterUserParams) (*model.User, error) {
	existing, err := s.repo.FindUserByEmail(ctx, params.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists: %w", params.Email, ErrDuplicateUser)
	}

	hash, err := s.hasher.Hash(params.Password)
	if err != nil {
		return nil, fmt.Errorf("hasher hash: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{Email: params.Email, PasswordHash: hash})
	if err != nil {
		return nil, fmt.Errorf("create user %s: %w", params.Email, err)
	}

	go s.sendVerificationEmail(user)

	return user, nil
}

func (s *userService) sendVerificationEmail(user *model.User) {
	slog.Info("Sending verification email...")

	const (
		title   = "Email verification"
		subject = "Verify your email"
		ttl     = 5 * time.Minute
	)

	audience := s.cfg.URL + "/verify"
	token, err := s.tokenSvc.Sign(user.Email, []string{audience}, ttl)
	if err != nil {
		slog.Error("failed to generate token", "reason", err)
		return
	}

	if err := s.tokenSvc.SaveToken(context.Background(), token, user.Email, ttl); err != nil {
		slog.Error("unable to save token", "reason", err)
		return
	}

	data := map[string]string{
		"Title":  title,
		"Header": subject,
		"Link":   audience + "?token=" + token,
	}
	if err := s.mailer.SendHTML([]string{user.Email}, subject, "verification", data); err != nil {
		slog.Error("failed to send email", "reason", err)
		return
	}
}
