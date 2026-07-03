package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type ReporteService struct {
	reporteRepository interfaces.ReporteRepository
	loteRepository    interfaces.LoteRepository
}

func NewReporteService(
	reporteRepository interfaces.ReporteRepository,
	loteRepository interfaces.LoteRepository,
) interfaces.ReporteService {
	return &ReporteService{
		reporteRepository: reporteRepository,
		loteRepository:    loteRepository,
	}
}

func (s *ReporteService) RequestReporte(ctx context.Context, req *entities.SolicitudReporteRequest, usuarioID int) (*entities.Reporte, error) {
	lote, err := s.loteRepository.GetByID(ctx, req.IDLote)
	if err != nil {
		return nil, fmt.Errorf("error getting lote: %w", err)
	}
	if lote == nil {
		return nil, errors.New("lote not found")
	}
	if lote.UsuarioID != usuarioID {
		return nil, errors.New("unauthorized")
	}

	reporte := &entities.Reporte{
		LoteID:      req.IDLote,
		UsuarioID:   usuarioID,
		TipoReporte: req.TipoReporte,
		Formato:     req.Formato,
	}

	if err := s.reporteRepository.Create(ctx, reporte); err != nil {
		return nil, fmt.Errorf("error creating reporte: %w", err)
	}
	return reporte, nil
}

func (s *ReporteService) GetReportes(ctx context.Context, usuarioID int) ([]entities.Reporte, error) {
	reportes, err := s.reporteRepository.GetByUsuarioID(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error getting reportes: %w", err)
	}
	return reportes, nil
}
