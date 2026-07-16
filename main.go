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
	"github.com/kajve/api-mobile/internal/infrastructure/reportgen"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	postgres, err := db.NewPostgresDB(cfg.DBConnString())
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer postgres.Close()

	// Repositorios
	usuarioRepo := db.NewUsuarioRepository(postgres)
	sensorRepo := db.NewSensorRepository(postgres)
	loteRepo := db.NewLoteRepository(postgres)
	lecturaRepo := db.NewLecturaRepository(postgres)
	alertaRepo := db.NewAlertaRepository(postgres)
	prediccionRepo := db.NewPrediccionRepository(postgres)
	recomendacionRepo := db.NewRecomendacionRepository(postgres)
	historialRepo := db.NewHistorialRepository(postgres)
	reporteRepo := db.NewReporteRepository(postgres)

	// Generador de reportes (PDF/Excel)
	reporteCollector := reportgen.NewCollector(lecturaRepo, alertaRepo, prediccionRepo, recomendacionRepo, historialRepo)
	reporteGenerator := reportgen.NewGenerator(reporteCollector, cfg.ReportsDir)

	// Servicios
	authService := usecases.NewAuthService(cfg, usuarioRepo)
	registerService := usecases.NewRegisterService(usuarioRepo)
	profileService := usecases.NewProfileService(usuarioRepo)
	deviceService := usecases.NewDeviceService(sensorRepo, loteRepo, historialRepo)
	loteService := usecases.NewLoteService(loteRepo, historialRepo, lecturaRepo, alertaRepo, prediccionRepo, cfg.PlaceholderLoteUserID)
	lecturaService := usecases.NewLecturaService(lecturaRepo, loteRepo)
	alertaService := usecases.NewAlertaService(alertaRepo, loteRepo, historialRepo)
	prediccionService := usecases.NewPrediccionService(prediccionRepo, loteRepo)
	recomendacionService := usecases.NewRecomendacionService(recomendacionRepo, loteRepo)
	historialService := usecases.NewHistorialService(historialRepo, loteRepo)
	reporteService := usecases.NewReporteService(reporteRepo, loteRepo, usuarioRepo, reporteGenerator, cfg.ReportsDir)
	dashboardService := usecases.NewDashboardService(loteRepo, alertaRepo, lecturaRepo, prediccionRepo, reporteRepo, sensorRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, registerService)
	profileHandler := handlers.NewProfileHandler(profileService)
	deviceHandler := handlers.NewDeviceHandler(deviceService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	loteHandler := handlers.NewLoteHandler(loteService)
	lecturaHandler := handlers.NewLecturaHandler(lecturaService)
	alertaHandler := handlers.NewAlertaHandler(alertaService)
	prediccionHandler := handlers.NewPrediccionHandler(prediccionService)
	recomendacionHandler := handlers.NewRecomendacionHandler(recomendacionService)
	historialHandler := handlers.NewHistorialHandler(historialService)
	reporteHandler := handlers.NewReporteHandler(reporteService)

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(httpmiddleware.JSONContentType)
	router.Use(httpmiddleware.CORSMiddleware(cfg.CORSAllowedOrigins))

	rateLimiter := httpmiddleware.NewRateLimiter(cfg.RateLimitReqPerMin, cfg.RateLimitBurst)
	router.Use(rateLimiter.Middleware())

	// Rutas públicas
	router.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/register", authHandler.Register)
	})

	// Rutas protegidas
	router.Route("/", func(r chi.Router) {
		r.Use(httpmiddleware.JWTAuth(authService))

		// Perfil
		r.Route("/perfil", func(r chi.Router) {
			r.Get("/", profileHandler.GetPerfil)
			r.Put("/", profileHandler.UpdatePerfil)
			r.Put("/password", profileHandler.ChangePassword)
		})

		// Dashboard
		r.Get("/dashboard", dashboardHandler.GetDashboard)

		// Devices
		r.Post("/devices/link", deviceHandler.LinkDevice)

		// Alertas (acción individual)
		r.Put("/alertas/{id}/atender", alertaHandler.AtenderAlerta)

		// Lotes
		r.Route("/lotes", func(r chi.Router) {
			r.Get("/", loteHandler.GetLotes)
			r.Post("/", loteHandler.CreateLote)
			r.Put("/reclamar", loteHandler.ReclamarLote)
			r.Get("/{id}", loteHandler.GetLote)
			r.Put("/{id}", loteHandler.UpdateLote)
			r.Delete("/{id}", loteHandler.CancelarLote)
			r.Put("/{id}/finalizar", loteHandler.FinalizarLote)
			r.Get("/{id}/qr", loteHandler.GetQR)
			r.Get("/{id}/lecturas", lecturaHandler.GetLecturas)
			r.Get("/{id}/estadisticas", lecturaHandler.GetEstadisticas)
			r.Get("/{id}/alertas", alertaHandler.GetAlertas)
			r.Get("/{id}/predicciones", prediccionHandler.GetPredicciones)
			r.Get("/{id}/recomendaciones", recomendacionHandler.GetRecomendaciones)
			r.Get("/{id}/historial", historialHandler.GetHistorial)
		})

		// Reportes
		r.Route("/reportes", func(r chi.Router) {
			r.Post("/", reporteHandler.RequestReporte)
			r.Get("/", reporteHandler.GetReportes)
			r.Get("/{id}/descargar", reporteHandler.DescargarReporte)
			r.Delete("/{id}", reporteHandler.DeleteReporte)
		})
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
