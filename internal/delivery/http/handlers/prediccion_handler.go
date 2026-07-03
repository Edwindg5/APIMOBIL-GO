package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
)

type PrediccionHandler struct {
	prediccionService interfaces.PrediccionService
}

// NewPrediccionHandler crea una nueva instancia del handler
func NewPrediccionHandler(prediccionService interfaces.PrediccionService) *PrediccionHandler {
	return &PrediccionHandler{
		prediccionService: prediccionService,
	}
}

// GetPredicciones maneja GET /lotes/{id}/predicciones
func (h *PrediccionHandler) GetPredicciones(w http.ResponseWriter, r *http.Request) {
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

	predicciones, err := h.prediccionService.GetPredicciones(r.Context(), loteID, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "unauthorized" {
			statusCode = http.StatusForbidden
		} else if err.Error() == "lote not found" {
			statusCode = http.StatusNotFound
		} else {
			log.Printf("GetPredicciones error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
		}
		http.Error(w, `{"error": "`+err.Error()+`"}`, statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(predicciones)
}
