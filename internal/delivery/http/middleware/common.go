package middleware

import (
	"net/http"
)

// ErrorResponse es la estructura de respuesta de error
type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

// JSONContentType añade el header de Content-Type para JSON
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware configura CORS permitiendo una lista de orígenes.
// Como se usa AllowCredentials, el Access-Control-Allow-Origin no puede ser "*":
// se refleja el origen de la petición solo si está en la lista permitida.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Vary: Origin evita que caches/proxies compartan una respuesta
			// generada para un origen entre distintos orígenes.
			w.Header().Set("Vary", "Origin")

			if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
