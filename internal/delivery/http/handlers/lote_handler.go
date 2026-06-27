package handlers

import (
	"encoding/json"
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

	lotes, err := h.loteService.GetLotes(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "error fetching lotes"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lotes)
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

	lote, err := h.loteService.GetLote(r.Context(), loteID, userID)
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
	json.NewEncoder(w).Encode(lote)
}

// CreateLote maneja POST /lotes
func (h *LoteHandler) CreateLote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	type CreateLoteRequest struct {
		Nombre      string  `json:"nombre" validate:"required"`
		Descripcion string  `json:"descripcion"`
		Area        float64 `json:"area" validate:"required,min=0"`
	}

	var req CreateLoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	loteID, err := h.loteService.CreateLote(r.Context(), req.Nombre, req.Descripcion, req.Area, userID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": loteID})
}
