package db

import (
	"context"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type RecomendacionRepository struct {
	db *PostgresDB
}

func NewRecomendacionRepository(db *PostgresDB) interfaces.RecomendacionRepository {
	return &RecomendacionRepository{db: db}
}

func (r *RecomendacionRepository) GetByLoteID(ctx context.Context, loteID int) ([]entities.Recomendacion, error) {
	rows, err := r.db.GetPool().Query(ctx, `
		SELECT id_recomendacion, id_lote, texto, origen, fecha_generada
		FROM recomendaciones
		WHERE id_lote = $1
		ORDER BY fecha_generada DESC
	`, loteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recomendaciones []entities.Recomendacion
	for rows.Next() {
		var rec entities.Recomendacion
		if err := rows.Scan(&rec.ID, &rec.LoteID, &rec.Texto, &rec.Origen, &rec.FechaGenerada); err != nil {
			return nil, err
		}
		recomendaciones = append(recomendaciones, rec)
	}
	return recomendaciones, rows.Err()
}

func (r *RecomendacionRepository) Create(ctx context.Context, rec *entities.Recomendacion) error {
	return r.db.GetPool().QueryRow(ctx, `
		INSERT INTO recomendaciones (id_lote, texto, origen, fecha_generada)
		VALUES ($1, $2, $3, NOW())
		RETURNING id_recomendacion, fecha_generada
	`, rec.LoteID, rec.Texto, rec.Origen,
	).Scan(&rec.ID, &rec.FechaGenerada)
}
