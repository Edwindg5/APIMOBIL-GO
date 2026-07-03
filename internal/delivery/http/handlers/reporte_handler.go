package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type ReporteHandler struct {
	reporteService interfaces.ReporteService
	validator      *validator.Validate
}

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

	var req entities.SolicitudReporteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	reporte, err := h.reporteService.RequestReporte(r.Context(), &req, userID)
	if err != nil {
		switch err.Error() {
		case "lote not found":
			http.Error(w, `{"error": "lote not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("RequestReporte error (user_id=%d): %v", userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reporte)
}

// GetReportes maneja GET /reportes
func (h *ReporteHandler) GetReportes(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	reportes, err := h.reporteService.GetReportes(r.Context(), userID)
	if err != nil {
		log.Printf("GetReportes error (user_id=%d): %v", userID, err)
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reportes)
}
