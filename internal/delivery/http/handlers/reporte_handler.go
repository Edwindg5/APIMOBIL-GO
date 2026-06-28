package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
)

type ReporteHandler struct {
	reporteService interfaces.ReporteService
	validator      *validator.Validate
}

// NewReporteHandler crea una nueva instancia del handler
func NewReporteHandler(reporteService interfaces.ReporteService) *ReporteHandler {
	return &ReporteHandler{
		reporteService: reporteService,
		validator:      validator.New(),
	}
}

// RequestReporte maneja POST /reportes
func (h *ReporteHandler) RequestReporte(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	type RequestReporteRequest struct {
		LoteID int    `json:"lote_id" validate:"required"`
		Tipo   string `json:"tipo" validate:"required,oneof=diario semanal mensual"`
	}

	var req RequestReporteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	reporteID, err := h.reporteService.RequestReporte(r.Context(), req.LoteID, userID, req.Tipo)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "unauthorized" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "lote not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "invalid report type" {
			statusCode = http.StatusBadRequest
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, statusCode)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": reporteID})
}

// GetReporte maneja GET /reportes/{id}
func (h *ReporteHandler) GetReporte(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Esta función es un placeholder para futura implementación
	_ = userID

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "endpoint not yet implemented"})
}
