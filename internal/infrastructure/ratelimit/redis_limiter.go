// Package ratelimit provides Redis-backed rate limiting for the betting platform.
//
// It implements:
// - Per-user rate limiting
// - Per-IP rate limiting
// - Sliding window algorithm
// - Distributed rate limiting across multiple instances
// - Configurable limits and windows
package ratelimit

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisLimiter implements rate limiting using Redis
type RedisLimiter struct {
	client *redis.Client
	config *Config
}

// NewRedisLimiter creates a new Redis-backed rate limiter
func NewRedisLimiter(ctx context.Context, config *Config) (*RedisLimiter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisLimiter{
		client: rdb,
		config: config,
	}, nil
}

// CheckUserLimit checks if a user has exceeded their rate limit
func (rl *RedisLimiter) CheckUserLimit(ctx context.Context, userID uuid.UUID) (*LimitResult, error) {
	key := rl.config.UserPrefix + userID.String()
	return rl.checkLimit(ctx, key, rl.config.UserRequestsPerWindow, rl.config.UserWindow, "user")
}

// CheckIPLimit checks if an IP has exceeded their rate limit
func (rl *RedisLimiter) CheckIPLimit(ctx context.Context, ip string) (*LimitResult, error) {
	// Normalize IP address
	normalizedIP := net.ParseIP(ip)
	if normalizedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Use IPv4 or IPv6 representation consistently
	var key string
	if normalizedIP.To4() != nil {
		key = rl.config.IPPrefix + normalizedIP.String()
	} else {
		key = rl.config.IPPrefix + normalizedIP.String()
	}

	return rl.checkLimit(ctx, key, rl.config.IPRequestsPerWindow, rl.config.IPWindow, "ip")
}

// CheckGlobalLimit checks if the global limit has been exceeded
func (rl *RedisLimiter) CheckGlobalLimit(ctx context.Context) (*LimitResult, error) {
	key := rl.config.GlobalPrefix + "all"
	return rl.checkLimit(ctx, key, rl.config.GlobalRequestsPerWindow, rl.config.GlobalWindow, "global")
}

// checkLimit performs the actual rate limit check using sliding window
func (rl *RedisLimiter) checkLimit(ctx context.Context, key string, limit int, window time.Duration, limitType string) (*LimitResult, error) {
	now := time.Now()
	windowStart := now.Truncate(window)
	windowEnd := windowStart.Add(window)

	// Use Lua script for atomic sliding window implementation
	luaScript := `
local key = KEYS[1]
local window_start = ARGV[1]
local window_end = ARGV[2]
local limit = tonumber(ARGV[3])
local now = ARGV[4]

-- Remove expired entries
redis.call('ZREMRANGEBYSCORE', key, 0, window_start - 1)

-- Count current requests
local current = redis.call('ZCARD', key)

-- Check if limit exceeded
if current >= limit then
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    return {0, limit, current, oldest[2]}
end

-- Add current request
redis.call('ZADD', key, now, now)
redis.call('EXPIRE', key, math.ceil(tonumber(window_end - window_start) / 1000000000))

return {1, limit - current, limit - current, now}
`

	result, err := rl.client.Eval(ctx, luaScript, []string{key},
		windowStart.UnixNano(),
		windowEnd.UnixNano(),
		limit,
		now.UnixNano()).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	// Parse Lua script result
	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	remaining := resultSlice[1].(int64)
	var resetTime time.Time
	if len(resultSlice) > 3 {
		resetTimestamp := resultSlice[3].(int64)
		resetTime = time.Unix(0, resetTimestamp)
	} else {
		resetTime = windowEnd
	}

	return &LimitResult{
		Allowed:    allowed,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: resetTime.Sub(now),
		LimitType:  limitType,
		Key:        key,
	}, nil
}

// ResetUserLimit resets the rate limit for a specific user
func (rl *RedisLimiter) ResetUserLimit(ctx context.Context, userID uuid.UUID) error {
	key := rl.config.UserPrefix + userID.String()
	return rl.client.Del(ctx, key).Err()
}

// ResetIPLimit resets the rate limit for a specific IP
func (rl *RedisLimiter) ResetIPLimit(ctx context.Context, ip string) error {
	normalizedIP := net.ParseIP(ip)
	if normalizedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	var key string
	if normalizedIP.To4() != nil {
		key = rl.config.IPPrefix + normalizedIP.String()
	} else {
		key = rl.config.IPPrefix + normalizedIP.String()
	}

	return rl.client.Del(ctx, key).Err()
}

// ResetGlobalLimit resets the global rate limit
func (rl *RedisLimiter) ResetGlobalLimit(ctx context.Context) error {
	key := rl.config.GlobalPrefix + "all"
	return rl.client.Del(ctx, key).Err()
}

// GetUsageStats returns usage statistics for a specific key
func (rl *RedisLimiter) GetUsageStats(ctx context.Context, key string, window time.Duration) (*UsageStats, error) {
	now := time.Now()
	windowStart := now.Truncate(window)
	windowEnd := windowStart.Add(window)

	// Count requests in current window
	current, err := rl.client.ZCount(ctx, key, fmt.Sprintf("%d", windowStart.UnixNano()), fmt.Sprintf("%d", now.UnixNano())).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	// Get oldest request timestamp
	oldest, err := rl.client.ZRange(ctx, key, 0, 0).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get oldest request: %w", err)
	}

	var oldestTime time.Time
	if len(oldest) > 0 {
		timestamp, err := time.Parse(time.RFC3339Nano, oldest[0])
		if err == nil {
			oldestTime = timestamp
		}
	}

	return &UsageStats{
		Key:           key,
		CurrentCount:  int64(current),
		WindowStart:   windowStart,
		WindowEnd:     windowEnd,
		OldestRequest: oldestTime,
		NewestRequest: now,
	}, nil
}

// CleanupExpiredKeys removes expired rate limit keys
func (rl *RedisLimiter) CleanupExpiredKeys(ctx context.Context) error {
	patterns := []string{
		rl.config.UserPrefix + "*",
		rl.config.IPPrefix + "*",
		rl.config.GlobalPrefix + "*",
	}

	for _, pattern := range patterns {
		keys, err := rl.client.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}

		for _, key := range keys {
			ttl := rl.client.TTL(ctx, key).Val()
			if ttl == -1 {
				// Key has no expiration, set a reasonable default
				rl.client.Expire(ctx, key, time.Hour)
			}
		}
	}

	return nil
}

// Close closes the Redis connection
func (rl *RedisLimiter) Close() error {
	return rl.client.Close()
}

// UsageStats represents usage statistics for a rate limit key
type UsageStats struct {
	Key           string    `json:"key"`
	CurrentCount  int64     `json:"current_count"`
	WindowStart   time.Time `json:"window_start"`
	WindowEnd     time.Time `json:"window_end"`
	OldestRequest time.Time `json:"oldest_request"`
	NewestRequest time.Time `json:"newest_request"`
}

// HealthCheck checks if the rate limiter is healthy
func (rl *RedisLimiter) HealthCheck(ctx context.Context) error {
	return rl.client.Ping(ctx).Err()
}

// GetRedisClient returns the underlying Redis client (for advanced usage)
func (rl *RedisLimiter) GetRedisClient() *redis.Client {
	return rl.client
}
