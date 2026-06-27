package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type SensorRepository struct {
	db *PostgresDB
}

// NewSensorRepository crea una nueva instancia del repositorio
func NewSensorRepository(db *PostgresDB) interfaces.SensorRepository {
	return &SensorRepository{db: db}
}

// GetByESP32ID obtiene un sensor por su ID ESP32
func (r *SensorRepository) GetByESP32ID(ctx context.Context, esp32ID string) (*entities.Sensor, error) {
	sensor := &entities.Sensor{}
	
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id, esp32_id, lote_id, linked_at, last_seen, estado, created_at, updated_at
		FROM sensores
		WHERE esp32_id = $1
	`, esp32ID).Scan(
		&sensor.ID,
		&sensor.ESP32ID,
		&sensor.LoteID,
		&sensor.LinkedAt,
		&sensor.LastSeen,
		&sensor.Estado,
		&sensor.CreatedAt,
		&sensor.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return sensor, nil
}

// GetByID obtiene un sensor por ID
func (r *SensorRepository) GetByID(ctx context.Context, id int) (*entities.Sensor, error) {
	sensor := &entities.Sensor{}
	
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id, esp32_id, lote_id, linked_at, last_seen, estado, created_at, updated_at
		FROM sensores
		WHERE id = $1
	`, id).Scan(
		&sensor.ID,
		&sensor.ESP32ID,
		&sensor.LoteID,
		&sensor.LinkedAt,
		&sensor.LastSeen,
		&sensor.Estado,
		&sensor.CreatedAt,
		&sensor.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return sensor, nil
}

// Create crea un nuevo sensor
func (r *SensorRepository) Create(ctx context.Context, sensor *entities.Sensor) (int, error) {
	var id int
	
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO sensores (esp32_id, lote_id, estado, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`, sensor.ESP32ID, sensor.LoteID, sensor.Estado).Scan(&id)

	return id, err
}

// LinkToLote vincula un sensor a un lote
func (r *SensorRepository) LinkToLote(ctx context.Context, sensorID, loteID int) error {
	tag, err := r.db.GetPool().Exec(ctx, `
		UPDATE sensores
		SET lote_id = $1, linked_at = NOW(), estado = 'activo', updated_at = NOW()
		WHERE id = $2
	`, loteID, sensorID)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("sensor %d not found", sensorID)
	}

	return nil
}
