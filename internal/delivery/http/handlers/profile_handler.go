package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	httpmiddleware "github.com/kajve/api-mobile/internal/delivery/http/middleware"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type ProfileHandler struct {
	profileService interfaces.ProfileService
	validator      *validator.Validate
}

// NewProfileHandler crea una nueva instancia del handler de perfil
func NewProfileHandler(profileService interfaces.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		validator:      validator.New(),
	}
}

// GetPerfil maneja GET /perfil
func (h *ProfileHandler) GetPerfil(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	perfil, err := h.profileService.GetProfile(r.Context(), userID)
	if err != nil {
		log.Printf("GetPerfil error (user_id=%d): %v", userID, err)
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(perfil)
}

// UpdatePerfil maneja PUT /perfil
func (h *ProfileHandler) UpdatePerfil(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req entities.UpdatePerfilRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	perfil, err := h.profileService.UpdateProfile(r.Context(), userID, req.Nombre, req.Telefono)
	if err != nil {
		log.Printf("UpdatePerfil error (user_id=%d): %v", userID, err)
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(perfil)
}

// ChangePassword maneja PUT /perfil/password
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpmiddleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req entities.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	err := h.profileService.ChangePassword(r.Context(), userID, req.PasswordActual, req.PasswordNueva)
	if err != nil {
		if errors.Is(err, errInvalidPassword) || err.Error() == "invalid current password" {
			http.Error(w, `{"error": "contraseña actual incorrecta"}`, http.StatusUnauthorized)
			return
		}
		log.Printf("ChangePassword error (user_id=%d): %v", userID, err)
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Contraseña actualizada"})
}

var errInvalidPassword = errors.New("invalid current password")
