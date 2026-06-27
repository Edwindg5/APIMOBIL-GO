package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type RecomendacionService struct {
	recomendacionRepository interfaces.RecomendacionRepository
	loteRepository          interfaces.LoteRepository
}

// NewRecomendacionService crea una nueva instancia del servicio
func NewRecomendacionService(
	recomendacionRepository interfaces.RecomendacionRepository,
	loteRepository interfaces.LoteRepository,
) interfaces.RecomendacionService {
	return &RecomendacionService{
		recomendacionRepository: recomendacionRepository,
		loteRepository:          loteRepository,
	}
}

// GetRecomendaciones obtiene las recomendaciones de un lote
func (s *RecomendacionService) GetRecomendaciones(ctx context.Context, loteID, usuarioID int) ([]entities.Recomendacion, error) {
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

	recomendaciones, err := s.recomendacionRepository.GetByLoteID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting recomendaciones: %w", err)
	}

	return recomendaciones, nil
}
