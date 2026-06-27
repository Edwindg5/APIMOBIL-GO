package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
)

type HistorialHandler struct {
	historialService interfaces.HistorialService
}

// NewHistorialHandler crea una nueva instancia del handler
func NewHistorialHandler(historialService interfaces.HistorialService) *HistorialHandler {
	return &HistorialHandler{
		historialService: historialService,
	}
}

// GetHistorial maneja GET /lotes/{id}/historial
func (h *HistorialHandler) GetHistorial(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	loteID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid lote id"}`, http.StatusBadRequest)
		return
	}

	historial, err := h.historialService.GetHistorial(r.Context(), loteID, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "unauthorized" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "lote not found" {
			statusCode = http.StatusNotFound
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(historial)
}
