package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type UsuarioRepository struct {
	db *PostgresDB
}

// NewUsuarioRepository crea una nueva instancia del repositorio
func NewUsuarioRepository(db *PostgresDB) interfaces.UsuarioRepository {
	return &UsuarioRepository{db: db}
}

// GetByEmail obtiene un usuario por email (sin RLS, es para login)
func (r *UsuarioRepository) GetByEmail(ctx context.Context, email string) (*entities.Usuario, error) {
	usuario := &entities.Usuario{}
	
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id, email, password, nombre_completo, rol, estado, created_at, updated_at
		FROM usuarios
		WHERE email = $1 AND estado = 'activo'
	`, email).Scan(
		&usuario.ID,
		&usuario.Email,
		&usuario.Password,
		&usuario.NombreCompleto,
		&usuario.Rol,
		&usuario.Estado,
		&usuario.CreatedAt,
		&usuario.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No encontrado
		}
		return nil, err
	}

	return usuario, nil
}

// GetByID obtiene un usuario por ID
func (r *UsuarioRepository) GetByID(ctx context.Context, id int) (*entities.Usuario, error) {
	usuario := &entities.Usuario{}
	
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id, email, password, nombre_completo, rol, estado, created_at, updated_at
		FROM usuarios
		WHERE id = $1
	`, id).Scan(
		&usuario.ID,
		&usuario.Email,
		&usuario.Password,
		&usuario.NombreCompleto,
		&usuario.Rol,
		&usuario.Estado,
		&usuario.CreatedAt,
		&usuario.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return usuario, nil
}

// Create crea un nuevo usuario
func (r *UsuarioRepository) Create(ctx context.Context, usuario *entities.Usuario) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO usuarios (email, password, nombre_completo, rol, estado, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`, usuario.Email, usuario.Password, usuario.NombreCompleto, usuario.Rol, usuario.Estado).Scan(
		&usuario.ID,
		&usuario.CreatedAt,
		&usuario.UpdatedAt,
	)

	return err
}
