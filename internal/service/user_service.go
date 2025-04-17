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
	VerifyUser(ctx context.Context, token string) error
	LoginUser(ctx context.Context, params LoginUserParams) (string, error)
}

type UserServiceDeps struct {
	Repo   repository.UserRepository
	Hasher security.Hasher
	Signer security.Signer
	Mailer email.Mailer
	Cfg    *config.Config
}

type userService struct {
	repo   repository.UserRepository
	hasher security.Hasher
	signer security.Signer
	mailer email.Mailer
	cfg    *config.Config
}

var _ UserService = (*userService)(nil)
var (
	ErrUserNotFound    = errors.New("invalid email or password")
	ErrUserNotVerified = errors.New("email not verified")
)

func NewUserService(deps *UserServiceDeps) UserService {
	return &userService{
		repo:   deps.Repo,
		hasher: deps.Hasher,
		mailer: deps.Mailer,
		signer: deps.Signer,
		cfg:    deps.Cfg,
	}
}

type RegisterUserParams struct {
	Email    string
	Password string
}

func (p *RegisterUserParams) LogValue() slog.Value {
	return slog.AnyValue(nil)
}

type LoginUserParams struct {
	Email    string
	Password string
}

func (p *LoginUserParams) LogValue() slog.Value {
	return slog.AnyValue(nil)
}

type DuplicateUserError struct {
	Email string
}

func (d *DuplicateUserError) Error() string {
	return fmt.Sprintf("user with email %s already exists", d.Email)
}

func (s *userService) RegisterUser(ctx context.Context, params RegisterUserParams) (*model.User, error) {
	email := params.Email
	existing, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existing != nil {
		return nil, &DuplicateUserError{Email: email}
	}

	hash, err := s.hasher.Hash(params.Password)
	if err != nil {
		return nil, fmt.Errorf("hasher hash: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{Email: email, PasswordHash: hash})
	if err != nil {
		return nil, fmt.Errorf("create user %s: %w", email, err)
	}

	go s.sendVerificationEmail(user)

	return user, nil
}

func (s *userService) sendVerificationEmail(user *model.User) {
	slog.Info("Sending verification email...")

	const (
		title   = "Email verification"
		subject = "Verify your email"
	)

	audience := s.cfg.Server.URL + "/auth/verify"
	ttl := time.Duration(s.cfg.Email.Options.VerifyTTL) * time.Second
	token, err := s.signer.Sign(user.ID, []string{audience}, ttl)
	if err != nil {
		slog.Error("failed to generate token", "reason", err)
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

func (s *userService) VerifyUser(ctx context.Context, token string) error {
	userID, err := s.signer.Verify(token)
	if err != nil {
		return err
	}

	return s.repo.VerifyUser(ctx, userID)
}

func (s *userService) LoginUser(ctx context.Context, params LoginUserParams) (string, error) {
	user, err := s.repo.FindUserByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	if !user.VerifiedAt.Valid {
		return "", ErrUserNotVerified
	}

	ok, err := s.hasher.Verify(params.Password, user.PasswordHash)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", ErrUserNotFound
	}

	ttl := time.Duration(s.cfg.JWT.Duration) * time.Minute
	accessToken, err := s.signer.Sign(user.ID, []string{s.cfg.JWT.Issuer}, ttl)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
