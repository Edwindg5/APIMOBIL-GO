package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type ReporteRepository struct {
	db *PostgresDB
}

// NewReporteRepository crea una nueva instancia del repositorio
func NewReporteRepository(db *PostgresDB) interfaces.ReporteRepository {
	return &ReporteRepository{db: db}
}

// GetByID obtiene un reporte por ID
func (r *ReporteRepository) GetByID(ctx context.Context, id int) (*entities.Reporte, error) {
	reporte := &entities.Reporte{}

	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id, lote_id, usuario_id, tipo, estado, url, created_at, updated_at
		FROM reportes
		WHERE id = $1
	`, id).Scan(
		&reporte.ID,
		&reporte.LoteID,
		&reporte.UsuarioID,
		&reporte.Tipo,
		&reporte.Estado,
		&reporte.URL,
		&reporte.CreatedAt,
		&reporte.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return reporte, nil
}

// Create crea un nuevo reporte
func (r *ReporteRepository) Create(ctx context.Context, reporte *entities.Reporte) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO reportes (lote_id, usuario_id, tipo, estado, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`, reporte.LoteID, reporte.UsuarioID, reporte.Tipo, reporte.Estado).Scan(
		&reporte.ID,
		&reporte.CreatedAt,
		&reporte.UpdatedAt,
	)

	return err
}

// Update actualiza un reporte
func (r *ReporteRepository) Update(ctx context.Context, reporte *entities.Reporte) error {
	_, err := r.db.GetPool().Exec(ctx, `
		UPDATE reportes
		SET tipo = $1, estado = $2, url = $3, updated_at = NOW()
		WHERE id = $4
	`, reporte.Tipo, reporte.Estado, reporte.URL, reporte.ID)

	return err
}
