package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error("failed to encode JSON response", "error", err)
		}
	}
}

// ErrorResponse represents a standard error response.
type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Error writes an error JSON response with the given status code and message.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}

// ValidationError writes a 400 validation error with field-level details.
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusBadRequest, ErrorResponse{
		Error:  "validation failed",
		Fields: fields,
	})
}
