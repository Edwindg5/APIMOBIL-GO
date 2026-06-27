package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type LoteRepository struct {
	db *PostgresDB
}

// NewLoteRepository crea una nueva instancia del repositorio
func NewLoteRepository(db *PostgresDB) interfaces.LoteRepository {
	return &LoteRepository{db: db}
}

// GetByID obtiene un lote por ID
func (r *LoteRepository) GetByID(ctx context.Context, id int) (*entities.LoteCafe, error) {
	lote := &entities.LoteCafe{}
	
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id, usuario_id, nombre, descripcion, area, sensor_id, estado, created_at, updated_at
		FROM lotes_cafe
		WHERE id = $1
	`, id).Scan(
		&lote.ID,
		&lote.UsuarioID,
		&lote.Nombre,
		&lote.Descripcion,
		&lote.Area,
		&lote.SensorID,
		&lote.Estado,
		&lote.CreatedAt,
		&lote.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return lote, nil
}

// GetByUsuarioID obtiene todos los lotes de un usuario
func (r *LoteRepository) GetByUsuarioID(ctx context.Context, usuarioID int) ([]entities.LoteCafe, error) {
	rows, err := r.db.GetPool().Query(ctx, `
		SELECT id, usuario_id, nombre, descripcion, area, sensor_id, estado, created_at, updated_at
		FROM lotes_cafe
		WHERE usuario_id = $1
		ORDER BY created_at DESC
	`, usuarioID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lotes []entities.LoteCafe
	for rows.Next() {
		var lote entities.LoteCafe
		err := rows.Scan(
			&lote.ID,
			&lote.UsuarioID,
			&lote.Nombre,
			&lote.Descripcion,
			&lote.Area,
			&lote.SensorID,
			&lote.Estado,
			&lote.CreatedAt,
			&lote.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		lotes = append(lotes, lote)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return lotes, nil
}

// Create crea un nuevo lote
func (r *LoteRepository) Create(ctx context.Context, lote *entities.LoteCafe) (int, error) {
	var id int
	
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO lotes_cafe (usuario_id, nombre, descripcion, area, estado, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`, lote.UsuarioID, lote.Nombre, lote.Descripcion, lote.Area, lote.Estado).Scan(&id)

	return id, err
}

// Update actualiza un lote
func (r *LoteRepository) Update(ctx context.Context, lote *entities.LoteCafe) error {
	_, err := r.db.GetPool().Exec(ctx, `
		UPDATE lotes_cafe
		SET nombre = $1, descripcion = $2, area = $3, estado = $4, sensor_id = $5, updated_at = NOW()
		WHERE id = $6
	`, lote.Nombre, lote.Descripcion, lote.Area, lote.Estado, lote.SensorID, lote.ID)

	return err
}
