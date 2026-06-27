package interfaces

import (
	"context"

	"github.com/kajve/api-mobile/internal/domain/entities"
)

// UsuarioRepository define las operaciones para usuarios
type UsuarioRepository interface {
	// GetByEmail obtiene un usuario por email
	GetByEmail(ctx context.Context, email string) (*entities.Usuario, error)
	
	// GetByID obtiene un usuario por ID
	GetByID(ctx context.Context, id int) (*entities.Usuario, error)
	
	// Create crea un nuevo usuario
	Create(ctx context.Context, usuario *entities.Usuario) error
}

// LoteRepository define las operaciones para lotes de café
type LoteRepository interface {
	// GetByID obtiene un lote por ID
	GetByID(ctx context.Context, id int) (*entities.LoteCafe, error)
	
	// GetByUsuarioID obtiene todos los lotes de un usuario
	GetByUsuarioID(ctx context.Context, usuarioID int) ([]entities.LoteCafe, error)
	
	// Create crea un nuevo lote
	Create(ctx context.Context, lote *entities.LoteCafe) (int, error)
	
	// Update actualiza un lote
	Update(ctx context.Context, lote *entities.LoteCafe) error
}

// SensorRepository define las operaciones para sensores
type SensorRepository interface {
	// GetByESP32ID obtiene un sensor por su ID ESP32
	GetByESP32ID(ctx context.Context, esp32ID string) (*entities.Sensor, error)
	
	// GetByID obtiene un sensor por ID
	GetByID(ctx context.Context, id int) (*entities.Sensor, error)
	
	// Create crea un nuevo sensor
	Create(ctx context.Context, sensor *entities.Sensor) (int, error)
	
	// LinkToLote vincula un sensor a un lote
	LinkToLote(ctx context.Context, sensorID, loteID int) error
}

// ProvisioningTokenRepository define las operaciones para tokens de provisioning
type ProvisioningTokenRepository interface {
	// Create crea un nuevo token de provisioning
	Create(ctx context.Context, token *entities.ProvisioningToken) error
	
	// GetByToken obtiene un token por su valor
	GetByToken(ctx context.Context, tokenHash string) (*entities.ProvisioningToken, error)
	
	// MarkAsUsed marca un token como usado
	MarkAsUsed(ctx context.Context, tokenID int) error
}

// LecturaRepository define las operaciones para lecturas ambientales
type LecturaRepository interface {
	// GetLatestByLoteID obtiene las últimas N lecturas de un lote
	GetLatestByLoteID(ctx context.Context, loteID int, limit int) ([]entities.LecturaAmbiental, error)
	
	// Create crea una nueva lectura
	Create(ctx context.Context, lectura *entities.LecturaAmbiental) error
}

// AlertaRepository define las operaciones para alertas
type AlertaRepository interface {
	// GetByLoteID obtiene todas las alertas de un lote
	GetByLoteID(ctx context.Context, loteID int) ([]entities.Alerta, error)
	
	// Create crea una nueva alerta
	Create(ctx context.Context, alerta *entities.Alerta) error
}

// PrediccionRepository define las operaciones para predicciones
type PrediccionRepository interface {
	// GetByLoteID obtiene todas las predicciones de un lote
	GetByLoteID(ctx context.Context, loteID int) ([]entities.Prediccion, error)
	
	// Create crea una nueva predicción
	Create(ctx context.Context, prediccion *entities.Prediccion) error
}

// RecomendacionRepository define las operaciones para recomendaciones
type RecomendacionRepository interface {
	// GetByLoteID obtiene todas las recomendaciones de un lote
	GetByLoteID(ctx context.Context, loteID int) ([]entities.Recomendacion, error)
	
	// Create crea una nueva recomendación
	Create(ctx context.Context, recomendacion *entities.Recomendacion) error
}

// HistorialRepository define las operaciones para el historial
type HistorialRepository interface {
	// GetByLoteID obtiene todos los eventos del historial de un lote
	GetByLoteID(ctx context.Context, loteID int) ([]entities.HistorialEvento, error)
	
	// Create crea un nuevo evento en el historial
	Create(ctx context.Context, evento *entities.HistorialEvento) error
}

// ReporteRepository define las operaciones para reportes
type ReporteRepository interface {
	// GetByID obtiene un reporte por ID
	GetByID(ctx context.Context, id int) (*entities.Reporte, error)
	
	// Create crea un nuevo reporte
	Create(ctx context.Context, reporte *entities.Reporte) error
	
	// Update actualiza un reporte
	Update(ctx context.Context, reporte *entities.Reporte) error
}
