package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
)

type RecomendacionHandler struct {
	recomendacionService interfaces.RecomendacionService
}

// NewRecomendacionHandler crea una nueva instancia del handler
func NewRecomendacionHandler(recomendacionService interfaces.RecomendacionService) *RecomendacionHandler {
	return &RecomendacionHandler{
		recomendacionService: recomendacionService,
	}
}

// GetRecomendaciones maneja GET /lotes/{id}/recomendaciones
func (h *RecomendacionHandler) GetRecomendaciones(w http.ResponseWriter, r *http.Request) {
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

	recomendaciones, err := h.recomendacionService.GetRecomendaciones(r.Context(), loteID, userID)
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
	json.NewEncoder(w).Encode(recomendaciones)
}
