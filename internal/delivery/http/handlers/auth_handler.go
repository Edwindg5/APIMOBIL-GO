package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type AuthHandler struct {
	authService     interfaces.AuthService
	registerService interfaces.RegisterService
	validator       *validator.Validate
}

// NewAuthHandler crea una nueva instancia del handler
func NewAuthHandler(authService interfaces.AuthService, registerService interfaces.RegisterService) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		registerService: registerService,
		validator:       validator.New(),
	}
}

// Login maneja POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req entities.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() != "invalid email or password" {
			log.Printf("Login error (email=%s): %v", req.Email, err)
		}
		http.Error(w, `{"error": "invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Refresh maneja POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req entities.RefreshTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	response, err := h.authService.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		log.Printf("Refresh error: %v", err)
		http.Error(w, `{"error": "invalid refresh token"}`, http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Register maneja POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req entities.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	response, err := h.registerService.Register(r.Context(), req.Nombre, req.Email, req.Password, req.Telefono)
	if err != nil {
		if errors.Is(err, errEmailExists) || err.Error() == "email already exists" {
			http.Error(w, `{"error": "email already registered"}`, http.StatusConflict)
			return
		}
		log.Printf("Register error (email=%s): %v", req.Email, err)
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// errEmailExists es el sentinel para detección de email duplicado
var errEmailExists = errors.New("email already exists")
