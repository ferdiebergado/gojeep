package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ferdiebergado/gojeep/internal/pkg/message"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/go-playground/validator/v10"
)

func DecodeJSON[T any]() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("Checking content-type...")
			contentType := r.Header.Get(HeaderContentType)

			if contentType != MimeJSON {
				unsupportedContentTypeResponse(w, fmt.Errorf("Invalid content-type: %s", contentType), message.UserInputInvalid)
				return
			}

			slog.Info("Decoding json payload...")
			var decoded T
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&decoded); err != nil {
				badRequestResponse(w, err, message.UserInputInvalid)
				return
			}

			slog.Info("Payload decoded", slog.Any("payload", &decoded))

			ctx := NewParamsContext(r.Context(), decoded)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func ValidateInput[T any](validate *validator.Validate) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("Validating input...")
			ctxVal, params, ok := FromParamsContext[T](r.Context())

			if !ok {
				var t T
				badRequestResponse(w, fmt.Errorf("cannot type assert context value %v to %T", ctxVal, t), message.UserInputInvalid)
				return
			}

			if err := validate.Struct(params); err != nil {
				invalidInputResponse(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(signer security.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr, err := extractBearerToken(r.Header.Get("Authorization"))
			if err != nil || tokenStr == "" {
				unauthorizedResponse(w, err, "Unauthorized")
				return
			}

			sub, err := signer.Verify(tokenStr)
			if err != nil {
				unauthorizedResponse(w, err, "Unauthorized")
				return
			}

			userCtx := NewUserContext(r.Context(), sub)
			r = r.WithContext(userCtx)
			next.ServeHTTP(w, r)
		})
	}
}

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("missing Authorization header")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("missing Bearer prefix")
	}
	return strings.TrimSpace(header[len(prefix):]), nil
}

type SafeResponseWriter struct {
	http.ResponseWriter
	ctx       context.Context
	mu        sync.Mutex
	status    int
	written   bool
	bytesSent int
}

func NewSafeResponseWriter(ctx context.Context, w http.ResponseWriter) *SafeResponseWriter {
	return &SafeResponseWriter{
		ResponseWriter: w,
		ctx:            ctx,
		status:         http.StatusOK,
	}
}

func (w *SafeResponseWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ctx.Err() != nil || w.written {
		return
	}

	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
	w.written = true
}

func (w *SafeResponseWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.ctx.Err() != nil {
		return 0, nil
	}

	if !w.written {
		w.ResponseWriter.WriteHeader(http.StatusOK)
		w.status = http.StatusOK
		w.written = true
	}

	n, err := w.ResponseWriter.Write(b)
	w.bytesSent += n
	return n, err
}

func (w *SafeResponseWriter) Status() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status
}

func (w *SafeResponseWriter) BytesWritten() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.bytesSent
}

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		safeW := NewSafeResponseWriter(r.Context(), w)

		next.ServeHTTP(safeW, r)

		duration := time.Since(start)
		slog.Info("incoming request",
			"user_agent", r.UserAgent(),
			"remote", getIPAddress(r),
			"method", r.Method,
			"url", r.URL.String(),
			"proto", r.Proto,
			slog.Int("status_code", safeW.Status()),
			slog.Int("bytes", safeW.BytesWritten()),
			"duration", duration,
		)
	})
}

// getIPAddress extracts the client's IP address from the request.
func getIPAddress(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if forwardedFor := r.Header.Values("X-Forwarded-For"); len(forwardedFor) > 0 {
		firstIP := forwardedFor[0]
		ips := strings.Split(firstIP, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	return r.RemoteAddr
}
