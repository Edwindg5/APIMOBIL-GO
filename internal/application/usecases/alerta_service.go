package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type AlertaService struct {
	alertaRepository interfaces.AlertaRepository
	loteRepository   interfaces.LoteRepository
}

// NewAlertaService crea una nueva instancia del servicio
func NewAlertaService(
	alertaRepository interfaces.AlertaRepository,
	loteRepository interfaces.LoteRepository,
) interfaces.AlertaService {
	return &AlertaService{
		alertaRepository: alertaRepository,
		loteRepository:   loteRepository,
	}
}

// GetAlertas obtiene todas las alertas de un lote
func (s *AlertaService) GetAlertas(ctx context.Context, loteID, usuarioID int) ([]entities.Alerta, error) {
	// Verificar que el lote pertenece al usuario
	lote, err := s.loteRepository.GetByID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting lote: %w", err)
	}

	if lote == nil {
		return nil, errors.New("lote not found")
	}

	if lote.UsuarioID != usuarioID {
		return nil, errors.New("unauthorized")
	}

	alertas, err := s.alertaRepository.GetByLoteID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting alertas: %w", err)
	}

	return alertas, nil
}
