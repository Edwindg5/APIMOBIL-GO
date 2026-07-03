package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type DashboardService struct {
	loteRepo       interfaces.LoteRepository
	alertaRepo     interfaces.AlertaRepository
	lecturaRepo    interfaces.LecturaRepository
	prediccionRepo interfaces.PrediccionRepository
	reporteRepo    interfaces.ReporteRepository
	sensorRepo     interfaces.SensorRepository
}

func NewDashboardService(
	loteRepo interfaces.LoteRepository,
	alertaRepo interfaces.AlertaRepository,
	lecturaRepo interfaces.LecturaRepository,
	prediccionRepo interfaces.PrediccionRepository,
	reporteRepo interfaces.ReporteRepository,
	sensorRepo interfaces.SensorRepository,
) interfaces.DashboardService {
	return &DashboardService{
		loteRepo:       loteRepo,
		alertaRepo:     alertaRepo,
		lecturaRepo:    lecturaRepo,
		prediccionRepo: prediccionRepo,
		reporteRepo:    reporteRepo,
		sensorRepo:     sensorRepo,
	}
}

func (s *DashboardService) GetDashboard(ctx context.Context, usuarioID int) (*entities.DashboardResponse, error) {
	lotes, _, err := s.loteRepo.GetByUsuarioID(ctx, usuarioID, "", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting lotes: %w", err)
	}

	reportes, err := s.reporteRepo.GetByUsuarioID(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error getting reportes: %w", err)
	}

	totalSensores, err := s.sensorRepo.CountByUsuarioID(ctx, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("error getting sensores: %w", err)
	}

	resp := &entities.DashboardResponse{
		TotalLotes:    len(lotes),
		TotalReportes: len(reportes),
		TotalSensores: totalSensores,
		LotesResumen:  make([]entities.DashboardLoteResumen, 0, len(lotes)),
	}

	var tempSum, humSum float64
	var tempCount int
	var latestPredTime time.Time

	for _, lote := range lotes {
		switch lote.Estado {
		case "en_proceso":
			resp.LotesActivos++
		case "finalizado":
			resp.LotesFinalizados++
		}

		resumen := entities.DashboardLoteResumen{
			IDLote:     lote.ID,
			NombreLote: lote.NombreLote,
			Estado:     lote.Estado,
			DiasSecado: int(time.Since(lote.FechaInicioSecado).Hours() / 24),
		}

		// Última lectura del lote
		lecturas, _ := s.lecturaRepo.GetLatestByLoteID(ctx, lote.ID, 1)
		if len(lecturas) > 0 {
			t := lecturas[0].Temperatura
			h := lecturas[0].Humedad
			resumen.UltimaTemperatura = &t
			resumen.UltimaHumedad = &h
			tempSum += t
			humSum += h
			tempCount++
		}

		// Alertas activas (no atendidas)
		alertas, _ := s.alertaRepo.GetByLoteID(ctx, lote.ID)
		for _, a := range alertas {
			if !a.Atendida {
				resumen.AlertasActivas++
				resp.AlertasSinAtender++
				if a.NivelSeveridad == "critica" {
					resp.AlertasCriticasActivas++
				}
			}
		}

		// Última predicción del lote (buscar la más reciente global)
		preds, _ := s.prediccionRepo.GetByLoteID(ctx, lote.ID)
		if len(preds) > 0 && preds[0].FechaPrediccion.After(latestPredTime) {
			latestPredTime = preds[0].FechaPrediccion
			resp.UltimaPrediccion = &entities.UltimaPrediccionDashboard{
				IDLote:              lote.ID,
				NombreLote:          lote.NombreLote,
				TiempoEstimadoHoras: preds[0].TiempoEstimadoHoras,
				CalidadEstimada:     preds[0].CalidadEstimada,
				FechaPrediccion:     preds[0].FechaPrediccion,
			}
		}

		resp.LotesResumen = append(resp.LotesResumen, resumen)
	}

	if tempCount > 0 {
		resp.TemperaturaPromedioActual = tempSum / float64(tempCount)
		resp.HumedadPromedioActual = humSum / float64(tempCount)
	}

	return resp, nil
}
