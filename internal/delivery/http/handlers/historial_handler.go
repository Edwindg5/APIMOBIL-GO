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

type HistorialHandler struct {
	historialService interfaces.HistorialService
}

func NewHistorialHandler(historialService interfaces.HistorialService) *HistorialHandler {
	return &HistorialHandler{historialService: historialService}
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

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	eventos, total, err := h.historialService.GetHistorial(r.Context(), loteID, userID, page, limit)
	if err != nil {
		switch err.Error() {
		case "lote not found":
			http.Error(w, `{"error": "lote not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("GetHistorial error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data":  eventos,
		"total": total,
	})
}
