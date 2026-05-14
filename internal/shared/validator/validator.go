package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// FieldError represents a single field validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Get returns the singleton validator instance.
func Get() *validator.Validate {
	once.Do(func() {
		instance = validator.New()
		// Register custom tags here if needed
		_ = instance.RegisterValidation("slug", validateSlug)
	})
	return instance
}

// Validate validates a struct and returns formatted field errors.
func Validate(s interface{}) []FieldError {
	v := Get()
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	var errs []FieldError
	for _, e := range err.(validator.ValidationErrors) {
		errs = append(errs, FieldError{
			Field:   toSnakeCase(e.Field()),
			Message: formatMessage(e),
		})
	}
	return errs
}

func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + e.Param() + " characters"
	case "max":
		return "must be at most " + e.Param() + " characters"
	case "oneof":
		return "must be one of: " + e.Param()
	case "uuid":
		return "must be a valid UUID"
	case "slug":
		return "must be a valid slug (lowercase letters, numbers, hyphens)"
	default:
		return "invalid value"
	}
}

func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return len(slug) > 0
}

func toSnakeCase(s string) string {
	result := make([]rune, 0, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}