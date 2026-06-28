package entities

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Usuario representa un usuario del sistema
type Usuario struct {
	ID        int       `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	NombreCompleto string `db:"nombre_completo"`
	Rol       string    `db:"rol"` // 'productor', 'admin', etc.
	Estado    string    `db:"estado"` // 'activo', 'inactivo'
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// LoginRequest es la solicitud de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse es la respuesta de login
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // segundos
	Usuario      UsuarioPublicInfo `json:"usuario"`
}

// UsuarioPublicInfo es la información pública del usuario
type UsuarioPublicInfo struct {
	ID             int    `json:"id"`
	Email          string `json:"email"`
	NombreCompleto string `json:"nombre_completo"`
	Rol            string `json:"rol"`
}

// RefreshTokenRequest es la solicitud para renovar token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse es la respuesta de renovación
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// JWTClaims son los claims del JWT
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Rol    string `json:"rol"`
}

// ProvisioningToken representa un token de provisioning de un solo uso
type ProvisioningToken struct {
	ID        int       `db:"id"`
	ESP32ID   string    `db:"esp32_id"`
	Token     string    `db:"token"` // hash del token original
	UsuarioID int       `db:"usuario_id"`
	UsedAt    *time.Time `db:"used_at"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// LinkDeviceRequest es la solicitud para vincular un dispositivo
type LinkDeviceRequest struct {
	ESP32ID           string `json:"esp32_id" validate:"required"`
	ProvisioningToken string `json:"provisioning_token" validate:"required"`
	LoteName          string `json:"lote_name" validate:"required"`
}

// LinkDeviceResponse es la respuesta de vinculación
type LinkDeviceResponse struct {
	SensorID  int       `json:"sensor_id"`
	LoteID    int       `json:"lote_id"`
	Message   string    `json:"message"`
	LinkedAt  time.Time `json:"linked_at"`
}

// Sensor representa un ESP32
type Sensor struct {
	ID        int        `db:"id"`
	ESP32ID   string     `db:"esp32_id"`
	LoteID    *int       `db:"lote_id"` // NULL hasta que se vincule
	LinkedAt  *time.Time `db:"linked_at"`
	LastSeen  *time.Time `db:"last_seen"`
	Estado    string     `db:"estado"` // 'activo', 'inactivo'
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

// LoteCafe representa un lote de café
type LoteCafe struct {
	ID          int        `db:"id"`
	UsuarioID   int        `db:"usuario_id"`
	Nombre      string     `db:"nombre"`
	Descripcion string     `db:"descripcion"`
	Area        float64    `db:"area"`
	SensorID    *int       `db:"sensor_id"`
	Estado      string     `db:"estado"` // 'activo', 'cosechado', 'inactivo'
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// LoteListItem es un item en la lista de lotes
type LoteListItem struct {
	ID          int        `json:"id"`
	Nombre      string     `json:"nombre"`
	Descripcion string     `json:"descripcion"`
	Area        float64    `json:"area"`
	Estado      string     `json:"estado"`
	SensorID    *int       `json:"sensor_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// LecturAmbiental representa una lectura ambiental
type LecturaAmbiental struct {
	ID        int       `db:"id"`
	LoteID    int       `db:"lote_id"`
	SensorID  int       `db:"sensor_id"`
	Temperatura float64 `db:"temperatura"`
	Humedad   float64   `db:"humedad"`
	Presion   float64   `db:"presion"`
	CreatedAt time.Time `db:"created_at"`
}

// Alerta representa una alerta para un lote
type Alerta struct {
	ID        int       `db:"id"`
	LoteID    int       `db:"lote_id"`
	Tipo      string    `db:"tipo"` // 'temperatura', 'humedad', etc.
	Mensaje   string    `db:"mensaje"`
	Nivel     string    `db:"nivel"` // 'info', 'warning', 'critical'
	Leida     bool      `db:"leida"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Prediccion representa una predicción del ml-service
type Prediccion struct {
	ID        int       `db:"id"`
	LoteID    int       `db:"lote_id"`
	Prediccion string   `db:"prediccion"`
	Probabilidad float64 `db:"probabilidad"`
	CreatedAt time.Time `db:"created_at"`
}

// Recomendacion representa una recomendación
type Recomendacion struct {
	ID        int       `db:"id"`
	LoteID    int       `db:"lote_id"`
	Accion    string    `db:"accion"`
	Razon     string    `db:"razon"`
	Prioridad string    `db:"prioridad"` // 'baja', 'media', 'alta'
	CreatedAt time.Time `db:"created_at"`
}

// HistorialEvento representa un evento en el historial
type HistorialEvento struct {
	ID        int       `db:"id"`
	LoteID    int       `db:"lote_id"`
	Tipo      string    `db:"tipo"`
	Descripcion string `db:"descripcion"`
	CreatedAt time.Time `db:"created_at"`
}

// Reporte representa un reporte solicitado
type Reporte struct {
	ID        int       `db:"id"`
	LoteID    int       `db:"lote_id"`
	UsuarioID int       `db:"usuario_id"`
	Tipo      string    `db:"tipo"` // 'diario', 'semanal', 'mensual'
	Estado    string    `db:"estado"` // 'pendiente', 'procesando', 'completado'
	URL       *string   `db:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ErrorResponse es la respuesta de error
type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}
