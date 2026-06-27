package utils

import (
	"encoding/json"
	"net/http"
)

// APIResponse es la estructura estándar de respuesta
type APIResponse struct {
	Success bool        `json:"success"`
	Data    any         `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Status  int         `json:"status"`
}

// SuccessResponse envía una respuesta exitosa
func SuccessResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: true,
		Data:    data,
		Status:  statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

// ErrorResponse envía una respuesta de error
func ErrorResponse(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: false,
		Error:   errorMsg,
		Status:  statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

// ValidationError representa un error de validación
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorsResponse envía errores de validación
func ValidationErrorsResponse(w http.ResponseWriter, errors []ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]any{
		"success": false,
		"error":   "validation failed",
		"errors":  errors,
		"status":  http.StatusBadRequest,
	}

	json.NewEncoder(w).Encode(response)
}
