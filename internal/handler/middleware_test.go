package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ferdiebergado/gojeep/internal/handler"
	"github.com/ferdiebergado/gojeep/internal/pkg/security/mock"
	"go.uber.org/mock/gomock"
)

func TestRequireAuth(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := handler.FromUserContext(r.Context())
		if !ok {
			t.Fatal("User context not found")
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(user))
		if err != nil {
			t.Fatal(err)
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

func TestWriteHeaderOnce(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := context.Background()
	w := handler.NewSafeResponseWriter(ctx, rec)

	w.WriteHeader(http.StatusAccepted)
	w.WriteHeader(http.StatusTeapot) // Should be ignored

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if w.Status() != http.StatusAccepted {
		t.Errorf("SafeResponseWriter status mismatch: got %d", w.Status())
	}
}

func TestWriteImplicitHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := context.Background()
	w := handler.NewSafeResponseWriter(ctx, rec)

	body := []byte("hello")
	n, err := w.Write(body)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if n != len(body) {
		t.Errorf("expected %d bytes written, got %d", len(body), n)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected default status %d, got %d", http.StatusOK, rec.Code)
	}
	if w.Status() != http.StatusOK {
		t.Errorf("SafeResponseWriter status mismatch: got %d", w.Status())
	}
	if w.BytesWritten() != len(body) {
		t.Errorf("expected %d bytes written, got %d", len(body), w.BytesWritten())
	}
}

func TestNoWriteAfterContextCancel(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w := handler.NewSafeResponseWriter(ctx, rec)

	w.WriteHeader(http.StatusOK)
	n, err := w.Write([]byte("should not write"))

	if n != 0 {
		t.Errorf("expected 0 bytes written, got %d", n)
	}
	if err != nil {
		t.Errorf("unexpected error on canceled write: %v", err)
	}
	if w.BytesWritten() != 0 {
		t.Errorf("expected 0 bytes written, got %d", w.BytesWritten())
	}
}

func TestConcurrentWrites(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := context.Background()
	w := handler.NewSafeResponseWriter(ctx, rec)

	done := make(chan struct{}, 2)

	go func() {
		w.WriteHeader(http.StatusCreated)
		done <- struct{}{}
	}()
	go func() {
		_, _ = w.Write([]byte("hi"))
		done <- struct{}{}
	}()

	timeout := time.After(1 * time.Second)
	for range 2 {
		select {
		case <-done:
		case <-timeout:
			t.Fatal("concurrent writes timed out")
		}
	}
}
