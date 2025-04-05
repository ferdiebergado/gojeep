package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/test"
	"github.com/ferdiebergado/gojeep/internal/repository/mock"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	secMock "github.com/ferdiebergado/gojeep/internal/pkg/security/mock"
)

func TestTokenServiceSaveToken(t *testing.T) {
	const (
		token = "sampletoken"
		email = "123@test.com"
		ttl   = 5 * time.Second
	)

	expectedTZ := time.Now().Add(ttl)

	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockTokenRepo(ctrl)
	mockSigner := secMock.NewMockSigner(ctrl)

	ctx := context.Background()
	mockRepo.EXPECT().SaveToken(ctx, token, email, test.WithinDuration(expectedTZ, 100*time.Millisecond)).Return(nil)
	tokenService := service.NewTokenService(mockRepo, mockSigner)
	err := tokenService.SaveToken(ctx, token, email, ttl)
	assert.NoError(t, err)
}

func TestTokenServiceSignToken(t *testing.T) {
	const (
		token = "sampletoken"
		email = "123@test.com"
		ttl   = 5 * time.Second
	)

	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockTokenRepo(ctrl)
	mockSigner := secMock.NewMockSigner(ctrl)

	cfg, err := config.New("../../config.json")
	if err != nil {
		t.Fatal("failed to load config", err)
	}
	audience := cfg.App.URL + "/verify"

	mockSigner.EXPECT().Sign(email, []string{audience}, ttl).Return(token, nil)
	tokenService := service.NewTokenService(mockRepo, mockSigner)
	_, err = tokenService.Sign(email, []string{audience}, ttl)
	assert.NoError(t, err)
}

func TestTokenServiceVerifyToken(t *testing.T) {
	const (
		token = "sampletoken"
		email = "123@test.com"
		ttl   = 5 * time.Second
	)

	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockTokenRepo(ctrl)
	mockSigner := secMock.NewMockSigner(ctrl)

	mockSigner.EXPECT().Verify(token).Return(email, nil)
	tokenService := service.NewTokenService(mockRepo, mockSigner)
	_, err := tokenService.Verify(token)
	assert.NoError(t, err)
}
