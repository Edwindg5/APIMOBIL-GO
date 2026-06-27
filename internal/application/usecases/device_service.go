package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type DeviceService struct {
	sensorRepository             interfaces.SensorRepository
	loteRepository               interfaces.LoteRepository
	provisioningTokenRepository  interfaces.ProvisioningTokenRepository
}

// NewDeviceService crea una nueva instancia del servicio
func NewDeviceService(
	sensorRepository interfaces.SensorRepository,
	loteRepository interfaces.LoteRepository,
	provisioningTokenRepository interfaces.ProvisioningTokenRepository,
) interfaces.DeviceService {
	return &DeviceService{
		sensorRepository:            sensorRepository,
		loteRepository:              loteRepository,
		provisioningTokenRepository: provisioningTokenRepository,
	}
}

// LinkDevice vincula un ESP32 a un usuario usando un provisioning token
func (s *DeviceService) LinkDevice(
	ctx context.Context,
	esp32ID, provisioningToken, loteName string,
	usuarioID int,
) (*entities.LinkDeviceResponse, error) {
	
	// Verificar que el ESP32 no esté ya vinculado
	existingSensor, err := s.sensorRepository.GetByESP32ID(ctx, esp32ID)
	if err != nil {
		return nil, fmt.Errorf("error checking sensor: %w", err)
	}

	if existingSensor != nil && existingSensor.LinkedAt != nil {
		return nil, errors.New("device already linked")
	}

	// Validar provisioning token
	tokenHash := HashTokenForStorage(provisioningToken)
	provToken, err := s.provisioningTokenRepository.GetByToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("error validating token: %w", err)
	}

	if provToken == nil {
		return nil, errors.New("invalid or expired provisioning token")
	}

	if provToken.UsuarioID != usuarioID {
		return nil, errors.New("provisioning token does not belong to user")
	}

	if provToken.ESP32ID != esp32ID {
		return nil, errors.New("provisioning token ESP32 ID mismatch")
	}

	// Crear o actualizar lote
	loteID, err := s.createOrGetLote(ctx, usuarioID, loteName)
	if err != nil {
		return nil, fmt.Errorf("error creating lote: %w", err)
	}

	// Crear o vincular sensor
	var sensorID int
	if existingSensor != nil {
		// Actualizar sensor existente
		sensorID = existingSensor.ID
		if err := s.sensorRepository.LinkToLote(ctx, sensorID, loteID); err != nil {
			return nil, fmt.Errorf("error linking sensor: %w", err)
		}
	} else {
		// Crear nuevo sensor
		newSensor := &entities.Sensor{
			ESP32ID: esp32ID,
			LoteID:  &loteID,
			Estado:  "activo",
		}
		sensorID, err = s.sensorRepository.Create(ctx, newSensor)
		if err != nil {
			return nil, fmt.Errorf("error creating sensor: %w", err)
		}

		// Vincular sensor al lote
		if err := s.sensorRepository.LinkToLote(ctx, sensorID, loteID); err != nil {
			return nil, fmt.Errorf("error linking sensor: %w", err)
		}
	}

	// Marcar token como usado
	if err := s.provisioningTokenRepository.MarkAsUsed(ctx, provToken.ID); err != nil {
		return nil, fmt.Errorf("error marking token as used: %w", err)
	}

	now := time.Now()

	return &entities.LinkDeviceResponse{
		SensorID: sensorID,
		LoteID:   loteID,
		Message:  "Device linked successfully",
		LinkedAt: now,
	}, nil
}

// createOrGetLote crea o obtiene un lote existente
func (s *DeviceService) createOrGetLote(ctx context.Context, usuarioID int, loteName string) (int, error) {
	// Si ya existe un lote con ese nombre para el usuario, retornarlo
	lotes, err := s.loteRepository.GetByUsuarioID(ctx, usuarioID)
	if err != nil {
		return 0, err
	}

	for _, lote := range lotes {
		if lote.Nombre == loteName {
			return lote.ID, nil
		}
	}

	// Crear nuevo lote
	newLote := &entities.LoteCafe{
		UsuarioID:   usuarioID,
		Nombre:      loteName,
		Descripcion: fmt.Sprintf("Lote %s vinculado automáticamente", loteName),
		Area:        0, // Se puede actualizar después
		Estado:      "activo",
	}

	loteID, err := s.loteRepository.Create(ctx, newLote)
	return loteID, err
}

// GenerateProvisioningToken genera un token de provisioning único
func GenerateProvisioningToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}
