package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type HistorialService struct {
	historialRepository interfaces.HistorialRepository
	loteRepository      interfaces.LoteRepository
}

// NewHistorialService crea una nueva instancia del servicio
func NewHistorialService(
	historialRepository interfaces.HistorialRepository,
	loteRepository interfaces.LoteRepository,
) interfaces.HistorialService {
	return &HistorialService{
		historialRepository: historialRepository,
		loteRepository:      loteRepository,
	}
}

// GetHistorial obtiene el historial de un lote
func (s *HistorialService) GetHistorial(ctx context.Context, loteID, usuarioID int) ([]entities.HistorialEvento, error) {
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

	historial, err := s.historialRepository.GetByLoteID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting historial: %w", err)
	}

	return historial, nil
}
