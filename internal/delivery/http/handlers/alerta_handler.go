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

type AlertaHandler struct {
	alertaService interfaces.AlertaService
}

func NewAlertaHandler(alertaService interfaces.AlertaService) *AlertaHandler {
	return &AlertaHandler{alertaService: alertaService}
}

// GetAlertas maneja GET /lotes/{id}/alertas
func (h *AlertaHandler) GetAlertas(w http.ResponseWriter, r *http.Request) {
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

	var atendida *bool
	if v := r.URL.Query().Get("atendida"); v != "" {
		b := v == "true"
		atendida = &b
	}
	nivel := r.URL.Query().Get("severidad")

	alertas, err := h.alertaService.GetAlertas(r.Context(), loteID, userID, atendida, nivel)
	if err != nil {
		switch err.Error() {
		case "lote not found":
			http.Error(w, `{"error": "lote not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("GetAlertas error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alertas)
}

// AtenderAlerta maneja PUT /alertas/{id}/atender
func (h *AlertaHandler) AtenderAlerta(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	alertaID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid alerta id"}`, http.StatusBadRequest)
		return
	}

	alerta, err := h.alertaService.AtenderAlerta(r.Context(), alertaID, userID)
	if err != nil {
		switch err.Error() {
		case "alerta not found":
			http.Error(w, `{"error": "alerta not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("AtenderAlerta error (alerta_id=%d, user_id=%d): %v", alertaID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(alerta)
}
