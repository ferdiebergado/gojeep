package service_test

import (
	"context"
	"database/sql"
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
	secMock "github.com/ferdiebergado/gojeep/internal/pkg/security/mock"
	srvMock "github.com/ferdiebergado/gojeep/internal/service/mock"
)

func TestUserServiceRegisterUser(t *testing.T) {
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
	mockTokenSvc := srvMock.NewMockTokenService(ctrl)
	mockHasher := secMock.NewMockHasher(ctrl)
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

	cfg, err := config.New("../../config.json")
	if err != nil {
		t.Fatal("failed to load config", err)
	}

	audience := cfg.App.URL + "/verify"
	ttl := 5 * time.Minute
	ctx := context.Background()
	mockRepo.EXPECT().FindUserByEmail(ctx, testEmail).Return(nil, sql.ErrNoRows)
	mockHasher.EXPECT().Hash(regParams.Password).Return(testPassHashed, nil)

	data := map[string]string{
		"Title":  title,
		"Header": subject,
		"Link":   audience + "?token=" + token,
	}

	var wg sync.WaitGroup
	wg.Add(3)

	mockMailer.EXPECT().SendHTML([]string{testEmail}, subject, tmpl, data).Do(func(to []string, subj, tmplName string, data map[string]string) {
		defer wg.Done()
	})

	mockTokenSvc.EXPECT().Sign(testEmail, []string{audience}, ttl).Do(func(string, []string, time.Duration) {
		defer wg.Done()
	}).Return(token, nil)

	mockTokenSvc.EXPECT().SaveToken(ctx, token, testEmail, ttl).Do(func(context.Context, string, string, time.Duration) {
		defer wg.Done()
	}).Return(nil)

	mockRepo.EXPECT().CreateUser(ctx, createParams).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockTokenSvc, mockHasher, mockMailer, cfg.App)
	newUser, err := userService.RegisterUser(ctx, regParams)
	assert.NoError(t, err)
	assert.NotNil(t, newUser)
	assert.NotZero(t, newUser.ID)
	assert.Equal(t, createParams.Email, newUser.Email, "Emails must match")
	assert.NotZero(t, newUser.CreatedAt)
	assert.NotZero(t, newUser.UpdatedAt)

	wg.Wait()
}
