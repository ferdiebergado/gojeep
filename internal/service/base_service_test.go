package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ferdiebergado/goweb/internal/repository/mock"
	"github.com/ferdiebergado/goweb/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestService_PingDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepo := mock.NewMockRepository(ctrl)
	ctx := context.Background()

	mockRepo.EXPECT().Ping(gomock.Any()).Do(func(ctx context.Context) {
		deadline, ok := ctx.Deadline()
		assert.True(t, ok, "Expected context to have a deadline")
		timeRemaining := time.Until(deadline)
		assert.LessOrEqual(t, timeRemaining, 5*time.Second, "Deadline should be within 5 seconds")
		assert.Greater(t, timeRemaining, 0*time.Second, "Deadline should be greater than zero") // Check it is not zero
	}).Return(nil)

	mockService := service.NewBaseService(mockRepo)

	err := mockService.PingDB(ctx)
	assert.NoError(t, err)
}
