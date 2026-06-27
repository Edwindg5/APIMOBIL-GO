package db

import (
	"context"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type HistorialRepository struct {
	db *PostgresDB
}

// NewHistorialRepository crea una nueva instancia del repositorio
func NewHistorialRepository(db *PostgresDB) interfaces.HistorialRepository {
	return &HistorialRepository{db: db}
}

// GetByLoteID obtiene todos los eventos del historial de un lote
func (r *HistorialRepository) GetByLoteID(ctx context.Context, loteID int) ([]entities.HistorialEvento, error) {
	rows, err := r.db.GetPool().Query(ctx, `
		SELECT id, lote_id, tipo, descripcion, created_at
		FROM historial_eventos
		WHERE lote_id = $1
		ORDER BY created_at DESC
	`, loteID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eventos []entities.HistorialEvento
	for rows.Next() {
		var evento entities.HistorialEvento
		err := rows.Scan(
			&evento.ID,
			&evento.LoteID,
			&evento.Tipo,
			&evento.Descripcion,
			&evento.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		eventos = append(eventos, evento)
	}

	return eventos, nil
}

// Create crea un nuevo evento en el historial
func (r *HistorialRepository) Create(ctx context.Context, evento *entities.HistorialEvento) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO historial_eventos (lote_id, tipo, descripcion, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`, evento.LoteID, evento.Tipo, evento.Descripcion).Scan(
		&evento.ID,
		&evento.CreatedAt,
	)

	return err
}
