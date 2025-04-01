package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ferdiebergado/gojeep/internal/pkg/test"
	"github.com/ferdiebergado/gojeep/internal/repository/mock"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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

	ctx := context.Background()

	mockRepo.EXPECT().SaveToken(ctx, token, email, test.WithinDuration(expectedTZ, 100*time.Millisecond)).Return(nil)

	tokenService := service.NewTokenService(mockRepo)
	err := tokenService.SaveToken(ctx, token, email, ttl)
	assert.NoError(t, err)
}
