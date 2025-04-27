package handler

import (
	"github.com/ferdiebergado/gojeep/internal/router"
	"github.com/go-playground/validator/v10"
)

func MountRoutes(r router.Router, h *Handler, v *validator.Validate) {
	r.Get("/health", h.Base.HandleHealth)
	r.Group("/auth", func(gr router.Router) router.Router {
		gr.Post("/register", h.Auth.HandleUserRegister,
			DecodeJSON[RegisterUserRequest](), ValidateInput[RegisterUserRequest](v))
		gr.Get("/verify", h.Auth.VerifyEmail)
		gr.Post("/login", h.Auth.HandleUserLogin,
			DecodeJSON[UserLoginRequest](), ValidateInput[UserLoginRequest](v))
		gr.Post("/refresh", h.Auth.HandleRefreshToken)
		gr.Post("/logout", h.Auth.HandleLogout)
		return gr
	})
}
