package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func New() *validator.Validate {
	validate := validator.New()

	// register function to get tag name from json tags.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return validate
}
