package service_test

import (
	"context"
	"database/sql"
	"errors"
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
		Server: config.ServerConfig{
			URL: "http://localhost:8888",
		},
		Email: config.SMTPConfig{
			Options: config.EmailOptions{
				VerifyTTL: 300,
			},
		},
	}

	audience := cfg.Server.URL + "/auth/verify"
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
	mockSigner.EXPECT().Sign(testEmail, []string{audience}, time.Duration(cfg.Email.Options.VerifyTTL)*time.Second).
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
		Server: config.ServerConfig{
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

func TestUserService_LoginUser(t *testing.T) {
	t.Parallel()
	const (
		testEmail  = "abc@example.com"
		testPass   = "test"
		hashedPass = "hashed"
	)

	cfg := &config.Config{Server: config.ServerConfig{URL: "http://localhost:8888"}}
	loginParams := service.LoginUserParams{Email: testEmail, Password: testPass}
	user := &model.User{
		Model:        model.Model{ID: "1"},
		Email:        testEmail,
		PasswordHash: hashedPass,
		VerifiedAt: sql.NullTime{
			Time:  time.Date(2024, 1, 1, 1, 1, 1, 1, time.UTC),
			Valid: true,
		},
	}

	testCases := []struct {
		name         string
		repoUser     *model.User
		repoErr      error
		hasherResult bool
		hasherErr    error
		wantOk       bool
		wantErr      error
	}{
		{
			name:         "Success_ValidCredentials",
			repoUser:     user,
			hasherResult: true,
			wantOk:       true,
		},
		{
			name:    "Failure_UserNotFound",
			repoErr: service.ErrUserNotFound,
			wantOk:  false,
			wantErr: service.ErrUserNotFound,
		},
		{
			name: "Failure_UserUnverified",
			repoUser: &model.User{
				Model:        model.Model{ID: "1"},
				Email:        testEmail,
				PasswordHash: hashedPass,
			},
			repoErr: service.ErrUserNotVerified,
			wantOk:  false,
			wantErr: service.ErrUserNotVerified,
		},
		{
			name:         "Failure_InvalidPassword",
			repoUser:     user,
			hasherResult: false,
			wantOk:       false,
		},
		{
			name:    "Failure_RepoError",
			repoErr: errors.New("database failure"),
			wantOk:  false,
			wantErr: errors.New("database failure"),
		},
		{
			name:      "Failure_HasherError",
			repoUser:  user,
			hasherErr: errors.New("hash mismatch"),
			wantOk:    false,
			wantErr:   errors.New("hash mismatch"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockUserRepo(ctrl)
			mockHasher := secMock.NewMockHasher(ctrl)

			ctx := context.Background()
			mockRepo.EXPECT().
				FindUserByEmail(ctx, testEmail).
				Return(tc.repoUser, tc.repoErr)

			if tc.repoErr == nil {
				mockHasher.EXPECT().
					Verify(testPass, hashedPass).
					Return(tc.hasherResult, tc.hasherErr)
			}

			svc := service.NewUserService(&service.UserServiceDeps{
				Repo:   mockRepo,
				Hasher: mockHasher,
				Cfg:    cfg,
			})

			ok, err := svc.LoginUser(ctx, loginParams)

			assert.Equal(t, tc.wantOk, ok)
			if tc.wantErr != nil {
				assert.ErrorContains(t, err, tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
