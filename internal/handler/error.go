package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/ferdiebergado/gopherkit/http/response"
)

func badRequestResponse(w http.ResponseWriter, err error, msg string) {
	errorResponse(w, http.StatusBadRequest, err, msg)
}

func unprocessableResponse(w http.ResponseWriter, err error, msg string) {
	errorResponse(w, http.StatusUnprocessableEntity, err, msg)
}

func unauthorizedResponse(w http.ResponseWriter, err error, msg string) {
	errorResponse(w, http.StatusUnauthorized, err, msg)
}

func unsupportedContentTypeResponse(w http.ResponseWriter, err error, msg string) {
	errorResponse(w, http.StatusUnsupportedMediaType, err, msg)
}

func errorResponse(w http.ResponseWriter, status int, err error, msg string) {
	slog.Error("Client error", "reason", err)
	res := Response[any]{
		Message: msg,
	}
	response.JSON(w, status, res)
}

func isContextError(err error) bool {
	if errors.Is(err, context.Canceled) {
		slog.Debug("Request cancelled")
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		slog.Debug("Request deadline exceeded")
		return true
	}

	return false
}
