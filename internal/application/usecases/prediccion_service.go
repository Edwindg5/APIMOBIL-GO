package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type PrediccionService struct {
	prediccionRepository interfaces.PrediccionRepository
	loteRepository       interfaces.LoteRepository
}

// NewPrediccionService crea una nueva instancia del servicio
func NewPrediccionService(
	prediccionRepository interfaces.PrediccionRepository,
	loteRepository interfaces.LoteRepository,
) interfaces.PrediccionService {
	return &PrediccionService{
		prediccionRepository: prediccionRepository,
		loteRepository:       loteRepository,
	}
}

// GetPredicciones obtiene las predicciones de un lote
func (s *PrediccionService) GetPredicciones(ctx context.Context, loteID, usuarioID int) ([]entities.Prediccion, error) {
	// Verificar que el lote pertenece al usuario
	lote, err := s.loteRepository.GetByID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting lote: %w", err)
	}

	if lote == nil {
		return nil, errors.New("lote not found")
	}

	log.Printf("DEBUG ownership check: usuarioID=%d lote.UsuarioID=%d loteID=%d", usuarioID, lote.UsuarioID, loteID)
	if lote.UsuarioID != usuarioID {
		return nil, errors.New("unauthorized")
	}

	predicciones, err := s.prediccionRepository.GetByLoteID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting predicciones: %w", err)
	}

	return predicciones, nil
}
