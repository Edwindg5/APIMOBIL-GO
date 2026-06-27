package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type LoteService struct {
	loteRepository interfaces.LoteRepository
}

// NewLoteService crea una nueva instancia del servicio
func NewLoteService(loteRepository interfaces.LoteRepository) interfaces.LoteService {
	return &LoteService{
		loteRepository: loteRepository,
	}
}

// GetLotes obtiene todos los lotes del usuario
func (s *LoteService) GetLotes(ctx context.Context, usuarioID int) ([]entities.LoteListItem, error) {
	lotes, err := s.loteRepository.GetByUsuarioID(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error getting lotes: %w", err)
	}

	items := make([]entities.LoteListItem, 0, len(lotes))
	for _, lote := range lotes {
		items = append(items, entities.LoteListItem{
			ID:          lote.ID,
			Nombre:      lote.Nombre,
			Descripcion: lote.Descripcion,
			Area:        lote.Area,
			Estado:      lote.Estado,
			SensorID:    lote.SensorID,
			CreatedAt:   lote.CreatedAt,
			UpdatedAt:   lote.UpdatedAt,
		})
	}

	return items, nil
}

// GetLote obtiene un lote específico
func (s *LoteService) GetLote(ctx context.Context, loteID, usuarioID int) (*entities.LoteCafe, error) {
	lote, err := s.loteRepository.GetByID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting lote: %w", err)
	}

	if lote == nil {
		return nil, errors.New("lote not found")
	}

	// Verificar que el lote pertenece al usuario
	if lote.UsuarioID != usuarioID {
		return nil, errors.New("unauthorized")
	}

	return lote, nil
}

// CreateLote crea un nuevo lote
func (s *LoteService) CreateLote(ctx context.Context, nombre, descripcion string, area float64, usuarioID int) (int, error) {
	if nombre == "" {
		return 0, errors.New("lote name is required")
	}

	if area < 0 {
		return 0, errors.New("area must be non-negative")
	}

	lote := &entities.LoteCafe{
		UsuarioID:   usuarioID,
		Nombre:      nombre,
		Descripcion: descripcion,
		Area:        area,
		Estado:      "activo",
	}

	loteID, err := s.loteRepository.Create(ctx, lote)
	if err != nil {
		return 0, fmt.Errorf("error creating lote: %w", err)
	}

	return loteID, nil
}
