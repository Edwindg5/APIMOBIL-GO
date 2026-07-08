package usecases

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
	"github.com/kajve/api-mobile/internal/infrastructure/reportgen"
)

type ReporteService struct {
	reporteRepository interfaces.ReporteRepository
	loteRepository    interfaces.LoteRepository
	usuarioRepository interfaces.UsuarioRepository
	generator         *reportgen.Generator
	reportsDir        string
}

func NewReporteService(
	reporteRepository interfaces.ReporteRepository,
	loteRepository interfaces.LoteRepository,
	usuarioRepository interfaces.UsuarioRepository,
	generator *reportgen.Generator,
	reportsDir string,
) interfaces.ReporteService {
	return &ReporteService{
		reporteRepository: reporteRepository,
		loteRepository:    loteRepository,
		usuarioRepository: usuarioRepository,
		generator:         generator,
		reportsDir:        reportsDir,
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

	usuarioNombre := ""
	if usuario, err := s.usuarioRepository.GetByID(ctx, usuarioID); err == nil && usuario != nil {
		usuarioNombre = usuario.Nombre
	}

	// La generación del archivo corre en background: la solicitud queda
	// registrada de inmediato y el archivo queda disponible poco después
	// (el cliente hace polling de GET /reportes hasta ver url_archivo).
	go s.generarEnBackground(reporte.ID, lote, usuarioNombre, reporte.TipoReporte, reporte.Formato)

	return reporte, nil
}

func (s *ReporteService) generarEnBackground(reporteID int, lote *entities.LoteCafe, usuarioNombre, tipoReporte, formato string) {
	ctx := context.Background()

	if _, err := s.generator.Generate(ctx, reporteID, lote, usuarioNombre, tipoReporte, formato); err != nil {
		log.Printf("error generando reporte %d: %v", reporteID, err)
		return
	}

	descargaURL := fmt.Sprintf("/reportes/%d/descargar", reporteID)
	if err := s.reporteRepository.UpdateURLArchivo(ctx, reporteID, descargaURL); err != nil {
		log.Printf("error actualizando url_archivo del reporte %d: %v", reporteID, err)
	}
}

func (s *ReporteService) GetReportes(ctx context.Context, usuarioID int) ([]entities.Reporte, error) {
	reportes, err := s.reporteRepository.GetByUsuarioID(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error getting reportes: %w", err)
	}
	return reportes, nil
}

func (s *ReporteService) DescargarReporte(ctx context.Context, id, usuarioID int) (string, string, error) {
	reporte, err := s.reporteRepository.GetByID(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("error getting reporte: %w", err)
	}
	if reporte == nil {
		return "", "", errors.New("reporte not found")
	}
	if reporte.UsuarioID != usuarioID {
		return "", "", errors.New("unauthorized")
	}
	if reporte.URLArchivo == nil {
		return "", "", errors.New("archivo no disponible")
	}

	fileName := reportgen.FileName(reporte.ID, reporte.Formato)
	return filepath.Join(s.reportsDir, fileName), fileName, nil
}
