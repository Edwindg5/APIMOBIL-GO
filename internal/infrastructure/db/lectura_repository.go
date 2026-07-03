package db

import (
	"context"
	"time"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type LecturaRepository struct {
	db *PostgresDB
}

func NewLecturaRepository(db *PostgresDB) interfaces.LecturaRepository {
	return &LecturaRepository{db: db}
}

const lecturaCols = `id_lectura, id_lote, id_sensor, temperatura, humedad, timestamp`

func scanLectura(rows interface{ Scan(...any) error }, l *entities.LecturaAmbiental) error {
	return rows.Scan(&l.ID, &l.LoteID, &l.SensorID, &l.Temperatura, &l.Humedad, &l.Timestamp)
}

func (r *LecturaRepository) GetLatestByLoteID(ctx context.Context, loteID int, limit int) ([]entities.LecturaAmbiental, error) {
	return r.GetByLoteIDFiltered(ctx, loteID, limit, time.Time{})
}

func (r *LecturaRepository) GetByLoteIDFiltered(ctx context.Context, loteID int, limit int, desde time.Time) ([]entities.LecturaAmbiental, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	var rows interface {
		Next() bool
		Scan(...any) error
		Close()
		Err() error
	}
	var err error

	if desde.IsZero() {
		rows, err = r.db.GetPool().Query(ctx, `
			SELECT `+lecturaCols+`
			FROM lecturas_ambientales
			WHERE id_lote = $1
			ORDER BY timestamp DESC
			LIMIT $2
		`, loteID, limit)
	} else {
		rows, err = r.db.GetPool().Query(ctx, `
			SELECT `+lecturaCols+`
			FROM lecturas_ambientales
			WHERE id_lote = $1 AND timestamp >= $2
			ORDER BY timestamp DESC
			LIMIT $3
		`, loteID, desde, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lecturas []entities.LecturaAmbiental
	for rows.Next() {
		var l entities.LecturaAmbiental
		if err := scanLectura(rows, &l); err != nil {
			return nil, err
		}
		lecturas = append(lecturas, l)
	}
	return lecturas, rows.Err()
}

func (r *LecturaRepository) GetEstadisticas(ctx context.Context, loteID int) (*entities.EstadisticasLote, error) {
	stats := &entities.EstadisticasLote{}

	// Aggregate lecturas
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT
			COALESCE(AVG(temperatura), 0),
			COALESCE(MIN(temperatura), 0),
			COALESCE(MAX(temperatura), 0),
			COALESCE(AVG(humedad), 0),
			COALESCE(MIN(humedad), 0),
			COALESCE(MAX(humedad), 0),
			COUNT(*),
			MAX(timestamp)
		FROM lecturas_ambientales
		WHERE id_lote = $1
	`, loteID).Scan(
		&stats.TemperaturaPromedio, &stats.TemperaturaMin, &stats.TemperaturaMax,
		&stats.HumedadPromedio, &stats.HumedadMin, &stats.HumedadMax,
		&stats.TotalLecturas, &stats.UltimaLectura,
	)
	if err != nil {
		return nil, err
	}

	// Aggregate alertas
	err = r.db.GetPool().QueryRow(ctx, `
		SELECT
			COUNT(*),
			SUM(CASE WHEN nivel_severidad = 'critica' THEN 1 ELSE 0 END),
			SUM(CASE WHEN NOT atendida THEN 1 ELSE 0 END)
		FROM alertas
		WHERE id_lote = $1
	`, loteID).Scan(&stats.TotalAlertas, &stats.AlertasCriticas, &stats.AlertasSinAtender)
	if err != nil {
		return nil, err
	}

	// Días de secado desde fecha_inicio_secado
	err = r.db.GetPool().QueryRow(ctx, `
		SELECT GREATEST(0, EXTRACT(DAY FROM (NOW() - fecha_inicio_secado))::int)
		FROM lotes_cafe WHERE id_lote = $1
	`, loteID).Scan(&stats.DiasSecado)
	if err != nil {
		stats.DiasSecado = 0
	}

	return stats, nil
}

func (r *LecturaRepository) Create(ctx context.Context, lectura *entities.LecturaAmbiental) error {
	return r.db.GetPool().QueryRow(ctx, `
		INSERT INTO lecturas_ambientales (id_lote, id_sensor, temperatura, humedad, timestamp)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id_lectura, timestamp
	`, lectura.LoteID, lectura.SensorID, lectura.Temperatura, lectura.Humedad,
	).Scan(&lectura.ID, &lectura.Timestamp)
}
