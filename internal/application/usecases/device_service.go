package usecases

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type DeviceService struct {
	sensorRepository interfaces.SensorRepository
	loteRepository   interfaces.LoteRepository
	historialRepo    interfaces.HistorialRepository
}

func NewDeviceService(
	sensorRepository interfaces.SensorRepository,
	loteRepository interfaces.LoteRepository,
	historialRepo interfaces.HistorialRepository,
) interfaces.DeviceService {
	return &DeviceService{
		sensorRepository: sensorRepository,
		loteRepository:   loteRepository,
		historialRepo:    historialRepo,
	}
}

func (s *DeviceService) LinkDevice(ctx context.Context, esp32ID, provisioningToken string, usuarioID int) (*entities.LinkDeviceResponse, error) {
	sensor, err := s.sensorRepository.GetByIdentifier(ctx, esp32ID)
	if err != nil {
		return nil, fmt.Errorf("error finding sensor: %w", err)
	}
	if sensor == nil {
		return nil, errors.New("sensor not found")
	}

	if sensor.TokenUsado {
		return nil, errors.New("token already used")
	}

	if sensor.ProvisioningToken == nil || *sensor.ProvisioningToken != provisioningToken {
		return nil, errors.New("invalid provisioning token")
	}

	if err := s.sensorRepository.MarcarTokenUsado(ctx, sensor.ID); err != nil {
		return nil, fmt.Errorf("error marking token: %w", err)
	}

	tipoProceso := "lavado"
	lote := &entities.LoteCafe{
		UsuarioID:   usuarioID,
		NombreLote:  fmt.Sprintf("Lote %s", esp32ID),
		Variedad:    nil,
		TipoProceso: &tipoProceso,
		PesoKg:      nil,
		Ubicacion:   nil,
		IDSensor:    &sensor.ID,
		Estado:      "en_proceso",
	}

	created, err := s.loteRepository.Create(ctx, lote)
	if err != nil {
		return nil, fmt.Errorf("error creating lote: %w", err)
	}

	if err := s.sensorRepository.LinkToLote(ctx, sensor.ID, created.ID); err != nil {
		return nil, fmt.Errorf("error linking sensor: %w", err)
	}

	evento := &entities.HistorialEvento{
		LoteID:      created.ID,
		TipoEvento:  "dispositivo_enlazado",
		Descripcion: fmt.Sprintf("Sensor %s enlazado al lote", esp32ID),
	}
	_ = s.historialRepo.Create(ctx, evento)

	return &entities.LinkDeviceResponse{
		Lote:    created,
		Message: "Dispositivo enlazado exitosamente",
	}, nil
}

// GenerateProvisioningToken genera un token de provisioning único
func GenerateProvisioningToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}
