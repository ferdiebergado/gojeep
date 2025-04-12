package service_test

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/model"
	"github.com/ferdiebergado/gojeep/internal/repository"
	"github.com/ferdiebergado/gojeep/internal/repository/mock"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	mailMock "github.com/ferdiebergado/gojeep/internal/pkg/email/mock"
	"github.com/ferdiebergado/gojeep/internal/pkg/logging"
	secMock "github.com/ferdiebergado/gojeep/internal/pkg/security/mock"
)

func TestMain(m *testing.M) {
	logging.SetLogger(os.Stdout, "testing", "error")
	os.Exit(m.Run())
}

func TestUserService_RegisterUser(t *testing.T) {
	t.Parallel()
	const (
		testEmail      = "abc@example.com"
		testPass       = "test"
		testPassHashed = "hashed"
		token          = "testtoken"
		title          = "Email verification" // TODO: move strings to message package
		subject        = "Verify your email"
		tmpl           = "verification"
	)

	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockHasher := secMock.NewMockHasher(ctrl)
	mockSigner := secMock.NewMockSigner(ctrl)
	mockMailer := mailMock.NewMockMailer(ctrl)

	regParams := service.RegisterUserParams{
		Email:    testEmail,
		Password: testPass,
	}

	createParams := repository.CreateUserParams{
		Email:        regParams.Email,
		PasswordHash: testPassHashed,
	}

	user := &model.User{
		Model: model.Model{ID: "1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Email: testEmail,
	}

	cfg := &config.Config{
		App: config.AppConfig{
			URL: "http://localhost:8888",
		},
		Options: config.Options{
			Email: config.EmailOptions{
				VerifyTTL: 300,
			}},
	}

	audience := cfg.App.URL + "/verify"
	ctx := context.Background()
	mockRepo.EXPECT().FindUserByEmail(ctx, testEmail).Return(nil, sql.ErrNoRows)
	mockHasher.EXPECT().Hash(regParams.Password).Return(testPassHashed, nil)
	data := map[string]string{
		"Title":  title,
		"Header": subject,
		"Link":   audience + "?token=" + token,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	mockSigner.EXPECT().Sign(testEmail, []string{audience}, time.Duration(cfg.Options.Email.VerifyTTL)*time.Second).
		Do(func(_ string, _ []string, _ time.Duration) {
			defer wg.Done()
		}).Return(token, nil)
	mockMailer.EXPECT().SendHTML([]string{testEmail}, subject, tmpl, data).
		Do(func(_ []string, _, _ string, _ map[string]string) {
			defer wg.Done()
		})

	mockRepo.EXPECT().CreateUser(ctx, createParams).Return(user, nil)

	deps := &service.UserServiceDeps{
		Repo:   mockRepo,
		Hasher: mockHasher,
		Signer: mockSigner,
		Mailer: mockMailer,
		Cfg:    cfg,
	}

	userService := service.NewUserService(deps)
	newUser, err := userService.RegisterUser(ctx, regParams)
	assert.NoError(t, err)
	assert.NotNil(t, newUser)
	assert.NotZero(t, newUser.ID)
	assert.Equal(t, createParams.Email, newUser.Email, "Emails must match")
	assert.NotZero(t, newUser.CreatedAt)
	assert.NotZero(t, newUser.UpdatedAt)

	wg.Wait()
}

func TestUserService_VerifyUser(t *testing.T) {
	t.Parallel()
	const (
		email = "test@example.com"
		token = "token"
	)

	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockHasher := secMock.NewMockHasher(ctrl)
	mockMailer := mailMock.NewMockMailer(ctrl)
	mockSigner := secMock.NewMockSigner(ctrl)

	cfg := &config.Config{
		App: config.AppConfig{
			URL: "http://localhost:8888",
		},
	}

	ctx := context.Background()
	mockRepo.EXPECT().VerifyUser(ctx, email).Return(nil)
	mockSigner.EXPECT().Verify(token).Return(email, nil)

	deps := &service.UserServiceDeps{
		Repo:   mockRepo,
		Hasher: mockHasher,
		Signer: mockSigner,
		Mailer: mockMailer,
		Cfg:    cfg,
	}
	svc := service.NewUserService(deps)
	err := svc.VerifyUser(ctx, token)

	assert.NoError(t, err)
}
