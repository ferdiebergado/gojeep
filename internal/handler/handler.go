package handler

import (
	"log/slog"
	"net/http"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/security"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/ferdiebergado/gopherkit/http/response"
)

const (
	HeaderContentType = "Content-Type"
	MimeJSON          = "application/json"
	maskChar          = "*"
)

type Response[T any] struct {
	Message string            `json:"message,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
	Data    T                 `json:"data,omitempty"`
}

type Handler struct {
	Base BaseHandler
	Auth AuthHandler
}

func New(svc service.Service, signer security.Signer, cfg *config.Config) *Handler {
	return &Handler{
		Base: *NewBaseHandler(svc.Base),
		Auth: *NewAuthHandler(svc.User, signer, cfg),
	}
}

type BaseHandler struct {
	svc service.BaseService
}

func NewBaseHandler(svc service.BaseService) *BaseHandler {
	return &BaseHandler{svc: svc}
}

func (h *BaseHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	msg := "healthy"

	if err := h.svc.PingDB(r.Context()); err != nil {
		status = http.StatusServiceUnavailable
		msg = "unhealthy"
		slog.Error("failed to connect to the database", "reason", err)
	}

	response.JSON(w, status, Response[any]{Message: msg})
}
