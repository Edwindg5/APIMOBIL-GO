package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kajve/api-mobile/config"
	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	config              *config.Config
	usuarioRepository   interfaces.UsuarioRepository
}

// NewAuthService crea una nueva instancia del servicio
func NewAuthService(
	cfg *config.Config,
	usuarioRepository interfaces.UsuarioRepository,
) interfaces.AuthService {
	return &AuthService{
		config:            cfg,
		usuarioRepository: usuarioRepository,
	}
}

// Login autentica un usuario y retorna tokens JWT
func (s *AuthService) Login(ctx context.Context, email, password string) (*entities.LoginResponse, error) {
	// Obtener usuario por email
	usuario, err := s.usuarioRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	if usuario == nil {
		return nil, errors.New("invalid email or password")
	}

	// Validar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(usuario.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generar tokens
	accessToken, refreshToken, err := s.GenerateTokens(usuario.ID, usuario.Email, usuario.Rol)
	if err != nil {
		return nil, fmt.Errorf("error generating tokens: %w", err)
	}

	return &entities.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWTExpirationHours.Seconds()),
		Usuario: entities.UsuarioPublicInfo{
			ID:             usuario.ID,
			Email:          usuario.Email,
			NombreCompleto: usuario.Nombre,
			Rol:            usuario.Rol,
		},
	}, nil
}

// RefreshAccessToken genera un nuevo access token desde un refresh token
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*entities.RefreshTokenResponse, error) {
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims == nil {
		return nil, errors.New("invalid refresh token")
	}

	// Generar nuevo access token
	accessToken, _, err := s.GenerateTokens(claims.UserID, claims.Email, claims.Rol)
	if err != nil {
		return nil, fmt.Errorf("error generating tokens: %w", err)
	}

	return &entities.RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(s.config.JWTExpirationHours.Seconds()),
	}, nil
}

// ValidateToken valida un token JWT y retorna los claims
func (s *AuthService) ValidateToken(token string) (*entities.JWTClaims, error) {
	claims := &entities.JWTClaims{}

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		// Validar algoritmo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, errors.New("token is not valid")
	}

	return claims, nil
}

// GenerateTokens genera access y refresh tokens
func (s *AuthService) GenerateTokens(userID int, email, rol string) (accessToken, refreshToken string, err error) {
	// Access Token
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"rol":     rol,
		"exp":     time.Now().Add(s.config.JWTExpirationHours).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", fmt.Errorf("error signing access token: %w", err)
	}

	// Refresh Token
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"rol":     rol,
		"exp":     time.Now().Add(s.config.JWTRefreshExpirationHours).Unix(),
		"iat":     time.Now().Unix(),
		"type":    "refresh",
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", fmt.Errorf("error signing refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// HashPassword genera el hash de una contraseña
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// HashTokenForStorage genera un hash SHA256 para almacenar tokens de provisioning
func HashTokenForStorage(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
