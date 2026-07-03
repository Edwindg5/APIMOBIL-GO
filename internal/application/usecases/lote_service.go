package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type LoteService struct {
	loteRepo       interfaces.LoteRepository
	historialRepo  interfaces.HistorialRepository
	lecturaRepo    interfaces.LecturaRepository
	alertaRepo     interfaces.AlertaRepository
	prediccionRepo interfaces.PrediccionRepository
}

// NewLoteService crea una nueva instancia del servicio
func NewLoteService(
	loteRepo interfaces.LoteRepository,
	historialRepo interfaces.HistorialRepository,
	lecturaRepo interfaces.LecturaRepository,
	alertaRepo interfaces.AlertaRepository,
	prediccionRepo interfaces.PrediccionRepository,
) interfaces.LoteService {
	return &LoteService{
		loteRepo:       loteRepo,
		historialRepo:  historialRepo,
		lecturaRepo:    lecturaRepo,
		alertaRepo:     alertaRepo,
		prediccionRepo: prediccionRepo,
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
		Variedad:    req.Variedad,
		TipoProceso: req.TipoProceso,
		PesoKg:      req.PesoKg,
		Ubicacion:   req.Ubicacion,
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

// FinalizarLote cambia el estado a 'finalizado' y registra un evento en historial
func (s *LoteService) FinalizarLote(ctx context.Context, loteID, usuarioID int) (*entities.LoteCafe, error) {
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

	return lote, nil
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
