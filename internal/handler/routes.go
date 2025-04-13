package handler

import (
	"github.com/go-playground/validator/v10"
)

func mountAPIRoutes(r Router, h *Handler, v *validator.Validate) {
	r.Get("/health", h.Base.HandleHealth)
	r.Group("/auth", func(gr Router) Router {
		gr.Post("/register", h.User.HandleUserRegister,
			DecodeJSON[RegisterUserRequest](), ValidateInput[RegisterUserRequest](v))
		gr.Get("/verify", h.User.VerifyEmail)
		gr.Post("/login", h.User.HandleUserLogin,
			DecodeJSON[UserLoginRequest](), ValidateInput[UserLoginRequest](v))
		return gr
	})
}
