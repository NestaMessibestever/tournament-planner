// internal/database/connections.go
// This file manages all database connections and provides a unified interface
// for the application to access different data stores.

package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connections holds all database connections used by the application
type Connections struct {
	MySQL   *sql.DB
	MongoDB *mongo.Database
	Redis   *redis.Client
	logger  *log.Logger
}

// Config holds configuration for all databases
type Config struct {
	MySQL   MySQLConfig
	MongoDB MongoConfig
	Redis   RedisConfig
}

// MySQLConfig contains MySQL connection parameters
type MySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// MongoConfig contains MongoDB connection parameters
type MongoConfig struct {
	URI      string
	Database string
}

// RedisConfig contains Redis connection parameters
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Initialize creates and configures all database connections
func Initialize(ctx context.Context, cfg Config, logger *log.Logger) (*Connections, error) {
	conn := &Connections{logger: logger}

	// Initialize MySQL with retry logic
	if err := conn.initMySQL(ctx, cfg.MySQL); err != nil {
		return nil, fmt.Errorf("failed to initialize MySQL: %w", err)
	}

	// Initialize MongoDB
	if err := conn.initMongoDB(ctx, cfg.MongoDB); err != nil {
		conn.Close() // Clean up MySQL connection
		return nil, fmt.Errorf("failed to initialize MongoDB: %w", err)
	}

	// Initialize Redis
	if err := conn.initRedis(ctx, cfg.Redis); err != nil {
		conn.Close() // Clean up previous connections
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	logger.Println("All database connections established successfully")
	return conn, nil
}

// initMySQL establishes MySQL connection with retry logic
func (c *Connections) initMySQL(ctx context.Context, cfg MySQLConfig) error {
	var err error
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		c.MySQL, err = sql.Open("mysql", cfg.DSN)
		if err != nil {
			c.logger.Printf("Failed to open MySQL connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		// Configure connection pool
		c.MySQL.SetMaxOpenConns(cfg.MaxOpenConns)
		c.MySQL.SetMaxIdleConns(cfg.MaxIdleConns)
		c.MySQL.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		// Test the connection
		if err = c.MySQL.PingContext(ctx); err != nil {
			c.logger.Printf("Failed to ping MySQL (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		c.logger.Println("MySQL connection established")
		return nil
	}

	return fmt.Errorf("failed to connect to MySQL after %d attempts: %w", maxRetries, err)
}

// initMongoDB establishes MongoDB connection
func (c *Connections) initMongoDB(ctx context.Context, cfg MongoConfig) error {
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	c.MongoDB = client.Database(cfg.Database)
	c.logger.Println("MongoDB connection established")
	return nil
}

// initRedis establishes Redis connection
func (c *Connections) initRedis(ctx context.Context, cfg RedisConfig) error {
	c.Redis = redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	if err := c.Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	c.logger.Println("Redis connection established")
	return nil
}

// Close gracefully closes all database connections
func (c *Connections) Close() {
	if c.MySQL != nil {
		if err := c.MySQL.Close(); err != nil {
			c.logger.Printf("Error closing MySQL connection: %v", err)
		}
	}

	if c.MongoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.MongoDB.Client().Disconnect(ctx); err != nil {
			c.logger.Printf("Error closing MongoDB connection: %v", err)
		}
	}

	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			c.logger.Printf("Error closing Redis connection: %v", err)
		}
	}

	c.logger.Println("All database connections closed")
}

// HealthCheck verifies all database connections are healthy
func (c *Connections) HealthCheck(ctx context.Context) error {
	// Check MySQL
	if err := c.MySQL.PingContext(ctx); err != nil {
		return fmt.Errorf("MySQL health check failed: %w", err)
	}

	// Check MongoDB
	if err := c.MongoDB.Client().Ping(ctx, nil); err != nil {
		return fmt.Errorf("MongoDB health check failed: %w", err)
	}

	// Check Redis
	if err := c.Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}
