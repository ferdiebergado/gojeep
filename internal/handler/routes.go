package handler

import (
	"github.com/ferdiebergado/goexpress"
	"github.com/go-playground/validator/v10"
)

func mountAPIRoutes(r *goexpress.Router, h *Handler, v *validator.Validate) {
	r.Get("/health", h.Base.HandleHealth)
	r.Group("/auth", func(gr *goexpress.Router) *goexpress.Router {
		gr.Post("/register", h.User.HandleUserRegister,
			DecodeJSON[RegisterUserRequest](), ValidateInput[RegisterUserRequest](v))
		gr.Get("/verify", h.User.VerifyEmail)
		return gr
	})
}
