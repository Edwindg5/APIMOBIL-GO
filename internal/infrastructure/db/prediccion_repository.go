package db

import (
	"context"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type PrediccionRepository struct {
	db *PostgresDB
}

// NewPrediccionRepository crea una nueva instancia del repositorio
func NewPrediccionRepository(db *PostgresDB) interfaces.PrediccionRepository {
	return &PrediccionRepository{db: db}
}

// GetByLoteID obtiene todas las predicciones de un lote
func (r *PrediccionRepository) GetByLoteID(ctx context.Context, loteID int) ([]entities.Prediccion, error) {
	rows, err := r.db.GetPool().Query(ctx, `
		SELECT id, lote_id, prediccion, probabilidad, created_at
		FROM predicciones
		WHERE lote_id = $1
		ORDER BY created_at DESC
	`, loteID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predicciones []entities.Prediccion
	for rows.Next() {
		var prediccion entities.Prediccion
		err := rows.Scan(
			&prediccion.ID,
			&prediccion.LoteID,
			&prediccion.Prediccion,
			&prediccion.Probabilidad,
			&prediccion.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		predicciones = append(predicciones, prediccion)
	}

	return predicciones, nil
}

// Create crea una nueva predicción
func (r *PrediccionRepository) Create(ctx context.Context, prediccion *entities.Prediccion) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO predicciones (lote_id, prediccion, probabilidad, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`, prediccion.LoteID, prediccion.Prediccion, prediccion.Probabilidad).Scan(
		&prediccion.ID,
		&prediccion.CreatedAt,
	)

	return err
}
