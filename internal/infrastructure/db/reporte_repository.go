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

func NewReporteRepository(db *PostgresDB) interfaces.ReporteRepository {
	return &ReporteRepository{db: db}
}

const reporteCols = `id_reporte, id_lote, id_usuario, tipo_reporte, formato, url_archivo, fecha_generacion`

func scanReporte(row interface{ Scan(...any) error }, rep *entities.Reporte) error {
	return row.Scan(
		&rep.ID, &rep.LoteID, &rep.UsuarioID, &rep.TipoReporte,
		&rep.Formato, &rep.URLArchivo, &rep.FechaGeneracion,
	)
}

func (r *ReporteRepository) GetByID(ctx context.Context, id int) (*entities.Reporte, error) {
	rep := &entities.Reporte{}
	err := scanReporte(r.db.GetPool().QueryRow(ctx,
		`SELECT `+reporteCols+` FROM reportes WHERE id_reporte = $1`, id), rep)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rep, nil
}

func (r *ReporteRepository) GetByUsuarioID(ctx context.Context, usuarioID int) ([]entities.Reporte, error) {
	rows, err := r.db.GetPool().Query(ctx, `
		SELECT `+reporteCols+`
		FROM reportes
		WHERE id_usuario = $1
		ORDER BY fecha_generacion DESC
	`, usuarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reportes []entities.Reporte
	for rows.Next() {
		var rep entities.Reporte
		if err := scanReporte(rows, &rep); err != nil {
			return nil, err
		}
		reportes = append(reportes, rep)
	}
	return reportes, rows.Err()
}

func (r *ReporteRepository) Create(ctx context.Context, rep *entities.Reporte) error {
	return scanReporte(r.db.GetPool().QueryRow(ctx, `
		INSERT INTO reportes (id_lote, id_usuario, tipo_reporte, formato, fecha_generacion)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING `+reporteCols,
		rep.LoteID, rep.UsuarioID, rep.TipoReporte, rep.Formato,
	), rep)
}

// UpdateURLArchivo actualiza la url del archivo generado (reportes no tiene columna "estado")
func (r *ReporteRepository) UpdateURLArchivo(ctx context.Context, id int, urlArchivo string) error {
	_, err := r.db.GetPool().Exec(ctx, `
		UPDATE reportes SET url_archivo = $1 WHERE id_reporte = $2
	`, urlArchivo, id)
	return err
}

// CountByUsuarioID cuenta cuántos reportes tiene generados un usuario.
func (r *ReporteRepository) CountByUsuarioID(ctx context.Context, usuarioID int) (int, error) {
	var count int
	err := r.db.GetPool().QueryRow(ctx,
		`SELECT COUNT(*) FROM reportes WHERE id_usuario = $1`, usuarioID,
	).Scan(&count)
	return count, err
}

// Delete elimina el registro del reporte de la base de datos.
func (r *ReporteRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.GetPool().Exec(ctx, `DELETE FROM reportes WHERE id_reporte = $1`, id)
	return err
}
