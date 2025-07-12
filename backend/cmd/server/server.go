// internal/server/server.go
// HTTP server setup with dependency injection

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"tournament-planner/internal/api"
	"tournament-planner/internal/config"
	"tournament-planner/internal/database"
	"tournament-planner/internal/middleware"
	"tournament-planner/internal/services"
	"tournament-planner/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	config   *config.Config
	router   *gin.Engine
	services *services.Container
	logger   *log.Logger
	server   *http.Server
}

// New creates a new server with all dependencies
func New(cfg *config.Config, db *database.Connections, logger *log.Logger) *Server {
	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create service container with all business logic
	serviceContainer := services.NewContainer(db, cfg, logger)

	// Create router with middleware
	router := setupRouter(cfg, serviceContainer, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		config:   cfg,
		router:   router,
		services: serviceContainer,
		logger:   logger,
		server:   srv,
	}
}

// setupRouter configures all routes and middleware
func setupRouter(cfg *config.Config, services *services.Container, logger *log.Logger) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimiter(services.Cache))

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.External.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12 hours
	}))

	// Maintenance mode middleware
	if cfg.Features.MaintenanceMode {
		router.Use(middleware.MaintenanceMode())
	}

	// Health check (always available)
	router.GET("/health", api.HealthCheck(cfg))

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Mount all route groups
		api.RegisterAuthRoutes(v1, services)
		api.RegisterUserRoutes(v1, services)
		api.RegisterTournamentRoutes(v1, services)
		api.RegisterMatchRoutes(v1, services)
		api.RegisterPaymentRoutes(v1, services, cfg)
		api.RegisterAdminRoutes(v1, services)
	}

	// WebSocket endpoint (if enabled)
	if cfg.Features.EnableWebSocket {
		hub := websocket.NewHub(services, logger)
		go hub.Run()
		router.GET("/ws", middleware.OptionalAuth(services.Auth), websocket.HandleConnection(hub))
	}

	// Static file serving
	router.Static("/uploads", cfg.External.UploadPath)

	return router
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Println("Shutting down server...")
	return s.server.Shutdown(ctx)
}
