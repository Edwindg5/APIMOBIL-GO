package db

import (
	"context"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type LecturaRepository struct {
	db *PostgresDB
}

// NewLecturaRepository crea una nueva instancia del repositorio
func NewLecturaRepository(db *PostgresDB) interfaces.LecturaRepository {
	return &LecturaRepository{db: db}
}

// GetLatestByLoteID obtiene las últimas N lecturas de un lote
func (r *LecturaRepository) GetLatestByLoteID(ctx context.Context, loteID int, limit int) ([]entities.LecturaAmbiental, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	rows, err := r.db.GetPool().Query(ctx, `
		SELECT id, lote_id, sensor_id, temperatura, humedad, presion, created_at
		FROM lecturas_ambientales
		WHERE lote_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, loteID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lecturas []entities.LecturaAmbiental
	for rows.Next() {
		var lectura entities.LecturaAmbiental
		err := rows.Scan(
			&lectura.ID,
			&lectura.LoteID,
			&lectura.SensorID,
			&lectura.Temperatura,
			&lectura.Humedad,
			&lectura.Presion,
			&lectura.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		lecturas = append(lecturas, lectura)
	}

	return lecturas, nil
}

// Create crea una nueva lectura
func (r *LecturaRepository) Create(ctx context.Context, lectura *entities.LecturaAmbiental) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO lecturas_ambientales (lote_id, sensor_id, temperatura, humedad, presion, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`, lectura.LoteID, lectura.SensorID, lectura.Temperatura, lectura.Humedad, lectura.Presion).Scan(
		&lectura.ID,
		&lectura.CreatedAt,
	)

	return err
}
