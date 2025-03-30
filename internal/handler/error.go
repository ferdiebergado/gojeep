package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ferdiebergado/gopherkit/http/response"
)

func badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusBadRequest, err, "Invalid input.")
}

func unprocessableResponse(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusUnprocessableEntity, err, err.Error())
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, err error, msg string) {
	slog.Error("server error", "reason", err, "request", fmt.Sprint(r))
	res := Response[any]{
		Message: msg,
	}
	response.JSON(w, r, status, res)
}
