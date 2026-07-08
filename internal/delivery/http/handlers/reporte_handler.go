package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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
		case "limite de reportes alcanzado":
			http.Error(w, `{"error": "Has alcanzado el limite de 30 reportes generados. Elimina reportes antiguos (DELETE /reportes/{id}) o guardalos en un dispositivo externo (USB) antes de generar uno nuevo."}`, http.StatusConflict)
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

// DescargarReporte maneja GET /reportes/{id}/descargar
func (h *ReporteHandler) DescargarReporte(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}

	path, fileName, err := h.reporteService.DescargarReporte(r.Context(), id, userID)
	if err != nil {
		switch err.Error() {
		case "reporte not found":
			http.Error(w, `{"error": "reporte not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		case "archivo no disponible":
			http.Error(w, `{"error": "el archivo aun no esta disponible, intenta de nuevo en unos segundos"}`, http.StatusConflict)
		default:
			log.Printf("DescargarReporte error (user_id=%d): %v", userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	if _, err := os.Stat(path); err != nil {
		log.Printf("DescargarReporte missing file (id=%d): %v", id, err)
		http.Error(w, `{"error": "archivo no encontrado"}`, http.StatusNotFound)
		return
	}

	contentType := "application/pdf"
	if strings.HasSuffix(fileName, ".xlsx") {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	http.ServeFile(w, r, path)
}

// DeleteReporte maneja DELETE /reportes/{id}
func (h *ReporteHandler) DeleteReporte(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := h.reporteService.DeleteReporte(r.Context(), id, userID); err != nil {
		switch err.Error() {
		case "reporte not found":
			http.Error(w, `{"error": "reporte not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("DeleteReporte error (user_id=%d): %v", userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
