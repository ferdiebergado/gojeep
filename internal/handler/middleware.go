package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ferdiebergado/goexpress"
	"github.com/go-playground/validator/v10"
)

func DecodeJSON[T any]() goexpress.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("Checking content-type for application/json...")
			if r.Header.Get(HeaderContentType) == MimeJSONUTF8 {
				slog.Info("Decoding json body...")
				var decoded T
				decoder := json.NewDecoder(r.Body)
				decoder.DisallowUnknownFields()
				if err := decoder.Decode(&decoded); err != nil {
					badRequestError(w, r, err)
					return
				}
				ctx := NewParamsContext(r.Context(), decoded)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ValidateInput[T any](validate *validator.Validate) goexpress.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("Validating input...")
			ctxVal, params, ok := FromParamsContext[T](r.Context())

			if !ok {
				var t T
				badRequestError(w, r, fmt.Errorf("cannot type assert context value %v to %T", ctxVal, t))
				return
			}

			if err := validate.Struct(params); err != nil {
				validationError(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
