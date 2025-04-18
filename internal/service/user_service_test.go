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
		userID         = "1"
		testEmail      = "abc@example.com"
		testPass       = "test"
		testPassHashed = "hashed"
		token          = "testtoken"
		title          = "Email verification" // TODO: move strings to message package
		subject        = "Verify your email"
		tmpl           = "verification"
	)

	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockUserRepository(ctrl)
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
		Model: model.Model{ID: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
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
	mockSigner.EXPECT().Sign(userID, []string{audience}, time.Duration(cfg.Email.Options.VerifyTTL)*time.Second).
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
		id    = "1"
		email = "test@example.com"
		token = "token"
	)

	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockUserRepository(ctrl)
	mockHasher := secMock.NewMockHasher(ctrl)
	mockMailer := mailMock.NewMockMailer(ctrl)
	mockSigner := secMock.NewMockSigner(ctrl)

	cfg := &config.Config{
		Server: config.ServerConfig{
			URL: "http://localhost:8888",
		},
	}

	ctx := context.Background()
	mockRepo.EXPECT().VerifyUser(ctx, id).Return(nil)
	mockSigner.EXPECT().Verify(token).Return(id, nil)

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

	cfg := &config.Config{Server: config.ServerConfig{URL: "http://localhost:8888"}, JWT: config.JWTOptions{Duration: 30}}
	loginParams := service.LoginUserParams{Email: testEmail, Password: testPass}
	verifiedAt := time.Date(2024, 1, 1, 1, 1, 1, 1, time.UTC)
	user := &model.User{
		Model:        model.Model{ID: "1"},
		Email:        testEmail,
		PasswordHash: hashedPass,
		VerifiedAt:   &verifiedAt,
	}

	testCases := []struct {
		name         string
		repoUser     *model.User
		repoErr      error
		hasherResult bool
		hasherErr    error
		wantToken    string
		wantErr      error
	}{
		{
			name:         "Success_ValidCredentials",
			repoUser:     user,
			hasherResult: true,
			wantToken:    "mocked_access_token",
		},
		{
			name:    "Failure_UserNotFound",
			repoErr: service.ErrUserNotFound,
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
			wantErr: service.ErrUserNotVerified,
		},
		{
			name:         "Failure_InvalidPassword",
			repoUser:     user,
			hasherResult: false,
			wantErr:      service.ErrUserNotFound,
		},
		{
			name:    "Failure_RepoError",
			repoErr: errors.New("database failure"),
			wantErr: errors.New("database failure"),
		},
		{
			name:      "Failure_HasherError",
			repoUser:  user,
			hasherErr: errors.New("hash mismatch"),
			wantErr:   errors.New("hash mismatch"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockUserRepository(ctrl)
			mockHasher := secMock.NewMockHasher(ctrl)
			mockSigner := secMock.NewMockSigner(ctrl)

			if tc.repoUser != nil && tc.repoErr == nil && tc.wantToken != "" {
				mockSigner.EXPECT().Sign(tc.repoUser.ID, []string{cfg.JWT.Issuer}, 30*time.Minute).
					Return("mocked_access_token", nil)
			}

			ctx := context.Background()
			mockRepo.EXPECT().
				FindUserByEmail(ctx, testEmail).
				Return(tc.repoUser, tc.repoErr)

			if tc.repoErr == nil {
				mockHasher.EXPECT().
					Verify(testPass, tc.repoUser.PasswordHash).
					Return(tc.hasherResult, tc.hasherErr)
			}

			svc := service.NewUserService(&service.UserServiceDeps{
				Repo:   mockRepo,
				Hasher: mockHasher,
				Cfg:    cfg,
				Signer: mockSigner,
			})

			token, err := svc.LoginUser(ctx, loginParams)

			if tc.wantErr != nil {
				assert.ErrorContains(t, err, tc.wantErr.Error())
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantToken, token)
			}
		})
	}
}
