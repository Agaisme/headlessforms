// Package validator provides request validation for API handlers
package validator

import (
	"encoding/json"
	"net/http"
	"strings"

	"headless_form/internal/adapter/api/response"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return ve[0].Message
}

// Validator provides request validation
type Validator struct {
	errors ValidationErrors
}

// New creates a new Validator
func New() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Required checks if a string field is not empty
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: field + " is required",
		})
	}
	return v
}

// MinLength checks if a string has minimum length
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if len(value) < min {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: field + " must be at least " + itoa(min) + " characters",
		})
	}
	return v
}

// MaxLength checks if a string doesn't exceed maximum length
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if len(value) > max {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: field + " must not exceed " + itoa(max) + " characters",
		})
	}
	return v
}

// Email validates an email format (simple check)
func (v *Validator) Email(field, value string) *Validator {
	if value != "" && (!strings.Contains(value, "@") || !strings.Contains(value, ".")) {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: field + " must be a valid email address",
		})
	}
	return v
}

// OneOf checks if value is one of allowed values
func (v *Validator) OneOf(field, value string, allowed []string) *Validator {
	if value == "" {
		return v // Skip if empty (use Required for mandatory)
	}
	for _, a := range allowed {
		if value == a {
			return v
		}
	}
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: field + " must be one of: " + strings.Join(allowed, ", "),
	})
	return v
}

// Valid returns true if there are no validation errors
func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// Fail writes validation error response to the http.ResponseWriter
func (v *Validator) Fail(w http.ResponseWriter) {
	response.Error(w, http.StatusBadRequest, v.errors[0].Message, "VALIDATION_ERROR")
}

// FailWithDetails writes validation error response with all error details
func (v *Validator) FailWithDetails(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "fail",
		"message": "Validation failed",
		"errors":  v.errors,
	})
}

// itoa is a simple helper to convert int to string
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var s string
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
