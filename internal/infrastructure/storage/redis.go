package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements cache operations using Redis
type RedisCache struct {
	client *redis.Client
}

// Config holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(cfg RedisConfig) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: rdb}, nil
}

// Set stores a value in Redis with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a value from Redis
func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key not found")
		}
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return value, nil
}

// Delete removes a key from Redis
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result > 0, err
}

// SetString stores a string value in Redis
func (r *RedisCache) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// GetString retrieves a string value from Redis
func (r *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Increment increments a numeric value in Redis
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// SetWithExpiration sets a key with expiration in seconds
func (r *RedisCache) SetWithExpiration(ctx context.Context, key string, value interface{}, expiration int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, time.Duration(expiration)*time.Second).Err()
}

// GetDefaultRedisConfig returns default Redis configuration
func GetDefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       0,
	}
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}
