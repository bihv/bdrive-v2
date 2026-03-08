package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	v := validator.New()

	// Register tag name function to use JSON tag instead of struct field name
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct and returns human-readable errors.
func (v *Validator) Validate(data interface{}) map[string]string {
	errs := v.validate.Struct(data)
	if errs == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range errs.(validator.ValidationErrors) {
		field := err.Field()
		switch err.Tag() {
		case "required":
			errors[field] = fmt.Sprintf("%s is required", field)
		case "email":
			errors[field] = "Invalid email format"
		case "min":
			errors[field] = fmt.Sprintf("%s must be at least %s characters", field, err.Param())
		case "max":
			errors[field] = fmt.Sprintf("%s must be at most %s characters", field, err.Param())
		default:
			errors[field] = fmt.Sprintf("%s is invalid", field)
		}
	}
	return errors
}
