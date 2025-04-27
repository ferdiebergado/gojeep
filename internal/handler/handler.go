package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ferdiebergado/gojeep/internal/config"
	"github.com/ferdiebergado/gojeep/internal/pkg/message"
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
	User UserHandler
}

func New(svc service.Service, signer security.Signer, cfg *config.Config) *Handler {
	return &Handler{
		Base: *NewBaseHandler(svc.Base),
		User: *NewUserHandler(svc.User, signer, cfg),
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

type UserHandler struct {
	service service.UserService
	signer  security.Signer
	cfg     *config.Config
}

func NewUserHandler(userService service.UserService, signer security.Signer, cfg *config.Config) *UserHandler {
	return &UserHandler{
		service: userService,
		signer:  signer,
		cfg:     cfg,
	}
}

type RegisterUserRequest struct {
	Email           string `json:"email,omitempty" validate:"required,email"`
	Password        string `json:"password,omitempty" validate:"required"`
	PasswordConfirm string `json:"password_confirm,omitempty" validate:"required,eqfield=Password"`
}

func (r *RegisterUserRequest) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("email", maskChar),
		slog.String("password", maskChar),
		slog.String("password_confirm", maskChar),
	)
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
		if errors.Is(err, service.ErrUserExists) {
			unprocessableResponse(w, err, message.UserExists)
			return
		}
		response.ServerError(w, err)
		return
	}

	res := Response[*RegisterUserResponse]{
		Message: message.UserRegSuccess,
		Data: &RegisterUserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	response.JSON(w, http.StatusCreated, res)
}

func (h *UserHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if token == "" {
		badRequestResponse(w, service.ErrInvalidToken, message.TokenInvalid)
		return
	}

	if err := h.service.VerifyUser(r.Context(), token); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			badRequestResponse(w, service.ErrInvalidToken, message.TokenInvalid)
			return
		}
		response.ServerError(w, err)
		return
	}

	res := Response[any]{
		Message: message.UserVerifySuccess,
	}
	response.JSON(w, http.StatusOK, res)
}

type UserLoginRequest struct {
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

func (r *UserLoginRequest) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("email", maskChar),
		slog.String("password", maskChar),
	)
}

type UserLoginResponse struct {
	AccessToken string `json:"access_token,omitempty"`
}

func (h *UserHandler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	_, req, _ := FromParamsContext[UserLoginRequest](r.Context())
	params := service.LoginUserParams{
		Email:    req.Email,
		Password: req.Password,
	}
	accessToken, refreshToken, err := h.service.LoginUser(r.Context(), params)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			unauthorizedResponse(w, err, message.UserNotFound)
			return
		}

		if errors.Is(err, service.ErrUserNotVerified) {
			unauthorizedResponse(w, err, message.UserUnverified)
			return
		}

		response.ServerError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	res := Response[*UserLoginResponse]{
		Message: message.UserLoginSuccess,
		Data: &UserLoginResponse{
			AccessToken: accessToken,
		},
	}

	response.JSON(w, http.StatusOK, res)
}

func (h *UserHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		unauthorizedResponse(w, err, "Unauthorized")
		return
	}

	userID, err := h.signer.Verify(cookie.Value)
	if err != nil {
		unauthorizedResponse(w, err, "Unauthorized")
		return
	}

	newAccessToken, err := h.signer.Sign(userID, []string{h.cfg.JWT.Issuer}, time.Duration(h.cfg.JWT.Duration)*time.Minute)
	if err != nil {
		response.ServerError(w, err)
		return
	}

	res := Response[*UserLoginResponse]{
		Message: message.UserLoginSuccess,
		Data: &UserLoginResponse{
			AccessToken: newAccessToken,
		},
	}

	response.JSON(w, http.StatusOK, res)
}

func (h *UserHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("refresh_token")
	// if err == nil {
	//     _ = InvalidateRefreshToken(cookie.Value) // optional: best effort
	// }

	if err != nil {
		response.ServerError(w, err)
		return
	}

	// Expire the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // expire immediately
	})

	res := Response[any]{
		Message: message.UserLogoutSuccess,
	}

	response.JSON(w, http.StatusOK, res)
}
