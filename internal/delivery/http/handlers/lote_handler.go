package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type LoteHandler struct {
	loteService interfaces.LoteService
	validator   *validator.Validate
}

// NewLoteHandler crea una nueva instancia del handler
func NewLoteHandler(loteService interfaces.LoteService) *LoteHandler {
	return &LoteHandler{
		loteService: loteService,
		validator:   validator.New(),
	}
}

// GetLotes maneja GET /lotes
func (h *LoteHandler) GetLotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	estado := r.URL.Query().Get("estado")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	result, err := h.loteService.GetLotes(r.Context(), userID, estado, page, limit)
	if err != nil {
		log.Printf("GetLotes error (user_id=%d): %v", userID, err)
		http.Error(w, `{"error": "error fetching lotes"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// GetLote maneja GET /lotes/{id}
func (h *LoteHandler) GetLote(w http.ResponseWriter, r *http.Request) {
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

	detalle, err := h.loteService.GetLoteDetalle(r.Context(), loteID, userID)
	if err != nil {
		switch err.Error() {
		case "lote not found":
			http.Error(w, `{"error": "lote not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("GetLote error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(detalle)
}

// CreateLote maneja POST /lotes
func (h *LoteHandler) CreateLote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req entities.CreateLoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	lote, err := h.loteService.CreateLote(r.Context(), &req, userID)
	if err != nil {
		log.Printf("CreateLote error (user_id=%d): %v", userID, err)
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lote)
}

// UpdateLote maneja PUT /lotes/{id}
func (h *LoteHandler) UpdateLote(w http.ResponseWriter, r *http.Request) {
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

	var req entities.UpdateLoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	lote, err := h.loteService.UpdateLote(r.Context(), loteID, userID, &req)
	if err != nil {
		switch err.Error() {
		case "lote not found or not editable":
			http.Error(w, `{"error": "lote not found or not in process"}`, http.StatusNotFound)
		default:
			log.Printf("UpdateLote error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lote)
}

// FinalizarLote maneja PUT /lotes/{id}/finalizar
func (h *LoteHandler) FinalizarLote(w http.ResponseWriter, r *http.Request) {
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

	lote, err := h.loteService.FinalizarLote(r.Context(), loteID, userID)
	if err != nil {
		switch err.Error() {
		case "lote not found or not in process":
			http.Error(w, `{"error": "lote not found or not in process"}`, http.StatusNotFound)
		default:
			log.Printf("FinalizarLote error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lote)
}

// CancelarLote maneja DELETE /lotes/{id}
func (h *LoteHandler) CancelarLote(w http.ResponseWriter, r *http.Request) {
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

	if err := h.loteService.CancelarLote(r.Context(), loteID, userID); err != nil {
		switch err.Error() {
		case "lote not found or not in process":
			http.Error(w, `{"error": "lote not found or not in process"}`, http.StatusNotFound)
		default:
			log.Printf("CancelarLote error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Lote cancelado"})
}

// ReclamarLote maneja PUT /lotes/reclamar
func (h *LoteHandler) ReclamarLote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req entities.ReclamarLoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	lote, err := h.loteService.ReclamarLote(r.Context(), req.CodigoQR, userID)
	if err != nil {
		switch err.Error() {
		case "codigo qr invalido o ya utilizado":
			http.Error(w, `{"error": "Este código QR no es válido o ya fue utilizado"}`, http.StatusConflict)
		default:
			log.Printf("ReclamarLote error (user_id=%d): %v", userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lote)
}

// GetQR maneja GET /lotes/{id}/qr
func (h *LoteHandler) GetQR(w http.ResponseWriter, r *http.Request) {
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

	detalle, err := h.loteService.GetLoteDetalle(r.Context(), loteID, userID)
	if err != nil {
		switch err.Error() {
		case "lote not found":
			http.Error(w, `{"error": "lote not found"}`, http.StatusNotFound)
		case "unauthorized":
			http.Error(w, `{"error": "unauthorized"}`, http.StatusForbidden)
		default:
			log.Printf("GetQR error (lote_id=%d, user_id=%d): %v", loteID, userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"codigo_qr": detalle.CodigoQR})
}
