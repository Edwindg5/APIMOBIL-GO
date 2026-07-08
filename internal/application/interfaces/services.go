package interfaces

import (
	"context"
	"time"

	"github.com/kajve/api-mobile/internal/domain/entities"
)

// RegisterService define el caso de uso de registro de usuarios
type RegisterService interface {
	// Register crea un nuevo usuario con rol productor
	Register(ctx context.Context, nombre, email, password, telefono string) (*entities.RegisterResponse, error)
}

// ProfileService define los casos de uso de perfil del usuario autenticado
type ProfileService interface {
	// GetProfile retorna el perfil del usuario autenticado
	GetProfile(ctx context.Context, userID int) (*entities.PerfilResponse, error)

	// UpdateProfile actualiza nombre y telefono del usuario
	UpdateProfile(ctx context.Context, userID int, nombre, telefono string) (*entities.PerfilResponse, error)

	// ChangePassword valida la contraseña actual y establece la nueva
	ChangePassword(ctx context.Context, userID int, passwordActual, passwordNueva string) error
}

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
	LinkDevice(ctx context.Context, esp32ID, provisioningToken string, usuarioID int) (*entities.LinkDeviceResponse, error)
}

// LoteService define los casos de uso de lotes de café
type LoteService interface {
	// GetLotes lista lotes del usuario con filtro de estado y paginación
	GetLotes(ctx context.Context, usuarioID int, estado string, page, limit int) (*entities.LotesListResponse, error)

	// GetLoteDetalle retorna un lote con última lectura, alertas activas y última predicción
	GetLoteDetalle(ctx context.Context, loteID, usuarioID int) (*entities.LoteDetalle, error)

	// CreateLote crea un nuevo lote con codigo_qr automático
	CreateLote(ctx context.Context, req *entities.CreateLoteRequest, usuarioID int) (*entities.LoteCafe, error)

	// UpdateLote actualiza campos editables de un lote en estado 'en_proceso'
	UpdateLote(ctx context.Context, loteID, usuarioID int, req *entities.UpdateLoteRequest) (*entities.LoteCafe, error)

	// FinalizarLote cambia estado a 'finalizado' y registra evento en historial
	FinalizarLote(ctx context.Context, loteID, usuarioID int) (*entities.LoteCafe, error)

	// CancelarLote cambia estado a 'cancelado' (soft delete)
	CancelarLote(ctx context.Context, loteID, usuarioID int) error
}

// LecturaService define los casos de uso de lecturas
type LecturaService interface {
	GetLatestLecturas(ctx context.Context, loteID, usuarioID int, limit int) ([]entities.LecturaAmbiental, error)
	GetLecturasFiltradas(ctx context.Context, loteID, usuarioID int, limit int, desde time.Time) ([]entities.LecturaAmbiental, error)
	GetEstadisticas(ctx context.Context, loteID, usuarioID int) (*entities.EstadisticasLote, error)
}

// AlertaService define los casos de uso de alertas
type AlertaService interface {
	GetAlertas(ctx context.Context, loteID, usuarioID int, atendida *bool, nivel string) ([]entities.Alerta, error)
	AtenderAlerta(ctx context.Context, alertaID, usuarioID int) (*entities.Alerta, error)
}

// PrediccionService define los casos de uso de predicciones
type PrediccionService interface {
	GetPredicciones(ctx context.Context, loteID, usuarioID int) ([]entities.Prediccion, error)
}

// RecomendacionService define los casos de uso de recomendaciones
type RecomendacionService interface {
	GetRecomendaciones(ctx context.Context, loteID, usuarioID int) ([]entities.Recomendacion, error)
}

// HistorialService define los casos de uso del historial
type HistorialService interface {
	GetHistorial(ctx context.Context, loteID, usuarioID int, page, limit int) ([]entities.HistorialEvento, int, error)
}

// ReporteService define los casos de uso de reportes
type ReporteService interface {
	RequestReporte(ctx context.Context, req *entities.SolicitudReporteRequest, usuarioID int) (*entities.Reporte, error)
	GetReportes(ctx context.Context, usuarioID int) ([]entities.Reporte, error)
	// DescargarReporte valida la propiedad del reporte y devuelve la ruta en disco
	// y el nombre de archivo sugerido para la descarga.
	DescargarReporte(ctx context.Context, id, usuarioID int) (path string, fileName string, err error)
	// DeleteReporte elimina el reporte (registro y archivo en disco) del usuario
	DeleteReporte(ctx context.Context, id, usuarioID int) error
}

// DashboardService define los casos de uso del dashboard del productor
type DashboardService interface {
	GetDashboard(ctx context.Context, usuarioID int) (*entities.DashboardResponse, error)
}
