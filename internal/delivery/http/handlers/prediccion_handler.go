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

	// Default 15: sin esto, un lote con muchos ciclos de secado acumulados devolvía TODO su
	// historial de predicciones de golpe -- la app móvil lo renderizaba entero en una lista sin
	// paginar y se ponía lenta. Se bajó de 30 a 15 a pedido explícito ("solo los más
	// relevantes"). ?limit= lo puede subir/bajar (tope 200 para no reabrir el mismo problema si
	// algún caller manda un valor absurdo), o pedir explícitamente 0 para "sin límite".
	limit := 15
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			http.Error(w, `{"error": "invalid limit"}`, http.StatusBadRequest)
			return
		}
		if parsed == 0 {
			limit = 0 // sin límite, pedido explícitamente
		} else if parsed > 200 {
			limit = 200
		} else {
			limit = parsed
		}
	}

	predicciones, err := h.prediccionService.GetPredicciones(r.Context(), loteID, userID, limit)
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
