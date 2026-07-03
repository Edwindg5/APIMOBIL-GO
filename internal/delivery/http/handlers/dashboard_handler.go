package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
)

type DashboardHandler struct {
	dashboardService interfaces.DashboardService
}

func NewDashboardHandler(dashboardService interfaces.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetDashboard maneja GET /dashboard
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	dashboard, err := h.dashboardService.GetDashboard(r.Context(), userID)
	if err != nil {
		log.Printf("dashboard error (user_id=%d): %v", userID, err)
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dashboard)
}
