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

// GetByEmail obtiene un usuario activo por email (sin RLS, uso en login)
func (r *UsuarioRepository) GetByEmail(ctx context.Context, email string) (*entities.Usuario, error) {
	usuario := &entities.Usuario{}

	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id_usuario, nombre, email, password_hash, rol, telefono, estado, fecha_registro
		FROM usuarios
		WHERE email = $1 AND estado = 'activo'
	`, email).Scan(
		&usuario.ID,
		&usuario.Nombre,
		&usuario.Email,
		&usuario.PasswordHash,
		&usuario.Rol,
		&usuario.Telefono,
		&usuario.Estado,
		&usuario.FechaRegistro,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return usuario, nil
}

// GetByID obtiene un usuario por ID.
// es_premium_real recalcula la suscripción en tiempo real en vez de confiar en la
// columna cruda usuarios.es_premium: un ticker externo la apaga cuando vence
// premium_hasta, pero puede haber una ventana de minutos con datos desactualizados.
// COALESCE evita un NULL AND ... indeterminado si es_premium llega nulo.
func (r *UsuarioRepository) GetByID(ctx context.Context, id int) (*entities.Usuario, error) {
	usuario := &entities.Usuario{}

	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id_usuario, nombre, email, password_hash, rol, telefono, estado, fecha_registro,
			COALESCE(es_premium, false) AND (premium_hasta IS NULL OR premium_hasta > NOW()) AS es_premium_real,
			premium_hasta
		FROM usuarios
		WHERE id_usuario = $1
	`, id).Scan(
		&usuario.ID,
		&usuario.Nombre,
		&usuario.Email,
		&usuario.PasswordHash,
		&usuario.Rol,
		&usuario.Telefono,
		&usuario.Estado,
		&usuario.FechaRegistro,
		&usuario.EsPremium,
		&usuario.PremiumHasta,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return usuario, nil
}

// ExistsByEmail comprueba si ya existe algún usuario con ese email, independientemente del estado
func (r *UsuarioRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM usuarios WHERE email = $1)
	`, email).Scan(&exists)
	return exists, err
}

// Create crea un nuevo usuario
func (r *UsuarioRepository) Create(ctx context.Context, usuario *entities.Usuario) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO usuarios (nombre, email, password_hash, telefono, rol, estado, fecha_registro)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id_usuario, fecha_registro
	`, usuario.Nombre, usuario.Email, usuario.PasswordHash, usuario.Telefono, usuario.Rol, usuario.Estado).Scan(
		&usuario.ID,
		&usuario.FechaRegistro,
	)

	return err
}

// Update actualiza nombre y telefono del usuario; usa transacción con RLS
func (r *UsuarioRepository) Update(ctx context.Context, id int, nombre, telefono string) (*entities.Usuario, error) {
	tx, err := r.db.BeginTx(ctx, id)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	usuario := &entities.Usuario{}
	err = tx.QueryRow(ctx, `
		UPDATE usuarios
		SET nombre = $1, telefono = $2
		WHERE id_usuario = $3
		RETURNING id_usuario, nombre, email, telefono, rol, estado, fecha_registro
	`, nombre, telefono, id).Scan(
		&usuario.ID,
		&usuario.Nombre,
		&usuario.Email,
		&usuario.Telefono,
		&usuario.Rol,
		&usuario.Estado,
		&usuario.FechaRegistro,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return usuario, nil
}

// UpdatePassword actualiza la contraseña hasheada del usuario con RLS
func (r *UsuarioRepository) UpdatePassword(ctx context.Context, id int, hashedPassword string) error {
	_, err := r.db.Exec(ctx, id, `
		UPDATE usuarios SET password_hash = $1 WHERE id_usuario = $2
	`, hashedPassword, id)
	return err
}
