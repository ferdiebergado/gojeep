package handler

import (
	"log/slog"
	"net/http"

	"github.com/ferdiebergado/gopherkit/http/response"
)

func badRequestResponse(w http.ResponseWriter, err error) {
	errorResponse(w, http.StatusBadRequest, err, "Invalid input.")
}

func unprocessableResponse(w http.ResponseWriter, err error) {
	errorResponse(w, http.StatusUnprocessableEntity, err, err.Error())
}

func unauthorizedResponse(w http.ResponseWriter, err error) {
	errorResponse(w, http.StatusUnauthorized, err, err.Error())
}

func errorResponse(w http.ResponseWriter, status int, err error, msg string) {
	slog.Error("client error", "reason", err)
	res := Response[any]{
		Message: msg,
	}
	response.JSON(w, status, res)
}
