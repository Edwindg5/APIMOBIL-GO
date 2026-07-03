package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type AlertaRepository struct {
	db *PostgresDB
}

func NewAlertaRepository(db *PostgresDB) interfaces.AlertaRepository {
	return &AlertaRepository{db: db}
}

const alertaCols = `id_alerta, id_lote, id_sensor, tipo_alerta, mensaje, nivel_severidad, atendida, fecha_atencion, fecha_generada`

func scanAlertaRow(row interface{ Scan(...any) error }, a *entities.Alerta) error {
	return row.Scan(
		&a.ID, &a.LoteID, &a.SensorID, &a.TipoAlerta, &a.Mensaje, &a.NivelSeveridad,
		&a.Atendida, &a.FechaAtencion, &a.FechaGenerada,
	)
}

func (r *AlertaRepository) GetByID(ctx context.Context, id int) (*entities.Alerta, error) {
	a := &entities.Alerta{}
	err := scanAlertaRow(
		r.db.GetPool().QueryRow(ctx, `SELECT `+alertaCols+` FROM alertas WHERE id_alerta = $1`, id), a,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

func (r *AlertaRepository) GetByLoteID(ctx context.Context, loteID int) ([]entities.Alerta, error) {
	return r.GetByLoteIDFiltered(ctx, loteID, nil, "")
}

func (r *AlertaRepository) GetByLoteIDFiltered(ctx context.Context, loteID int, atendida *bool, nivel string) ([]entities.Alerta, error) {
	rows, err := r.db.GetPool().Query(ctx, `
		SELECT `+alertaCols+`
		FROM alertas
		WHERE id_lote = $1
		  AND ($2::boolean IS NULL OR atendida = $2)
		  AND ($3 = '' OR nivel_severidad = $3)
		ORDER BY fecha_generada DESC
	`, loteID, atendida, nivel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alertas []entities.Alerta
	for rows.Next() {
		var a entities.Alerta
		if err := scanAlertaRow(rows, &a); err != nil {
			return nil, err
		}
		alertas = append(alertas, a)
	}
	return alertas, rows.Err()
}

func (r *AlertaRepository) MarcarAtendida(ctx context.Context, alertaID int) (*entities.Alerta, error) {
	now := time.Now()
	a := &entities.Alerta{}
	err := scanAlertaRow(r.db.GetPool().QueryRow(ctx, `
		UPDATE alertas
		SET atendida = true, fecha_atencion = $1
		WHERE id_alerta = $2
		RETURNING `+alertaCols, now, alertaID), a)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

func (r *AlertaRepository) Create(ctx context.Context, alerta *entities.Alerta) error {
	return r.db.GetPool().QueryRow(ctx, `
		INSERT INTO alertas (id_lote, id_sensor, tipo_alerta, mensaje, nivel_severidad, atendida, fecha_generada)
		VALUES ($1, $2, $3, $4, $5, false, NOW())
		RETURNING id_alerta, fecha_generada
	`, alerta.LoteID, alerta.SensorID, alerta.TipoAlerta, alerta.Mensaje, alerta.NivelSeveridad,
	).Scan(&alerta.ID, &alerta.FechaGenerada)
}
