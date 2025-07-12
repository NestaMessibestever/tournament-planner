// cmd/server/main.go
// This is the main entry point for the Tournament Planner backend server.
// It initializes all dependencies and starts the HTTP server.

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tournament-planner/internal/config"
	"tournament-planner/internal/database"
	"tournament-planner/internal/server"
)

func main() {
	// Load configuration from environment variables and config files
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up structured logging based on environment
	logger := setupLogger(cfg.Environment)

	// Initialize database connections with retry logic
	dbConnections, err := initializeDatabases(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize databases: %v", err)
	}
	defer dbConnections.Close()

	// Create and configure the HTTP server with all dependencies
	srv := server.New(cfg, dbConnections, logger)

	// Start server in a goroutine to allow for graceful shutdown
	go func() {
		logger.Printf("Starting server on port %s in %s mode", cfg.Server.Port, cfg.Environment)
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	gracefulShutdown(srv, logger)
}

// initializeDatabases sets up all database connections with health checks
func initializeDatabases(cfg *config.Config, logger *log.Logger) (*database.Connections, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return database.Initialize(ctx, database.Config{
		MySQL: database.MySQLConfig{
			DSN:             cfg.Database.MySQL.DSN,
			MaxOpenConns:    cfg.Database.MySQL.MaxOpenConns,
			MaxIdleConns:    cfg.Database.MySQL.MaxIdleConns,
			ConnMaxLifetime: cfg.Database.MySQL.ConnMaxLifetime,
		},
		MongoDB: database.MongoConfig{
			URI:      cfg.Database.MongoDB.URI,
			Database: cfg.Database.MongoDB.Database,
		},
		Redis: database.RedisConfig{
			Addr:     cfg.Database.Redis.Addr,
			Password: cfg.Database.Redis.Password,
			DB:       cfg.Database.Redis.DB,
		},
	}, logger)
}

// setupLogger configures structured logging based on the environment
func setupLogger(env string) *log.Logger {
	// In production, you might want to use a more sophisticated logger
	// like zap or logrus for structured logging
	logger := log.New(os.Stdout, "[tournament-planner] ", log.LstdFlags|log.Lshortfile)

	if env == "production" {
		// In production, you might want to:
		// - Output JSON formatted logs
		// - Send logs to a centralized logging service
		// - Set appropriate log levels
	}

	return logger
}

// gracefulShutdown handles graceful shutdown of the server
func gracefulShutdown(srv *server.Server, logger *log.Logger) {
	quit := make(chan os.Signal, 1)
	// Listen for interrupt signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Printf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exited")
}
