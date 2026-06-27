package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kajve/api-mobile/internal/application/interfaces"
	"github.com/kajve/api-mobile/internal/domain/entities"
)

const AuthContextKey = "auth_user"

// JWTAuth es el middleware de autenticación JWT
func JWTAuth(authService interfaces.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener token del header Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Esperamos formato "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Validar token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
				return
			}

			// Almacenar claims en el contexto
			ctx := context.WithValue(r.Context(), AuthContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extrae el user_id del contexto
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	claims, ok := ctx.Value(AuthContextKey).(*entities.JWTClaims)
	if !ok {
		return 0, false
	}
	return claims.UserID, true
}

// GetClaimsFromContext extrae los claims del contexto
func GetClaimsFromContext(ctx context.Context) (*entities.JWTClaims, bool) {
	claims, ok := ctx.Value(AuthContextKey).(*entities.JWTClaims)
	return claims, ok
}
