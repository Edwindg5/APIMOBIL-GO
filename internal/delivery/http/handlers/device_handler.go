package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/delivery/http/middleware"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type DeviceHandler struct {
	deviceService interfaces.DeviceService
	validator     *validator.Validate
}

func NewDeviceHandler(deviceService interfaces.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		validator:     validator.New(),
	}
}

// LinkDevice maneja POST /devices/link
func (h *DeviceHandler) LinkDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req entities.LinkDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	response, err := h.deviceService.LinkDevice(r.Context(), req.ESP32ID, req.ProvisioningToken, userID)
	if err != nil {
		switch err.Error() {
		case "token already used":
			http.Error(w, `{"error": "token already used"}`, http.StatusConflict)
		case "invalid provisioning token":
			http.Error(w, `{"error": "invalid provisioning token"}`, http.StatusUnauthorized)
		case "sensor not found":
			http.Error(w, `{"error": "sensor not found"}`, http.StatusNotFound)
		default:
			log.Printf("LinkDevice error (user_id=%d): %v", userID, err)
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
