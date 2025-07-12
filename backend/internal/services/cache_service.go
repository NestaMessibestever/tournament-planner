// internal/services/cache_service.go
// Cache service for Redis/Memurai operations

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService handles all caching operations
type CacheService struct {
	client *redis.Client
	logger *log.Logger
}

// NewCacheService creates a new cache service
func NewCacheService(client *redis.Client, logger *log.Logger) *CacheService {
	return &CacheService{
		client: client,
		logger: logger,
	}
}

// Set stores a value in cache with expiration
func (s *CacheService) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := s.client.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (s *CacheService) Get(key string, dest interface{}) error {
	ctx := context.Background()

	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("key not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (s *CacheService) Delete(key string) error {
	ctx := context.Background()

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (s *CacheService) Exists(key string) (bool, error) {
	ctx := context.Background()

	count, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return count > 0, nil
}

// Increment increments a counter in cache
func (s *CacheService) Increment(key string, expiration time.Duration) (int, error) {
	ctx := context.Background()

	// Use pipeline for atomic operation
	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}

	return int(incr.Val()), nil
}

// SetNX sets a key only if it doesn't exist (for distributed locks)
func (s *CacheService) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	ctx := context.Background()

	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	ok, err := s.client.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setnx: %w", err)
	}

	return ok, nil
}

// GetOrSet gets a value from cache or sets it if not exists
func (s *CacheService) GetOrSet(key string, dest interface{}, fn func() (interface{}, error), expiration time.Duration) error {
	// Try to get from cache first
	if err := s.Get(key, dest); err == nil {
		return nil
	}

	// Not in cache, call function to get value
	value, err := fn()
	if err != nil {
		return err
	}

	// Set in cache
	if err := s.Set(key, value, expiration); err != nil {
		s.logger.Printf("Failed to cache value for key %s: %v", key, err)
	}

	// Marshal/unmarshal to ensure dest has the value
	data, _ := json.Marshal(value)
	return json.Unmarshal(data, dest)
}

// InvalidatePattern deletes all keys matching a pattern
func (s *CacheService) InvalidatePattern(pattern string) error {
	ctx := context.Background()

	// Get all keys matching pattern
	keys, err := s.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all matching keys
	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// Ping checks if cache is available
func (s *CacheService) Ping() error {
	ctx := context.Background()
	return s.client.Ping(ctx).Err()
}
