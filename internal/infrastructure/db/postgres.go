package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	pool *pgxpool.Pool
}

// NewPostgresDB crea una nueva conexión a PostgreSQL
func NewPostgresDB(connString string) (*PostgresDB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing database config: %w", err)
	}

	// Neon usa PgBouncer en modo transacción (-pooler): los prepared statements
	// nombrados (modo por defecto) chocan porque el backend físico puede cambiar
	// entre queries (SQLSTATE 08P01). QueryExecModeExec usa statements sin nombre
	// (no cacheados) pero mantiene el binding de parámetros server-side vía Bind,
	// sin interpolación client-side. No soporta múltiples sentencias en un mismo
	// string (ver injectUserID en usuario_repository.go:148 / UpdatePassword,
	// bug preexistente e independiente de este cambio).
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	// Configuración de pool
	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("error creating database pool: %w", err)
	}

	// Verificar conexión
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

// QueryRow ejecuta una query que retorna una fila
func (db *PostgresDB) QueryRow(ctx context.Context, userID int, sql string, args ...any) pgx.Row {
	return db.pool.QueryRow(ctx, db.injectUserID(userID, sql), args...)
}

// Query ejecuta una query que retorna múltiples filas
func (db *PostgresDB) Query(ctx context.Context, userID int, sql string, args ...any) (pgx.Rows, error) {
	return db.pool.Query(ctx, db.injectUserID(userID, sql), args...)
}

// Exec ejecuta una query sin retornar filas
func (db *PostgresDB) Exec(ctx context.Context, userID int, sql string, args ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, db.injectUserID(userID, sql), args...)
}

// BeginTx inicia una transacción
func (db *PostgresDB) BeginTx(ctx context.Context, userID int) (pgx.Tx, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// Inyectar app.current_user_id en la transacción
	_, err = tx.Exec(ctx, fmt.Sprintf("SET app.current_user_id = %d", userID))
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	return tx, nil
}

// injectUserID inyecta el SET app.current_user_id al principio de la query
func (db *PostgresDB) injectUserID(userID int, sql string) string {
	return fmt.Sprintf("SET app.current_user_id = %d; %s", userID, sql)
}

// Close cierra la conexión a la base de datos
func (db *PostgresDB) Close() {
	db.pool.Close()
}

// GetPool retorna el pool directo para casos especiales
func (db *PostgresDB) GetPool() *pgxpool.Pool {
	return db.pool
}
