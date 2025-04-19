package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ferdiebergado/gojeep/internal/pkg/message"
	"github.com/go-playground/validator/v10"
)

func DecodeJSON[T any]() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("Checking content-type...")
			contentType := r.Header.Get(HeaderContentType)

			if contentType != MimeJSON {
				badRequestResponse(w, fmt.Errorf("Invalid content-type: %s", contentType), message.UserInputInvalid)
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
