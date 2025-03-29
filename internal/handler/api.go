package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ferdiebergado/gopherkit/http/response"
	"github.com/ferdiebergado/goweb/internal/pkg/message"
	"github.com/ferdiebergado/goweb/internal/service"
)

type APIResponse[T any] struct {
	Message string            `json:"message,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
	Data    T                 `json:"data,omitempty"`
}

type APIHandler struct {
	Base BaseAPIHandler
	User UserAPIHandler
}

func NewAPIHandler(svc service.Service) *APIHandler {
	return &APIHandler{
		Base: *NewBaseAPIHandler(svc.Base),
		User: *NewUserAPIHandler(svc.User),
	}
}

type BaseAPIHandler struct {
	svc service.BaseService
}

func NewBaseAPIHandler(svc service.BaseService) *BaseAPIHandler {
	return &BaseAPIHandler{svc: svc}
}

func (h *BaseAPIHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	msg := "healthy"

	if err := h.svc.PingDB(r.Context()); err != nil {
		status = http.StatusServiceUnavailable
		msg = "unhealthy"
		slog.Error("failed to connect to the database", "reason", err)
	}

	response.JSON(w, r, status, APIResponse[any]{Message: msg})
}

type UserAPIHandler struct {
	service service.UserService
}

func NewUserAPIHandler(userService service.UserService) *UserAPIHandler {
	return &UserAPIHandler{
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

func (h *UserAPIHandler) HandleUserRegister(w http.ResponseWriter, r *http.Request) {
	_, req, _ := FromParamsContext[RegisterUserRequest](r.Context())
	params := service.RegisterUserParams{
		Email:    req.Email,
		Password: req.Password,
	}
	user, err := h.service.RegisterUser(r.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateUser) {
			unprocessableError(w, r, err)
			return
		}
		response.ServerError(w, r, err)
		return
	}

	res := APIResponse[*RegisterUserResponse]{
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
