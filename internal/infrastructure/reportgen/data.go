package reportgen

import (
	"context"
	"time"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

// ReportData agrupa toda la información de un lote necesaria para construir
// un reporte, tanto en PDF como en Excel, sin que los builders conozcan los
// repositorios de la base de datos.
type ReportData struct {
	TipoReporte   string
	Formato       string
	GeneradoEn    time.Time
	UsuarioNombre string

	Lote            entities.LoteCafe
	Estadisticas    *entities.EstadisticasLote
	Lecturas        []entities.LecturaAmbiental // orden cronológico ascendente
	Alertas         []entities.Alerta
	Predicciones    []entities.Prediccion
	Recomendaciones []entities.Recomendacion
	Historial       []entities.HistorialEvento
}

// Collector reúne los datos de un lote desde los distintos repositorios
// para alimentar la generación del reporte.
type Collector struct {
	lecturaRepo       interfaces.LecturaRepository
	alertaRepo        interfaces.AlertaRepository
	prediccionRepo    interfaces.PrediccionRepository
	recomendacionRepo interfaces.RecomendacionRepository
	historialRepo     interfaces.HistorialRepository
}

func NewCollector(
	lecturaRepo interfaces.LecturaRepository,
	alertaRepo interfaces.AlertaRepository,
	prediccionRepo interfaces.PrediccionRepository,
	recomendacionRepo interfaces.RecomendacionRepository,
	historialRepo interfaces.HistorialRepository,
) *Collector {
	return &Collector{
		lecturaRepo:       lecturaRepo,
		alertaRepo:        alertaRepo,
		prediccionRepo:    prediccionRepo,
		recomendacionRepo: recomendacionRepo,
		historialRepo:     historialRepo,
	}
}

func (c *Collector) Collect(ctx context.Context, lote *entities.LoteCafe, usuarioNombre, tipoReporte, formato string) (*ReportData, error) {
	estadisticas, err := c.lecturaRepo.GetEstadisticas(ctx, lote.ID)
	if err != nil {
		return nil, err
	}

	lecturas, err := c.lecturaRepo.GetByLoteIDFiltered(ctx, lote.ID, 500, time.Time{})
	if err != nil {
		return nil, err
	}
	// La consulta regresa orden DESC (más reciente primero); el reporte
	// se lee mejor en orden cronológico ascendente.
	for i, j := 0, len(lecturas)-1; i < j; i, j = i+1, j-1 {
		lecturas[i], lecturas[j] = lecturas[j], lecturas[i]
	}

	alertas, err := c.alertaRepo.GetByLoteID(ctx, lote.ID)
	if err != nil {
		return nil, err
	}

	// El reporte PDF/Excel sí quiere el historial completo de predicciones del lote (a diferencia
	// de la lista de la app móvil, que ahora se limita por default -- ver prediccion_handler.go).
	predicciones, err := c.prediccionRepo.GetByLoteID(ctx, lote.ID, 0)
	if err != nil {
		return nil, err
	}

	recomendaciones, err := c.recomendacionRepo.GetByLoteID(ctx, lote.ID)
	if err != nil {
		return nil, err
	}

	historial, err := c.historialRepo.GetByLoteID(ctx, lote.ID)
	if err != nil {
		return nil, err
	}

	return &ReportData{
		TipoReporte:     tipoReporte,
		Formato:         formato,
		GeneradoEn:      time.Now(),
		UsuarioNombre:   usuarioNombre,
		Lote:            *lote,
		Estadisticas:    estadisticas,
		Lecturas:        lecturas,
		Alertas:         alertas,
		Predicciones:    predicciones,
		Recomendaciones: recomendaciones,
		Historial:       historial,
	}, nil
}
