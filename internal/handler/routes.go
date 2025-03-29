package handler

import (
	"github.com/ferdiebergado/goexpress"
	"github.com/go-playground/validator/v10"
)

func mountAPIRoutes(r *goexpress.Router, h *APIHandler, v *validator.Validate) {
	r.Group("/api", func(gr *goexpress.Router) *goexpress.Router {
		gr.Get("/health", h.Base.HandleHealth)
		gr.Post("/auth/register", h.User.HandleUserRegister,
			DecodeJSON[RegisterUserRequest](), ValidateInput[RegisterUserRequest](v))

		return gr
	})
}

func mountRoutes(r *goexpress.Router, h *Handler) {
	r.Get("/dashboard", h.Base.HandleDashboard)
	r.Get("/auth/register", h.User.HandleRegister)
}
