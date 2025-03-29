package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ferdiebergado/gojeep/internal/pkg/message"
	"github.com/ferdiebergado/gojeep/internal/service"
	"github.com/ferdiebergado/gopherkit/http/response"
)

const (
	HeaderContentType = "Content-Type"
	MimeJSON          = "application/json"
)

type Response[T any] struct {
	Message string            `json:"message,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
	Data    T                 `json:"data,omitempty"`
}

type Handler struct {
	Base BaseHandler
	User UserHandler
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{
		Base: *NewBaseHandler(svc.Base),
		User: *NewUserHandler(svc.User),
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

	response.JSON(w, r, status, Response[any]{Message: msg})
}

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		service: userService,
	}
}

type RegisterUserRequest struct {
	Email           string `json:"email,omitempty" validate:"required,email"`
	Password        string `json:"password,omitempty" validate:"required"`
	PasswordConfirm string `json:"password_confirm,omitempty" validate:"required,eqfield=Password"`
}

type RegisterUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *UserHandler) HandleUserRegister(w http.ResponseWriter, r *http.Request) {
	_, req, _ := FromParamsContext[RegisterUserRequest](r.Context())
	params := service.RegisterUserParams{
		Email:    req.Email,
		Password: req.Password,
	}
	user, err := h.service.RegisterUser(r.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateUser) {
			unprocessableResponse(w, r, err)
			return
		}
		response.ServerError(w, r, err)
		return
	}

	res := Response[*RegisterUserResponse]{
		Message: message.Get("regSuccess"),
		Data: &RegisterUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	response.JSON(w, r, http.StatusCreated, res)
}
