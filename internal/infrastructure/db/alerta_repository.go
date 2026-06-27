package db

import (
	"context"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type AlertaRepository struct {
	db *PostgresDB
}

// NewAlertaRepository crea una nueva instancia del repositorio
func NewAlertaRepository(db *PostgresDB) interfaces.AlertaRepository {
	return &AlertaRepository{db: db}
}

// GetByLoteID obtiene todas las alertas de un lote
func (r *AlertaRepository) GetByLoteID(ctx context.Context, loteID int) ([]entities.Alerta, error) {
	rows, err := r.db.GetPool().Query(ctx, `
		SELECT id, lote_id, tipo, mensaje, nivel, leida, created_at, updated_at
		FROM alertas
		WHERE lote_id = $1
		ORDER BY created_at DESC
	`, loteID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alertas []entities.Alerta
	for rows.Next() {
		var alerta entities.Alerta
		err := rows.Scan(
			&alerta.ID,
			&alerta.LoteID,
			&alerta.Tipo,
			&alerta.Mensaje,
			&alerta.Nivel,
			&alerta.Leida,
			&alerta.CreatedAt,
			&alerta.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		alertas = append(alertas, alerta)
	}

	return alertas, nil
}

// Create crea una nueva alerta
func (r *AlertaRepository) Create(ctx context.Context, alerta *entities.Alerta) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO alertas (lote_id, tipo, mensaje, nivel, leida, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`, alerta.LoteID, alerta.Tipo, alerta.Mensaje, alerta.Nivel, alerta.Leida).Scan(
		&alerta.ID,
		&alerta.CreatedAt,
		&alerta.UpdatedAt,
	)

	return err
}
