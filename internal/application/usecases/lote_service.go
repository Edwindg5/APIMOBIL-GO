package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
	"github.com/kajve/api-mobile/internal/infrastructure/mll"
)

type LoteService struct {
	loteRepo             interfaces.LoteRepository
	historialRepo        interfaces.HistorialRepository
	lecturaRepo          interfaces.LecturaRepository
	alertaRepo           interfaces.AlertaRepository
	prediccionRepo       interfaces.PrediccionRepository
	mllClient            *mll.Client
	placeholderUsuarioID int
}

// NewLoteService crea una nueva instancia del servicio
func NewLoteService(
	loteRepo interfaces.LoteRepository,
	historialRepo interfaces.HistorialRepository,
	lecturaRepo interfaces.LecturaRepository,
	alertaRepo interfaces.AlertaRepository,
	prediccionRepo interfaces.PrediccionRepository,
	mllClient *mll.Client,
	placeholderUsuarioID int,
) interfaces.LoteService {
	return &LoteService{
		loteRepo:             loteRepo,
		historialRepo:        historialRepo,
		lecturaRepo:          lecturaRepo,
		alertaRepo:           alertaRepo,
		prediccionRepo:       prediccionRepo,
		mllClient:            mllClient,
		placeholderUsuarioID: placeholderUsuarioID,
	}
}

// GetLotes lista lotes del usuario con filtro opcional de estado y paginación
func (s *LoteService) GetLotes(ctx context.Context, usuarioID int, estado string, page, limit int) (*entities.LotesListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	lotes, total, err := s.loteRepo.GetByUsuarioID(ctx, usuarioID, estado, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error getting lotes: %w", err)
	}

	items := make([]entities.LoteListItem, 0, len(lotes))
	for _, l := range lotes {
		items = append(items, entities.LoteListItem{
			ID:                l.ID,
			NombreLote:        l.NombreLote,
			Variedad:          l.Variedad,
			TipoProceso:       l.TipoProceso,
			PesoKg:            l.PesoKg,
			Ubicacion:         l.Ubicacion,
			IDSensor:          l.IDSensor,
			CodigoQR:          l.CodigoQR,
			Estado:            l.Estado,
			FechaInicioSecado: l.FechaInicioSecado,
			FechaFinSecado:    l.FechaFinSecado,
			CreatedAt:         l.CreatedAt,
		})
	}

	return &entities.LotesListResponse{
		Data:  items,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// GetLoteDetalle retorna el lote con última lectura, alertas activas y última predicción
func (s *LoteService) GetLoteDetalle(ctx context.Context, loteID, usuarioID int) (*entities.LoteDetalle, error) {
	lote, err := s.loteRepo.GetByID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting lote: %w", err)
	}
	if lote == nil {
		return nil, errors.New("lote not found")
	}
	if lote.UsuarioID != usuarioID {
		return nil, errors.New("unauthorized")
	}

	detalle := &entities.LoteDetalle{LoteCafe: *lote}

	// Última lectura de sensor
	lecturas, err := s.lecturaRepo.GetLatestByLoteID(ctx, loteID, 1)
	if err == nil && len(lecturas) > 0 {
		temp := lecturas[0].Temperatura
		hum := lecturas[0].Humedad
		detalle.UltimaTemperatura = &temp
		detalle.UltimaHumedad = &hum
	}

	// Cantidad de alertas no atendidas
	alertas, err := s.alertaRepo.GetByLoteID(ctx, loteID)
	if err == nil {
		for _, a := range alertas {
			if !a.Atendida {
				detalle.AlertasActivas++
			}
		}
	}

	// Última predicción
	predicciones, err := s.prediccionRepo.GetByLoteID(ctx, loteID)
	if err == nil && len(predicciones) > 0 {
		p := predicciones[0]
		detalle.UltimaPrediccion = &p
	}

	return detalle, nil
}

// CreateLote crea un nuevo lote con código QR generado automáticamente
func (s *LoteService) CreateLote(ctx context.Context, req *entities.CreateLoteRequest, usuarioID int) (*entities.LoteCafe, error) {
	lote := &entities.LoteCafe{
		UsuarioID:   usuarioID,
		NombreLote:  req.NombreLote,
		Variedad:    &req.Variedad,
		TipoProceso: &req.TipoProceso,
		PesoKg:      &req.PesoKg,
		Ubicacion:   &req.Ubicacion,
		IDSensor:    req.IDSensor,
		Estado:      "en_proceso",
	}

	created, err := s.loteRepo.Create(ctx, lote)
	if err != nil {
		return nil, fmt.Errorf("error creating lote: %w", err)
	}
	return created, nil
}

// UpdateLote actualiza los campos editables de un lote en estado 'en_proceso'
func (s *LoteService) UpdateLote(ctx context.Context, loteID, usuarioID int, req *entities.UpdateLoteRequest) (*entities.LoteCafe, error) {
	lote, err := s.loteRepo.Update(ctx, loteID, usuarioID, req.NombreLote, req.Variedad, req.PesoKg, req.Ubicacion)
	if err != nil {
		return nil, fmt.Errorf("error updating lote: %w", err)
	}
	if lote == nil {
		return nil, errors.New("lote not found or not editable")
	}
	return lote, nil
}

// FinalizarLote cambia el estado a 'finalizado', registra un evento en historial, y reporta el
// tiempo real de secado a microservicioMLL (retroalimentacion_ml) como retroalimentación para
// reentrenar en el futuro el modelo de tiempo restante -- hasta antes de esta llamada esa tabla
// nunca se llenaba. El puntaje de catación ya NO se pide aquí -- ver ReportarCatacion, más abajo.
func (s *LoteService) FinalizarLote(ctx context.Context, loteID, usuarioID int, req *entities.FinalizarLoteRequest) (*entities.LoteCafe, error) {
	now := time.Now()
	lote, err := s.loteRepo.UpdateEstado(ctx, loteID, usuarioID, "finalizado", &now)
	if err != nil {
		return nil, fmt.Errorf("error finalizing lote: %w", err)
	}
	if lote == nil {
		return nil, errors.New("lote not found or not in process")
	}

	evento := &entities.HistorialEvento{
		LoteID:      loteID,
		TipoEvento:  "lote_finalizado",
		Descripcion: fmt.Sprintf("Secado del lote '%s' finalizado", lote.NombreLote),
	}
	// El error del historial no cancela la finalización
	_ = s.historialRepo.Create(ctx, evento)

	// Reporte a microservicioMLL: best-effort, nunca cancela ni retrasa la respuesta al usuario
	// más allá del timeout corto del cliente (ver internal/infrastructure/mll/client.go). El
	// lote ya quedó finalizado en esta base de datos independientemente de si este reporte
	// tiene éxito o no.
	//
	// TiempoRealHoras se manda tal cual venga (puede ser nil): si no viene, microservicioMLL ya
	// sabe calcularlo él mismo desde fecha_inicio_secado (ver
	// app/api/routes/internal.py::registrar_resultado_real) -- no se duplica ese cálculo aquí
	// para no arriesgar que las dos cuentas diverjan por manejo de zona horaria.
	s.mllClient.ReportarResultadoReal(ctx, loteID, req.TiempoRealHoras)

	return lote, nil
}

// ReportarCatacion reporta a microservicioMLL el puntaje real de catación (escala SCA 0-100) de
// un lote ya finalizado. A diferencia de FinalizarLote, aquí SÍ se regresa el error al caller: la
// finalización del lote no depende de este dato (puede llegar semanas después), pero reportar la
// catación es, en sí mismo, la única razón de ser de este endpoint -- si microservicioMLL no lo
// pudo guardar, el usuario debe enterarse para reintentar, no quedarse pensando que ya quedó.
func (s *LoteService) ReportarCatacion(ctx context.Context, loteID, usuarioID int, req *entities.ReportarCatacionRequest) error {
	lote, err := s.loteRepo.GetByID(ctx, loteID)
	if err != nil {
		return fmt.Errorf("error getting lote: %w", err)
	}
	if lote == nil {
		return errors.New("lote not found")
	}
	if lote.UsuarioID != usuarioID {
		return errors.New("unauthorized")
	}
	if lote.Estado != "finalizado" {
		return errors.New("lote not finalized yet")
	}

	if err := s.mllClient.ReportarCatacion(ctx, loteID, *req.PuntajeSCA); err != nil {
		return fmt.Errorf("error reportando catación a microservicioMLL: %w", err)
	}
	return nil
}

// ObtenerReporteNarrativo verifica que el lote pertenezca al usuario (mismo criterio que el
// resto de accesos de lectura del lote) y le pide a microservicioMLL el reporte NLG del lote --
// microservicioMLL hace su propia verificación de dueño también (comparte la misma Neon), pero
// esta capa evita una llamada de red innecesaria si el lote ni siquiera existe para este usuario
// en la copia de Go.
func (s *LoteService) ObtenerReporteNarrativo(ctx context.Context, loteID, usuarioID int) (*entities.ReporteNarrativo, error) {
	lote, err := s.loteRepo.GetByID(ctx, loteID)
	if err != nil {
		return nil, fmt.Errorf("error getting lote: %w", err)
	}
	if lote == nil {
		return nil, errors.New("lote not found")
	}
	if lote.UsuarioID != usuarioID {
		return nil, errors.New("unauthorized")
	}

	reporte, err := s.mllClient.ObtenerReporteNLG(ctx, loteID, usuarioID)
	if err != nil {
		return nil, err
	}
	return &entities.ReporteNarrativo{
		IDReporte:     reporte.IDReporte,
		IDLote:        reporte.IDLote,
		ReporteTexto:  reporte.ReporteTexto,
		FechaGenerado: reporte.FechaGenerado,
	}, nil
}

// CancelarLote cambia el estado a 'cancelado' (soft delete)
func (s *LoteService) CancelarLote(ctx context.Context, loteID, usuarioID int) error {
	lote, err := s.loteRepo.UpdateEstado(ctx, loteID, usuarioID, "cancelado", nil)
	if err != nil {
		return fmt.Errorf("error canceling lote: %w", err)
	}
	if lote == nil {
		return errors.New("lote not found or not in process")
	}
	return nil
}

// ReclamarLote asigna al usuario autenticado un lote pre-creado por api-web (dueño
// placeholder) mediante el codigo_qr escaneado, y registra el evento en historial.
func (s *LoteService) ReclamarLote(ctx context.Context, codigoQR string, usuarioID int) (*entities.LoteCafe, error) {
	lote, err := s.loteRepo.ReclamarLote(ctx, codigoQR, usuarioID, s.placeholderUsuarioID)
	if err != nil {
		return nil, fmt.Errorf("error reclamando lote: %w", err)
	}
	if lote == nil {
		return nil, errors.New("codigo qr invalido o ya utilizado")
	}

	evento := &entities.HistorialEvento{
		LoteID:      lote.ID,
		UsuarioID:   &usuarioID,
		TipoEvento:  "lote_reclamado",
		Descripcion: fmt.Sprintf("Lote '%s' reclamado mediante escaneo de QR", lote.NombreLote),
	}
	// El error del historial no cancela el reclamo
	_ = s.historialRepo.Create(ctx, evento)

	return lote, nil
}
