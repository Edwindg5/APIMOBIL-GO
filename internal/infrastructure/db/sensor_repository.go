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

func NewSensorRepository(db *PostgresDB) interfaces.SensorRepository {
	return &SensorRepository{db: db}
}

const sensorCols = `id_sensor, mac_address, id_cola_mqtt, provisioning_token, token_usado, tipo, modelo, estado, fecha_registro, ultima_conexion`

func scanSensor(row interface{ Scan(...any) error }, s *entities.Sensor) error {
	return row.Scan(
		&s.ID, &s.MacAddress, &s.IDColaMQTT, &s.ProvisioningToken, &s.TokenUsado,
		&s.Tipo, &s.Modelo, &s.Estado, &s.FechaRegistro, &s.UltimaConexion,
	)
}

// GetByESP32ID busca un sensor por el identificador enviado por la app móvil.
// La BD real no tiene columna esp32_id: el identificador se resuelve contra mac_address.
func (r *SensorRepository) GetByESP32ID(ctx context.Context, esp32ID string) (*entities.Sensor, error) {
	return r.GetByIdentifier(ctx, esp32ID)
}

func (r *SensorRepository) GetByIdentifier(ctx context.Context, identifier string) (*entities.Sensor, error) {
	s := &entities.Sensor{}
	err := scanSensor(r.db.GetPool().QueryRow(ctx, `
		SELECT `+sensorCols+`
		FROM sensores
		WHERE mac_address = $1
		LIMIT 1
	`, identifier), s)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (r *SensorRepository) GetByID(ctx context.Context, id int) (*entities.Sensor, error) {
	s := &entities.Sensor{}
	err := scanSensor(r.db.GetPool().QueryRow(ctx,
		`SELECT `+sensorCols+` FROM sensores WHERE id_sensor = $1`, id), s)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (r *SensorRepository) Create(ctx context.Context, sensor *entities.Sensor) (int, error) {
	var id int
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO sensores (mac_address, id_cola_mqtt, tipo, estado, fecha_registro)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id_sensor
	`, sensor.MacAddress, sensor.IDColaMQTT, sensor.Tipo, sensor.Estado).Scan(&id)
	return id, err
}

// LinkToLote vincula un sensor a un lote. La relación real vive en lotes_cafe.id_sensor
// (ya asignado al crear el lote); aquí solo se registra linked_at y se activa el sensor.
func (r *SensorRepository) LinkToLote(ctx context.Context, sensorID, loteID int) error {
	tag, err := r.db.GetPool().Exec(ctx, `
		UPDATE lotes_cafe
		SET linked_at = NOW()
		WHERE id_lote = $1 AND id_sensor = $2
	`, loteID, sensorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("lote %d not linked to sensor %d", loteID, sensorID)
	}

	_, err = r.db.GetPool().Exec(ctx, `
		UPDATE sensores SET estado = 'activo', ultima_conexion = NOW() WHERE id_sensor = $1
	`, sensorID)
	return err
}

func (r *SensorRepository) MarcarTokenUsado(ctx context.Context, sensorID int) error {
	_, err := r.db.GetPool().Exec(ctx, `
		UPDATE sensores SET token_usado = true WHERE id_sensor = $1
	`, sensorID)
	return err
}

// CountByUsuarioID cuenta sensores distintos asociados a los lotes del usuario.
// sensores no tiene FK directa a usuarios; la relación es indirecta vía lotes_cafe.
func (r *SensorRepository) CountByUsuarioID(ctx context.Context, usuarioID int) (int, error) {
	var count int
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT COUNT(DISTINCT s.id_sensor)
		FROM sensores s
		INNER JOIN lotes_cafe l ON l.id_sensor = s.id_sensor
		WHERE l.id_usuario = $1
	`, usuarioID).Scan(&count)
	return count, err
}
