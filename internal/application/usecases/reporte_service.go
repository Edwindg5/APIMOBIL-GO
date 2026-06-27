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

// NewReporteService crea una nueva instancia del servicio
func NewReporteService(
	reporteRepository interfaces.ReporteRepository,
	loteRepository interfaces.LoteRepository,
) interfaces.ReporteService {
	return &ReporteService{
		reporteRepository: reporteRepository,
		loteRepository:    loteRepository,
	}
}

// RequestReporte solicita la generación de un reporte
func (s *ReporteService) RequestReporte(ctx context.Context, loteID, usuarioID int, tipoReporte string) (int, error) {
	// Verificar que el lote pertenece al usuario
	lote, err := s.loteRepository.GetByID(ctx, loteID)
	if err != nil {
		return 0, fmt.Errorf("error getting lote: %w", err)
	}

	if lote == nil {
		return 0, errors.New("lote not found")
	}

	if lote.UsuarioID != usuarioID {
		return 0, errors.New("unauthorized")
	}

	// Validar tipo de reporte
	validTypes := map[string]bool{
		"diario":   true,
		"semanal":  true,
		"mensual":  true,
	}

	if !validTypes[tipoReporte] {
		return 0, errors.New("invalid report type")
	}

	// Crear reporte
	reporte := &entities.Reporte{
		LoteID:    loteID,
		UsuarioID: usuarioID,
		Tipo:      tipoReporte,
		Estado:    "pendiente",
	}

	reporteID, err := s.reporteRepository.Create(ctx, reporte)
	if err != nil {
		return 0, fmt.Errorf("error creating reporte: %w", err)
	}

	// TODO: Enviar evento a api-web o cola de mensajes para procesar

	return reporteID, nil
}
