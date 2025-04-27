//go:generate mockgen -destination=mock/user_service_mock.go -package=mock . UserService
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/model"
	"github.com/ferdiebergado/gojeep/internal/pkg/email"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/repository"
)

type AuthService interface {
	RegisterUser(ctx context.Context, params RegisterUserParams) (model.User, error)
	VerifyUser(ctx context.Context, token string) error
	LoginUser(ctx context.Context, params LoginUserParams) (accessToken, refreshToken string, err error)
}

type AuthServiceDeps struct {
	Repo   repository.UserRepository
	Hasher security.Hasher
	Signer security.Signer
	Mailer email.Mailer
	Cfg    *config.Config
}

type authService struct {
	repo   repository.UserRepository
	hasher security.Hasher
	signer security.Signer
	mailer email.Mailer
	cfg    *config.Config
}

var _ AuthService = (*authService)(nil)
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserNotVerified = errors.New("email not verified")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidToken    = errors.New("invalid token")
)

func NewAuthService(deps *AuthServiceDeps) AuthService {
	return &authService{
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
	return slog.GroupValue(
		slog.String("email", "*"),
		slog.String("password", "*"),
	)
}

func (s *authService) RegisterUser(ctx context.Context, params RegisterUserParams) (model.User, error) {
	email := params.Email
	existing, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.User{}, err
	}

	if !reflect.DeepEqual(existing, model.User{}) {
		return model.User{}, ErrUserExists
	}

	hash, err := s.hasher.Hash(params.Password)
	if err != nil {
		return model.User{}, fmt.Errorf("hasher hash: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, repository.CreateUserParams{Email: email, PasswordHash: hash})
	if err != nil {
		return model.User{}, fmt.Errorf("create user %s: %w", email, err)
	}

	go s.sendVerificationEmail(user)

	return user, nil
}

func (s *authService) sendVerificationEmail(user model.User) {
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

func (s *authService) VerifyUser(ctx context.Context, token string) error {
	userID, err := s.signer.Verify(token)
	if err != nil {
		return ErrInvalidToken
	}

	return s.repo.VerifyUser(ctx, userID)
}

func (s *authService) LoginUser(ctx context.Context, params LoginUserParams) (accessToken, refreshToken string, err error) {
	user, err := s.repo.FindUserByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", ErrUserNotFound
		}
		return "", "", err
	}

	if user.VerifiedAt == nil {
		return "", "", ErrUserNotVerified
	}

	ok, err := s.hasher.Verify(params.Password, user.PasswordHash)
	if err != nil {
		return "", "", err
	}

	if !ok {
		return "", "", ErrUserNotFound
	}

	ttl := time.Duration(s.cfg.JWT.Duration) * time.Minute
	accessToken, err = s.signer.Sign(user.ID, []string{s.cfg.JWT.Issuer}, ttl)
	if err != nil {
		return "", "", err
	}

	// TODO: add refresh token ttl to config
	refreshToken, err = s.signer.Sign(user.ID, []string{s.cfg.JWT.Issuer}, time.Hour*24*7)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
