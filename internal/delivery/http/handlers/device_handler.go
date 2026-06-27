package handlers

import (
	"encoding/json"
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

// NewDeviceHandler crea una nueva instancia del handler
func NewDeviceHandler(deviceService interfaces.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		validator:     validator.New(),
	}
}

// LinkDevice maneja POST /devices/link
func (h *DeviceHandler) LinkDevice(w http.ResponseWriter, r *http.Request) {
	// Obtener user_id del contexto
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

	// Validar request
	if err := h.validator.Struct(req); err != nil {
		http.Error(w, `{"error": "validation failed"}`, http.StatusBadRequest)
		return
	}

	// Ejecutar link device
	response, err := h.deviceService.LinkDevice(
		r.Context(),
		req.ESP32ID,
		req.ProvisioningToken,
		req.LoteName,
		userID,
	)
	if err != nil {
		// Determinar status code según el error
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid or expired provisioning token" {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "device already linked" {
			statusCode = http.StatusConflict
		} else if err.Error() == "provisioning token does not belong to user" {
			statusCode = http.StatusForbidden
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
