package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard structure for all API responses
type Envelope struct {
	Status  string      `json:"status"`            // "success", "error", "fail"
	Data    interface{} `json:"data,omitempty"`    // Results for success/fail
	Message string      `json:"message,omitempty"` // Error message for errors
	Code    string      `json:"code,omitempty"`    // Internal error code (e.g. "INVALID_EMAIL")
}

// Success sends a 200 OK with the given data
func Success(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Envelope{
		Status: "success",
		Data:   data,
	})
}

// Created sends a 201 Created with the given data
func Created(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Envelope{
		Status: "success",
		Data:   data,
	})
}

// Error sends a JSON error response with the specific status code
func Error(w http.ResponseWriter, statusCode int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Envelope{
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

// Check if there is an error and handle it (Helper for "if err != nil")
func HandleError(w http.ResponseWriter, err error) bool {
	if err != nil {
		// Log the error internally here if needed
		Error(w, http.StatusInternalServerError, "Internal Server Error", "INTERNAL_ERROR")
		return true
	}
	return false
}
