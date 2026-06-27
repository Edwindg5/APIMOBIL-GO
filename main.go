package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kajve/api-mobile/config"
	"github.com/kajve/api-mobile/internal/application/usecases"
	"github.com/kajve/api-mobile/internal/delivery/http/handlers"
	httpmiddleware "github.com/kajve/api-mobile/internal/delivery/http/middleware"
	"github.com/kajve/api-mobile/internal/infrastructure/db"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Conectar a la base de datos
	postgres, err := db.NewPostgresDB(cfg.DBConnString())
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer postgres.Close()

	// Inicializar repositorios
	usuarioRepo := db.NewUsuarioRepository(postgres)
	sensorRepo := db.NewSensorRepository(postgres)
	provisioningTokenRepo := db.NewProvisioningTokenRepository(postgres)
	loteRepo := db.NewLoteRepository(postgres)

	// Inicializar repositorios adicionales
	lecturaRepo := db.NewLecturaRepository(postgres)
	alertaRepo := db.NewAlertaRepository(postgres)
	prediccionRepo := db.NewPrediccionRepository(postgres)
	recomendacionRepo := db.NewRecomendacionRepository(postgres)
	historialRepo := db.NewHistorialRepository(postgres)
	reporteRepo := db.NewReporteRepository(postgres)

	// Inicializar servicios (usecases)
	authService := usecases.NewAuthService(cfg, usuarioRepo)
	deviceService := usecases.NewDeviceService(sensorRepo, loteRepo, provisioningTokenRepo)
	loteService := usecases.NewLoteService(loteRepo)
	lecturaService := usecases.NewLecturaService(lecturaRepo, loteRepo)
	alertaService := usecases.NewAlertaService(alertaRepo, loteRepo)
	prediccionService := usecases.NewPrediccionService(prediccionRepo, loteRepo)
	recomendacionService := usecases.NewRecomendacionService(recomendacionRepo, loteRepo)
	historialService := usecases.NewHistorialService(historialRepo, loteRepo)
	reporteService := usecases.NewReporteService(reporteRepo, loteRepo)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)
	deviceHandler := handlers.NewDeviceHandler(deviceService)
	loteHandler := handlers.NewLoteHandler(loteService)
	lecturaHandler := handlers.NewLecturaHandler(lecturaService)
	alertaHandler := handlers.NewAlertaHandler(alertaService)
	prediccionHandler := handlers.NewPrediccionHandler(prediccionService)
	recomendacionHandler := handlers.NewRecomendacionHandler(recomendacionService)
	historialHandler := handlers.NewHistorialHandler(historialService)
	reporteHandler := handlers.NewReporteHandler(reporteService)

	// Crear router chi
	router := chi.NewRouter()

	// Middleware global
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(httpmiddleware.JSONContentType)
	router.Use(httpmiddleware.CORSMiddleware(cfg.CORSAllowedOrigin))

	// Rate limiting
	rateLimiter := httpmiddleware.NewRateLimiter(cfg.RateLimitReqPerMin)
	router.Use(rateLimiter.Middleware())

	// Rutas públicas (sin autenticación)
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Rutas protegidas (con autenticación JWT)
	router.Route("/", func(r chi.Router) {
		r.Use(httpmiddleware.JWTAuth(authService))

		// Devices
		r.Route("/devices", func(r chi.Router) {
			r.Post("/link", deviceHandler.LinkDevice)
		})

		// Lotes
		r.Route("/lotes", func(r chi.Router) {
			r.Get("/", loteHandler.GetLotes)
			r.Post("/", loteHandler.CreateLote)
			r.Get("/{id}", loteHandler.GetLote)
			
			// Lecturas ambientales
			r.Get("/{id}/lecturas", lecturaHandler.GetLecturas)
			
			// Alertas
			r.Get("/{id}/alertas", alertaHandler.GetAlertas)
			
			// Predicciones
			r.Get("/{id}/predicciones", prediccionHandler.GetPredicciones)
			
			// Recomendaciones
			r.Get("/{id}/recomendaciones", recomendacionHandler.GetRecomendaciones)
			
			// Historial
			r.Get("/{id}/historial", historialHandler.GetHistorial)
		})

		// Reportes
		r.Route("/reportes", func(r chi.Router) {
			r.Post("/", reporteHandler.RequestReporte)
			r.Get("/{id}", reporteHandler.GetReporte)
		})
	})

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Iniciar servidor
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
