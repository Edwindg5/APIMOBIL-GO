package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port        int
	Environment string

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret                 string
	JWTExpirationHours        time.Duration
	JWTRefreshExpirationHours time.Duration

	// CORS
	CORSAllowedOrigins []string

	// API
	APIWebURL     string
	MQTTBrokerURL string
	MQTTUsername  string
	MQTTPassword  string

	// Rate Limiting
	RateLimitReqPerMin int
	RateLimitBurst     int

	// Reportes
	ReportsDir string

	// Lotes reclamados por QR (creados como placeholder por api-web)
	PlaceholderLoteUserID int
}

func Load() (*Config, error) {
	// Cargar .env si existe
	_ = godotenv.Load()

	cfg := &Config{
		Port:                      getEnvInt("PORT", 8080),
		Environment:               getEnv("ENVIRONMENT", "development"),
		DBHost:                    getEnv("DB_HOST", "localhost"),
		DBPort:                    getEnvInt("DB_PORT", 5432),
		DBUser:                    getEnv("DB_USER", "kajve_user"),
		DBPassword:                getEnv("DB_PASSWORD", "kajve_secure_password"),
		DBName:                    getEnv("DB_NAME", "kajve_db"),
		DBSSLMode:                 getEnv("DB_SSL_MODE", "disable"),
		RedisHost:                 getEnv("REDIS_HOST", "localhost"),
		RedisPort:                 getEnvInt("REDIS_PORT", 6379),
		RedisPassword:             getEnv("REDIS_PASSWORD", ""),
		RedisDB:                   getEnvInt("REDIS_DB", 0),
		JWTSecret:                 getEnv("JWT_SECRET", "your_super_secret_jwt_key_change_in_production_min_32_chars"),
		JWTExpirationHours:        time.Duration(getEnvInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,
		JWTRefreshExpirationHours: time.Duration(getEnvInt("JWT_REFRESH_EXPIRATION_HOURS", 720)) * time.Hour,
		CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{
			"http://localhost:4200",
			"http://movil-kajve.dnc-ed-denz.shop",
			"https://movil-kajve.dnc-ed-denz.shop",
			"http://dnc-ed-denz.shop",
			"https://dnc-ed-denz.shop",
		}),
		APIWebURL:          getEnv("API_WEB_URL", "http://api-web:3001"),
		MQTTBrokerURL:      getEnv("MQTT_BROKER_URL", "mqtt://mosquitto:1883"),
		MQTTUsername:       getEnv("MQTT_USERNAME", "kajve"),
		MQTTPassword:       getEnv("MQTT_PASSWORD", "kajve_password"),
		RateLimitReqPerMin: getEnvInt("RATE_LIMIT_REQ_PER_MIN", 100),
		RateLimitBurst:     getEnvInt("RATE_LIMIT_BURST", 20),
		ReportsDir:         getEnv("REPORTS_DIR", "./storage/reportes"),
	}

	placeholderLoteUserID, err := getEnvIntRequired("PLACEHOLDER_LOTE_USER_ID")
	if err != nil {
		return nil, err
	}
	cfg.PlaceholderLoteUserID = placeholderLoteUserID

	return cfg, nil
}

func (c *Config) DBConnString() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.DBSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvSlice lee una variable de entorno con valores separados por comas
// (ej: "http://a.com,http://b.com") y la convierte en un slice de strings,
// recortando espacios y descartando entradas vacías.
func getEnvSlice(key string, defaultValue []string) []string {
	valueStr, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(valueStr) == "" {
		return defaultValue
	}

	parts := strings.Split(valueStr, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

func getEnvInt(key string, defaultValue int) int {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvIntRequired lee una variable de entorno entera sin valor por defecto:
// falla explícito si falta o no es numérica, en vez de asumir un valor.
func getEnvIntRequired(key string) (int, error) {
	valueStr, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(valueStr) == "" {
		return 0, fmt.Errorf("required environment variable %s is not set", key)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s must be an integer: %w", key, err)
	}
	return value, nil
}
