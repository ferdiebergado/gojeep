package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferdiebergado/goexpress"
	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/pkg/message"
	"github.com/ferdiebergado/gojeep/internal/service/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandlerHandleVerifyToken(t *testing.T) {
	const (
		path         = "/api/verify"
		method       = http.MethodGet
		token        = "testtoken"
		wantedStatus = http.StatusOK
		wantedMime   = handler.MimeJSON
		wantedMsg    = "Verification successful!"
	)

	ctrl := gomock.NewController(t)
	mockTokenService := mock.NewMockTokenService(ctrl)
	mockTokenService.EXPECT().Verify(token).Return("", nil)
	tokenHandler := handler.NewTokenHandler(mockTokenService)
	r := goexpress.New()
	r.Get(path, tokenHandler.HandleVerifyToken)

	uri := fmt.Sprintf("%s?token=%s", path, token)
	req := httptest.NewRequest(method, uri, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	assert.Equal(t, wantedStatus, res.StatusCode)
	assert.Equal(t, wantedMime, res.Header.Get(handler.HeaderContentType))

	var apiRes handler.Response[any]
	if err := json.Unmarshal(rr.Body.Bytes(), &apiRes); err != nil {
		t.Fatal(message.Get("jsonFailed"), err)
	}

	assert.Equal(t, wantedMsg, apiRes.Message)
}
