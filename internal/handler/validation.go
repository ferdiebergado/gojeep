package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ferdiebergado/gojeep/internal/pkg/message"
	"github.com/ferdiebergado/gopherkit/http/response"
	"github.com/go-playground/validator/v10"
)

func invalidInputResponse(w http.ResponseWriter, err error) {
	errs := make(map[string]string, 0)
	valErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		badRequestResponse(w, fmt.Errorf("type assert: %w", errors.New("error is not a validator.ValidationErrors type")), message.UserInputInvalid)
		return
	}

	for _, e := range valErrs {
		errs[e.Field()] = validationMessage(e)
	}

	res := Response[any]{
		Message: "Invalid input.",
		Errors:  errs,
	}

	response.JSON(w, http.StatusBadRequest, res)
}

func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", e.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", e.Field(), e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", e.Field(), e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", e.Field(), e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", e.Field(), e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", e.Field(), e.Param())
	case "numeric":
		return fmt.Sprintf("%s must be a number", e.Field())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", e.Field())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", e.Field())
	case "eqfield":
		return fmt.Sprintf("%s should match %s", e.Field(), e.Param())
	default:
		return fmt.Sprintf("%s is invalid", e.Field())
	}
}
