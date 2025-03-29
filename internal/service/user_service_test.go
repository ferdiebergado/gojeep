package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ferdiebergado/goweb/internal/model"
	"github.com/ferdiebergado/goweb/internal/repository"
	"github.com/ferdiebergado/goweb/internal/repository/mock"
	"github.com/ferdiebergado/goweb/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	secMock "github.com/ferdiebergado/goweb/internal/pkg/security/mock"
)

func TestUserService_RegisterUser(t *testing.T) {
	const (
		testEmail      = "abc@example.com"
		testPass       = "test"
		testPassHashed = "hashed"
	)
	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockUserRepo(ctrl)
	mockHasher := secMock.NewMockHasher(ctrl)

	regParams := service.RegisterUserParams{
		Email:    testEmail,
		Password: testPass,
	}

	params := repository.CreateUserParams{
		Email:        regParams.Email,
		PasswordHash: testPassHashed,
	}

	user := &model.User{
		Model: model.Model{ID: "1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Email: testEmail,
	}

	mockRepo.EXPECT().FindUserByEmail(context.Background(), testEmail).Return(nil, sql.ErrNoRows)

	mockHasher.EXPECT().Hash(regParams.Password).Return(testPassHashed, nil)

	ctx := context.Background()
	mockRepo.EXPECT().CreateUser(ctx, params).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockHasher)

	newUser, err := userService.RegisterUser(ctx, regParams)
	assert.NoError(t, err)
	assert.NotNil(t, newUser)
	assert.NotZero(t, newUser.ID)
	assert.Equal(t, params.Email, newUser.Email, "Emails must match")
	assert.NotZero(t, newUser.CreatedAt)
	assert.NotZero(t, newUser.UpdatedAt)
}
