package response

import (
	"encoding/json"
	"errors"
	"headless_form/internal/core/domain"
	"log"
	"net/http"
)

// Envelope is the standard structure for all API responses
type Envelope struct {
	Status  string      `json:"status"`            // "success", "error", "fail"
	Data    interface{} `json:"data,omitempty"`    // Results for success/fail
	Message string      `json:"message,omitempty"` // Error message for errors
	Code    string      `json:"code,omitempty"`    // Internal error code (e.g. "INVALID_EMAIL")
}

// writeJSON encodes data to JSON and writes to response, logging any errors
func writeJSON(w http.ResponseWriter, data interface{}) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ERROR] Failed to encode JSON response: %v", err)
	}
}

// Success sends a 200 OK with the given data
func Success(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, Envelope{
		Status: "success",
		Data:   data,
	})
}

// Created sends a 201 Created with the given data
func Created(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, Envelope{
		Status: "success",
		Data:   data,
	})
}

// Error sends a JSON error response with the specific status code
func Error(w http.ResponseWriter, statusCode int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	writeJSON(w, Envelope{
		Status:  "error",
		Message: message,
		Code:    code,
	})
}

// BadRequest sends a 400 Bad Request
func BadRequest(w http.ResponseWriter, message string, code string) {
	Error(w, http.StatusBadRequest, message, code)
}

// NotFound sends a 404 Not Found
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, "NOT_FOUND")
}

// HandleError checks if there is an error and handles it (Helper for "if err != nil")
func HandleError(w http.ResponseWriter, err error) bool {
	if err != nil {
		// Log the actual error for debugging
		log.Printf("[ERROR] Internal error: %v", err)
		Error(w, http.StatusInternalServerError, "Internal Server Error", "INTERNAL_ERROR")
		return true
	}
	return false
}

// HandleDomainError handles domain errors using Go 1.13+ errors.Is() for proper error matching.
// Returns true if error was handled, false if caller should handle it.
func HandleDomainError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	// Form errors
	if errors.Is(err, domain.ErrFormNotFound) {
		NotFound(w, "Form not found")
		return true
	}
	if errors.Is(err, domain.ErrFormNameRequired) || errors.Is(err, domain.ErrFormNameTooLong) {
		BadRequest(w, err.Error(), "VALIDATION_ERROR")
		return true
	}

	// Submission errors
	if errors.Is(err, domain.ErrSubmissionNotFound) {
		NotFound(w, "Submission not found")
		return true
	}

	// Access control errors
	if errors.Is(err, domain.ErrInvalidSubmissionKey) {
		Error(w, http.StatusForbidden, "Invalid or missing submission key", "INVALID_KEY")
		return true
	}
	if errors.Is(err, domain.ErrAuthRequired) {
		Error(w, http.StatusUnauthorized, "Authentication required for this form", "AUTH_REQUIRED")
		return true
	}

	// User errors
	if errors.Is(err, domain.ErrUserNotFound) {
		NotFound(w, "User not found")
		return true
	}
	if errors.Is(err, domain.ErrUserExists) {
		Error(w, http.StatusConflict, "User already exists", "USER_EXISTS")
		return true
	}
	if errors.Is(err, domain.ErrInvalidCredentials) {
		Error(w, http.StatusUnauthorized, "Invalid credentials", "INVALID_CREDENTIALS")
		return true
	}
	if errors.Is(err, domain.ErrPasswordTooShort) {
		BadRequest(w, "Password must be at least 8 characters", "PASSWORD_TOO_SHORT")
		return true
	}
	if errors.Is(err, domain.ErrEmailRequired) {
		BadRequest(w, "Email is required", "EMAIL_REQUIRED")
		return true
	}
	if errors.Is(err, domain.ErrInvalidResetToken) {
		BadRequest(w, "Invalid or expired reset token", "INVALID_TOKEN")
		return true
	}

	// Not a known domain error - let caller handle or use HandleError
	return false
}
