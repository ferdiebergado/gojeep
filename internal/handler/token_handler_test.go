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
		path       = "/api/verify"
		method     = http.MethodGet
		wantedMime = handler.MimeJSON
	)

	ctrl := gomock.NewController(t)
	mockTokenService := mock.NewMockTokenService(ctrl)
	tokenHandler := handler.NewTokenHandler(mockTokenService)
	r := goexpress.New()
	r.Get(path, tokenHandler.HandleVerifyToken)

	var tests = []struct {
		name         string
		token        string
		shouldVerify bool
		wantedStatus int
		wantedMsg    string
	}{
		{"valid token", "testtoken", true, http.StatusOK, "Verification successful!"},
		{"empty token", "", false, http.StatusBadRequest, "Invalid input."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldVerify {
				mockTokenService.EXPECT().Verify(tt.token).Return("", nil)
			}
			uri := fmt.Sprintf("%s?token=%s", path, tt.token)
			req := httptest.NewRequest(method, uri, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantedStatus, res.StatusCode)
			assert.Equal(t, wantedMime, res.Header.Get(handler.HeaderContentType))

			var apiRes handler.Response[any]
			if err := json.Unmarshal(rr.Body.Bytes(), &apiRes); err != nil {
				t.Fatal(message.Get("jsonFailed"), err)
			}

			assert.Equal(t, tt.wantedMsg, apiRes.Message)
		})
	}
}
