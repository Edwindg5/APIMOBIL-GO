package db

import (
	"context"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type RecomendacionRepository struct {
	db *PostgresDB
}

// NewRecomendacionRepository crea una nueva instancia del repositorio
func NewRecomendacionRepository(db *PostgresDB) interfaces.RecomendacionRepository {
	return &RecomendacionRepository{db: db}
}

// GetByLoteID obtiene todas las recomendaciones de un lote
func (r *RecomendacionRepository) GetByLoteID(ctx context.Context, loteID int) ([]entities.Recomendacion, error) {
	rows, err := r.db.GetPool().Query(ctx, `
		SELECT id, lote_id, accion, razon, prioridad, created_at
		FROM recomendaciones
		WHERE lote_id = $1
		ORDER BY created_at DESC
	`, loteID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recomendaciones []entities.Recomendacion
	for rows.Next() {
		var rec entities.Recomendacion
		err := rows.Scan(
			&rec.ID,
			&rec.LoteID,
			&rec.Accion,
			&rec.Razon,
			&rec.Prioridad,
			&rec.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		recomendaciones = append(recomendaciones, rec)
	}

	return recomendaciones, nil
}

// Create crea una nueva recomendación
func (r *RecomendacionRepository) Create(ctx context.Context, recomendacion *entities.Recomendacion) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO recomendaciones (lote_id, accion, razon, prioridad, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`, recomendacion.LoteID, recomendacion.Accion, recomendacion.Razon, recomendacion.Prioridad).Scan(
		&recomendacion.ID,
		&recomendacion.CreatedAt,
	)

	return err
}
