package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
)

type LecturaHandler struct {
	lecturaService interfaces.LecturaService
}

func NewLecturaHandler(lecturaService interfaces.LecturaService) *LecturaHandler {
	return &LecturaHandler{lecturaService: lecturaService}
}

// GetLecturas maneja GET /lotes/{id}/lecturas
func (h *LecturaHandler) GetLecturas(w http.ResponseWriter, r *http.Request) {
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

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	var desde time.Time
	if d := r.URL.Query().Get("desde"); d != "" {
		if t, err := time.Parse(time.RFC3339, d); err == nil {
			desde = t
		}
	}

	lecturas, err := h.lecturaService.GetLecturasFiltradas(r.Context(), loteID, userID, limit, desde)
	if err != nil {
		switch err.Error() {
		case "lote not found":
			http.Error(w, `{"error": "lote not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("lectura handler error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lecturas)
}

// GetEstadisticas maneja GET /lotes/{id}/estadisticas
func (h *LecturaHandler) GetEstadisticas(w http.ResponseWriter, r *http.Request) {
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

	stats, err := h.lecturaService.GetEstadisticas(r.Context(), loteID, userID)
	if err != nil {
		switch err.Error() {
		case "lote not found":
			http.Error(w, `{"error": "lote not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("lectura handler error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}
