package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/pkg/message"
	"github.com/ferdiebergado/gojeep/internal/service/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandlerHandleHealth(t *testing.T) {
	t.Parallel()
	const (
		url = "/health"
		msg = "healthy"
	)

	ctrl := gomock.NewController(t)
	mockService := mock.NewMockService(ctrl)
	mockService.EXPECT().PingDB(context.Background()).Return(nil)
	baseHandler := handler.NewBaseHandler(mockService)
	healthHandler := http.HandlerFunc(baseHandler.HandleHealth)

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	healthHandler.ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, handler.MimeJSON, res.Header[handler.HeaderContentType][0])

	var apiRes handler.Response[any]
	if err := json.Unmarshal(rr.Body.Bytes(), &apiRes); err != nil {
		t.Fatal(message.JSONDecodeFailure, err)
	}

	assert.Equal(t, msg, apiRes.Message)
}
