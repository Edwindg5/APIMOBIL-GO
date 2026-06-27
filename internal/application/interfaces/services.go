package interfaces

import (
	"context"

	"github.com/kajve/api-mobile/internal/domain/entities"
)

// AuthService define los casos de uso de autenticación
type AuthService interface {
	// Login autentica un usuario y retorna tokens JWT
	Login(ctx context.Context, email, password string) (*entities.LoginResponse, error)
	
	// RefreshAccessToken genera un nuevo access token
	RefreshAccessToken(ctx context.Context, refreshToken string) (*entities.RefreshTokenResponse, error)
	
	// ValidateToken valida un token JWT y retorna los claims
	ValidateToken(token string) (*entities.JWTClaims, error)
	
	// GenerateTokens genera access y refresh tokens
	GenerateTokens(userID int, email, rol string) (accessToken, refreshToken string, err error)
}

// DeviceService define los casos de uso de dispositivos
type DeviceService interface {
	// LinkDevice vincula un ESP32 a un usuario usando un provisioning token
	LinkDevice(ctx context.Context, esp32ID, provisioningToken, loteName string, usuarioID int) (*entities.LinkDeviceResponse, error)
}

// LoteService define los casos de uso de lotes
type LoteService interface {
	// GetLotes obtiene todos los lotes del usuario
	GetLotes(ctx context.Context, usuarioID int) ([]entities.LoteListItem, error)
	
	// GetLote obtiene un lote específico
	GetLote(ctx context.Context, loteID, usuarioID int) (*entities.LoteCafe, error)
	
	// CreateLote crea un nuevo lote
	CreateLote(ctx context.Context, nombre, descripcion string, area float64, usuarioID int) (int, error)
}

// LecturaService define los casos de uso de lecturas
type LecturaService interface {
	// GetLatestLecturas obtiene las últimas lecturas de un lote
	GetLatestLecturas(ctx context.Context, loteID, usuarioID int, limit int) ([]entities.LecturaAmbiental, error)
}

// AlertaService define los casos de uso de alertas
type AlertaService interface {
	// GetAlertas obtiene todas las alertas de un lote
	GetAlertas(ctx context.Context, loteID, usuarioID int) ([]entities.Alerta, error)
}

// PrediccionService define los casos de uso de predicciones
type PrediccionService interface {
	// GetPredicciones obtiene las predicciones de un lote
	GetPredicciones(ctx context.Context, loteID, usuarioID int) ([]entities.Prediccion, error)
}

// RecomendacionService define los casos de uso de recomendaciones
type RecomendacionService interface {
	// GetRecomendaciones obtiene las recomendaciones de un lote
	GetRecomendaciones(ctx context.Context, loteID, usuarioID int) ([]entities.Recomendacion, error)
}

// HistorialService define los casos de uso del historial
type HistorialService interface {
	// GetHistorial obtiene el historial de un lote
	GetHistorial(ctx context.Context, loteID, usuarioID int) ([]entities.HistorialEvento, error)
}

// ReporteService define los casos de uso de reportes
type ReporteService interface {
	// RequestReporte solicita la generación de un reporte
	RequestReporte(ctx context.Context, loteID, usuarioID int, tipoReporte string) (int, error)
}
