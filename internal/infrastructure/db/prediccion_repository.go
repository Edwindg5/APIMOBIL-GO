package db

import (
	"context"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type PrediccionRepository struct {
	db *PostgresDB
}

func NewPrediccionRepository(db *PostgresDB) interfaces.PrediccionRepository {
	return &PrediccionRepository{db: db}
}

const prediccionCols = `id_prediccion, id_lote, id_modelo, tiempo_estimado_horas, calidad_estimada, confianza, riesgo_lluvia_proxima, horas_anticipacion_lluvia, fecha_prediccion`

func (r *PrediccionRepository) GetByLoteID(ctx context.Context, loteID int, limit int) ([]entities.Prediccion, error) {
	query := `
		SELECT ` + prediccionCols + `
		FROM predicciones
		WHERE id_lote = $1
		ORDER BY fecha_prediccion DESC
	`
	args := []any{loteID}
	if limit > 0 {
		query += ` LIMIT $2`
		args = append(args, limit)
	}

	rows, err := r.db.GetPool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var predicciones []entities.Prediccion
	for rows.Next() {
		var p entities.Prediccion
		if err := rows.Scan(
			&p.ID, &p.LoteID, &p.IDModelo, &p.TiempoEstimadoHoras, &p.CalidadEstimada,
			&p.Confianza, &p.RiesgoLluviaProxima, &p.HorasAnticipacionLluvia, &p.FechaPrediccion,
		); err != nil {
			return nil, err
		}
		predicciones = append(predicciones, p)
	}
	return predicciones, rows.Err()
}

func (r *PrediccionRepository) Create(ctx context.Context, prediccion *entities.Prediccion) error {
	return r.db.GetPool().QueryRow(ctx, `
		INSERT INTO predicciones
			(id_lote, id_modelo, tiempo_estimado_horas, calidad_estimada, confianza, fecha_prediccion)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id_prediccion, fecha_prediccion
	`, prediccion.LoteID, prediccion.IDModelo, prediccion.TiempoEstimadoHoras,
		prediccion.CalidadEstimada, prediccion.Confianza,
	).Scan(&prediccion.ID, &prediccion.FechaPrediccion)
}
