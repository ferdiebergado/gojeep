package handler

import (
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

func errorResponse(w http.ResponseWriter, status int, err error, msg string) {
	slog.Error("Client error", "reason", err)
	res := Response[any]{
		Message: msg,
	}
	response.JSON(w, status, res)
}
