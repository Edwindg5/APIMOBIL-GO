package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type LecturaService struct {
	lecturaRepository interfaces.LecturaRepository
	loteRepository    interfaces.LoteRepository
}

// NewLecturaService crea una nueva instancia del servicio
func NewLecturaService(
	lecturaRepository interfaces.LecturaRepository,
	loteRepository interfaces.LoteRepository,
) interfaces.LecturaService {
	return &LecturaService{
		lecturaRepository: lecturaRepository,
		loteRepository:    loteRepository,
	}
}

// GetLatestLecturas obtiene las últimas lecturas de un lote
func (s *LecturaService) GetLatestLecturas(ctx context.Context, loteID, usuarioID int, limit int) ([]entities.LecturaAmbiental, error) {
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

	lecturas, err := s.lecturaRepository.GetLatestByLoteID(ctx, loteID, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting lecturas: %w", err)
	}

	return lecturas, nil
}
