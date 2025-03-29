package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ferdiebergado/gopherkit/http/response"
)

func badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusBadRequest, err, "Invalid input.")

}

func unprocessableError(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusUnprocessableEntity, err, err.Error())
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, err error, msg string) {
	slog.Error("server error", "reason", err, "request", fmt.Sprint(r))

	if r.Header.Get(HeaderContentType) == MimeJSONUTF8 {
		res := APIResponse[any]{
			Message: msg,
		}
		response.JSON(w, r, status, res)
		return
	}
	http.Error(w, http.StatusText(status), status)
}
