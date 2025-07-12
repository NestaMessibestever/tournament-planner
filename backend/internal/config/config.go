// internal/config/config.go
// Configuration management using environment variables and optional config files

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Auth        AuthConfig
	External    ExternalConfig
	Features    FeatureFlags
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig contains all database connection settings
type DatabaseConfig struct {
	MySQL   MySQLConfig
	MongoDB MongoDBConfig
	Redis   RedisConfig
}

// MySQLConfig contains MySQL-specific settings
type MySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// MongoDBConfig contains MongoDB-specific settings
type MongoDBConfig struct {
	URI      string
	Database string
}

// RedisConfig contains Redis-specific settings
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// AuthConfig contains authentication and authorization settings
type AuthConfig struct {
	JWTSecret          string
	JWTExpiration      time.Duration
	RefreshTokenExpiry time.Duration
	BCryptCost         int
}

// ExternalConfig contains third-party service configurations
type ExternalConfig struct {
	StripeSecretKey     string
	StripeWebhookSecret string
	SendGridAPIKey      string
	FrontendURL         string
	UploadPath          string
	MaxUploadSize       int64
}

// FeatureFlags allows toggling features without code changes
type FeatureFlags struct {
	EnableWebSocket     bool
	EnableNotifications bool
	EnablePayments      bool
	MaintenanceMode     bool
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	cfg := &Config{
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:         getEnvOrDefault("PORT", "8080"),
			ReadTimeout:  getDurationOrDefault("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationOrDefault("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationOrDefault("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			MySQL: MySQLConfig{
				DSN:             getEnvOrDefault("MYSQL_DSN", ""),
				MaxOpenConns:    getIntOrDefault("MYSQL_MAX_OPEN_CONNS", 25),
				MaxIdleConns:    getIntOrDefault("MYSQL_MAX_IDLE_CONNS", 5),
				ConnMaxLifetime: getDurationOrDefault("MYSQL_CONN_MAX_LIFETIME", 5*time.Minute),
			},
			MongoDB: MongoDBConfig{
				URI:      getEnvOrDefault("MONGO_URI", ""),
				Database: getEnvOrDefault("MONGO_DATABASE", "tournament_planner"),
			},
			Redis: RedisConfig{
				Addr:     getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
				Password: getEnvOrDefault("REDIS_PASSWORD", ""),
				DB:       getIntOrDefault("REDIS_DB", 0),
			},
		},
		Auth: AuthConfig{
			JWTSecret:          getEnvOrDefault("JWT_SECRET", ""),
			JWTExpiration:      getDurationOrDefault("JWT_EXPIRATION", 15*time.Minute),
			RefreshTokenExpiry: getDurationOrDefault("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
			BCryptCost:         getIntOrDefault("BCRYPT_COST", 10),
		},
		External: ExternalConfig{
			StripeSecretKey:     getEnvOrDefault("STRIPE_SECRET_KEY", ""),
			StripeWebhookSecret: getEnvOrDefault("STRIPE_WEBHOOK_SECRET", ""),
			SendGridAPIKey:      getEnvOrDefault("SENDGRID_API_KEY", ""),
			FrontendURL:         getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),
			UploadPath:          getEnvOrDefault("UPLOAD_PATH", "./uploads"),
			MaxUploadSize:       getInt64OrDefault("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB
		},
		Features: FeatureFlags{
			EnableWebSocket:     getBoolOrDefault("ENABLE_WEBSOCKET", true),
			EnableNotifications: getBoolOrDefault("ENABLE_NOTIFICATIONS", true),
			EnablePayments:      getBoolOrDefault("ENABLE_PAYMENTS", true),
			MaintenanceMode:     getBoolOrDefault("MAINTENANCE_MODE", false),
		},
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required configuration is present
func (c *Config) Validate() error {
	if c.Database.MySQL.DSN == "" {
		return fmt.Errorf("MYSQL_DSN is required")
	}
	if c.Database.MongoDB.URI == "" {
		return fmt.Errorf("MONGO_URI is required")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.Environment == "production" {
		if c.External.StripeSecretKey == "" {
			return fmt.Errorf("STRIPE_SECRET_KEY is required in production")
		}
		if c.External.SendGridAPIKey == "" {
			return fmt.Errorf("SENDGRID_API_KEY is required in production")
		}
	}
	return nil
}

// Helper functions to read environment variables with defaults
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getInt64OrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
