package interfaces

import (
	"context"
	"time"

	"github.com/kajve/api-mobile/internal/domain/entities"
)

// UsuarioRepository define las operaciones para usuarios
type UsuarioRepository interface {
	GetByEmail(ctx context.Context, email string) (*entities.Usuario, error)
	GetByID(ctx context.Context, id int) (*entities.Usuario, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, usuario *entities.Usuario) error
	Update(ctx context.Context, id int, nombre, telefono string) (*entities.Usuario, error)
	UpdatePassword(ctx context.Context, id int, hashedPassword string) error
}

// LoteRepository define las operaciones para lotes de café
type LoteRepository interface {
	GetByID(ctx context.Context, id int) (*entities.LoteCafe, error)
	// GetByUsuarioID: limit=0 devuelve todos sin paginar
	GetByUsuarioID(ctx context.Context, usuarioID int, estado string, limit, offset int) ([]entities.LoteCafe, int, error)
	Create(ctx context.Context, lote *entities.LoteCafe) (*entities.LoteCafe, error)
	Update(ctx context.Context, loteID, usuarioID int, nombre, variedad string, pesoKg float64, ubicacion string) (*entities.LoteCafe, error)
	UpdateEstado(ctx context.Context, loteID, usuarioID int, estado string, fechaFin *time.Time) (*entities.LoteCafe, error)
}

// SensorRepository define las operaciones para sensores
type SensorRepository interface {
	GetByESP32ID(ctx context.Context, esp32ID string) (*entities.Sensor, error)
	// GetByIdentifier busca por esp32_id o mac_address
	GetByIdentifier(ctx context.Context, identifier string) (*entities.Sensor, error)
	GetByID(ctx context.Context, id int) (*entities.Sensor, error)
	Create(ctx context.Context, sensor *entities.Sensor) (int, error)
	LinkToLote(ctx context.Context, sensorID, loteID int) error
	MarcarTokenUsado(ctx context.Context, sensorID int) error
	// CountByUsuarioID cuenta sensores distintos vinculados a lotes del usuario (vía lotes_cafe.id_sensor)
	CountByUsuarioID(ctx context.Context, usuarioID int) (int, error)
}

// LecturaRepository define las operaciones para lecturas ambientales
type LecturaRepository interface {
	GetLatestByLoteID(ctx context.Context, loteID int, limit int) ([]entities.LecturaAmbiental, error)
	// GetByLoteIDFiltered: desde zero-value = sin filtro de fecha
	GetByLoteIDFiltered(ctx context.Context, loteID int, limit int, desde time.Time) ([]entities.LecturaAmbiental, error)
	GetEstadisticas(ctx context.Context, loteID int) (*entities.EstadisticasLote, error)
	Create(ctx context.Context, lectura *entities.LecturaAmbiental) error
}

// AlertaRepository define las operaciones para alertas
type AlertaRepository interface {
	GetByID(ctx context.Context, id int) (*entities.Alerta, error)
	// GetByLoteIDFiltered: atendida=nil sin filtro, nivel="" sin filtro
	GetByLoteIDFiltered(ctx context.Context, loteID int, atendida *bool, nivel string) ([]entities.Alerta, error)
	GetByLoteID(ctx context.Context, loteID int) ([]entities.Alerta, error)
	MarcarAtendida(ctx context.Context, alertaID int) (*entities.Alerta, error)
	Create(ctx context.Context, alerta *entities.Alerta) error
}

// PrediccionRepository define las operaciones para predicciones
type PrediccionRepository interface {
	GetByLoteID(ctx context.Context, loteID int) ([]entities.Prediccion, error)
	Create(ctx context.Context, prediccion *entities.Prediccion) error
}

// RecomendacionRepository define las operaciones para recomendaciones
type RecomendacionRepository interface {
	GetByLoteID(ctx context.Context, loteID int) ([]entities.Recomendacion, error)
	Create(ctx context.Context, recomendacion *entities.Recomendacion) error
}

// HistorialRepository define las operaciones para el historial
type HistorialRepository interface {
	GetByLoteIDPaginated(ctx context.Context, loteID int, limit, offset int) ([]entities.HistorialEvento, int, error)
	GetByLoteID(ctx context.Context, loteID int) ([]entities.HistorialEvento, error)
	Create(ctx context.Context, evento *entities.HistorialEvento) error
}

// ReporteRepository define las operaciones para reportes
type ReporteRepository interface {
	GetByID(ctx context.Context, id int) (*entities.Reporte, error)
	GetByUsuarioID(ctx context.Context, usuarioID int) ([]entities.Reporte, error)
	Create(ctx context.Context, reporte *entities.Reporte) error
	// UpdateURLArchivo actualiza la url del archivo generado (reportes no tiene columna "estado")
	UpdateURLArchivo(ctx context.Context, id int, urlArchivo string) error
}
