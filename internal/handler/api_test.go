package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/goweb/internal/handler"
	"github.com/ferdiebergado/goweb/internal/pkg/message"
	"github.com/ferdiebergado/goweb/internal/service/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandlerHandleHealth(t *testing.T) {
	const (
		url = "/api/health"
		msg = "healthy"
	)

	ctrl := gomock.NewController(t)
	mockService := mock.NewMockService(ctrl)
	mockService.EXPECT().PingDB(context.Background()).Return(nil)
	baseHandler := handler.NewBaseAPIHandler(mockService)
	r := goexpress.New()
	r.Get(url, baseHandler.HandleHealth)

	req := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, handler.MimeJSONUTF8, res.Header[handler.HeaderContentType][0])

	var apiRes handler.APIResponse[any]
	if err := json.Unmarshal(rr.Body.Bytes(), &apiRes); err != nil {
		t.Fatal(message.Get("jsonFailed"), err)
	}

	assert.Equal(t, msg, apiRes.Message)
}
