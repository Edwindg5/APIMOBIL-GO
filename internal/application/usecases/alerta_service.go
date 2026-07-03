package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type AlertaService struct {
	alertaRepository interfaces.AlertaRepository
	loteRepository   interfaces.LoteRepository
	historialRepo    interfaces.HistorialRepository
}

func NewAlertaService(
	alertaRepository interfaces.AlertaRepository,
	loteRepository interfaces.LoteRepository,
	historialRepo interfaces.HistorialRepository,
) interfaces.AlertaService {
	return &AlertaService{
		alertaRepository: alertaRepository,
		loteRepository:   loteRepository,
		historialRepo:    historialRepo,
	}
}

func (s *AlertaService) GetAlertas(ctx context.Context, loteID, usuarioID int, atendida *bool, nivel string) ([]entities.Alerta, error) {
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

	alertas, err := s.alertaRepository.GetByLoteIDFiltered(ctx, loteID, atendida, nivel)
	if err != nil {
		return nil, fmt.Errorf("error getting alertas: %w", err)
	}
	return alertas, nil
}

func (s *AlertaService) AtenderAlerta(ctx context.Context, alertaID, usuarioID int) (*entities.Alerta, error) {
	alerta, err := s.alertaRepository.GetByID(ctx, alertaID)
	if err != nil {
		return nil, fmt.Errorf("error getting alerta: %w", err)
	}
	if alerta == nil {
		return nil, errors.New("alerta not found")
	}

	lote, err := s.loteRepository.GetByID(ctx, alerta.LoteID)
	if err != nil {
		return nil, fmt.Errorf("error getting lote: %w", err)
	}
	if lote == nil || lote.UsuarioID != usuarioID {
		return nil, errors.New("unauthorized")
	}

	updated, err := s.alertaRepository.MarcarAtendida(ctx, alertaID)
	if err != nil {
		return nil, fmt.Errorf("error marking alerta: %w", err)
	}

	evento := &entities.HistorialEvento{
		LoteID:      alerta.LoteID,
		TipoEvento:  "alerta_atendida",
		Descripcion: fmt.Sprintf("Alerta '%s' marcada como atendida", alerta.TipoAlerta),
	}
	_ = s.historialRepo.Create(ctx, evento)

	return updated, nil
}
