package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

type ProvisioningTokenRepository struct {
	db *PostgresDB
}

// NewProvisioningTokenRepository crea una nueva instancia del repositorio
func NewProvisioningTokenRepository(db *PostgresDB) interfaces.ProvisioningTokenRepository {
	return &ProvisioningTokenRepository{db: db}
}

// Create crea un nuevo token de provisioning
func (r *ProvisioningTokenRepository) Create(ctx context.Context, token *entities.ProvisioningToken) error {
	err := r.db.GetPool().QueryRow(ctx, `
		INSERT INTO provisioning_tokens (esp32_id, token, usuario_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`, token.ESP32ID, token.Token, token.UsuarioID, token.ExpiresAt).Scan(
		&token.ID,
		&token.CreatedAt,
	)

	return err
}

// GetByToken obtiene un token por su valor (hash)
func (r *ProvisioningTokenRepository) GetByToken(ctx context.Context, tokenHash string) (*entities.ProvisioningToken, error) {
	token := &entities.ProvisioningToken{}
	
	err := r.db.GetPool().QueryRow(ctx, `
		SELECT id, esp32_id, token, usuario_id, used_at, expires_at, created_at
		FROM provisioning_tokens
		WHERE token = $1 AND used_at IS NULL AND expires_at > NOW()
	`, tokenHash).Scan(
		&token.ID,
		&token.ESP32ID,
		&token.Token,
		&token.UsuarioID,
		&token.UsedAt,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return token, nil
}

// MarkAsUsed marca un token como usado
func (r *ProvisioningTokenRepository) MarkAsUsed(ctx context.Context, tokenID int) error {
	now := time.Now()
	_, err := r.db.GetPool().Exec(ctx, `
		UPDATE provisioning_tokens
		SET used_at = $1
		WHERE id = $2
	`, now, tokenID)

	return err
}
