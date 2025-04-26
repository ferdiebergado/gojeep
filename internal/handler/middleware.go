package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
