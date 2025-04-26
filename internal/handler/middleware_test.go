package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/pkg/security/mock"
	"go.uber.org/mock/gomock"
)

func TestRequireAuth(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := handler.FromUserContext(r.Context())
		if !ok {
			t.Error("User context not found")
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(user))
		if err != nil {
			t.Error(err)
			return
		}
	})

	tests := []struct {
		name           string
		authHeader     string
		signerSub      string
		signerErr      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid Bearer Token",
			authHeader:     "Bearer valid.token.here",
			signerSub:      "user123",
			expectedStatus: http.StatusOK,
			expectedBody:   "user123",
		},
		{
			name:           "Missing Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"message":"Unauthorized"}`,
		},
		{
			name:           "Missing Bearer Prefix",
			authHeader:     "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"message":"Unauthorized"}`,
		},
		{
			name:           "Invalid Token - Signer Error",
			authHeader:     "Bearer invalid.token.here",
			signerErr:      errors.New("invalid signature"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"message":"Unauthorized"}`,
		},
		{
			name:           "Empty Token After Bearer",
			authHeader:     "Bearer  ",
			signerSub:      "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"message":"Unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)
			rr := httptest.NewRecorder()

			ctrl := gomock.NewController(t)

			mockSigner := mock.NewMockSigner(ctrl)

			if tt.signerErr != nil || tt.signerSub != "" {
				mockSigner.EXPECT().Verify(strings.ReplaceAll(tt.authHeader, "Bearer ", "")).Return(tt.signerSub, tt.signerErr)
			}

			handler := handler.RequireAuth(mockSigner)(nextHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}
			if strings.TrimSpace(rr.Body.String()) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, strings.TrimSpace(rr.Body.String()))
			}
		})
	}
}
