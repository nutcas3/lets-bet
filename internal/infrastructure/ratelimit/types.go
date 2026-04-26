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
	"time"

	"github.com/go-redis/redis/v8"
)

// Config holds rate limiting configuration
type Config struct {
	// Default limits
	DefaultRequestsPerWindow int
	DefaultWindow            time.Duration

	// User-specific limits
	UserRequestsPerWindow int
	UserWindow            time.Duration

	// IP-specific limits
	IPRequestsPerWindow int
	IPWindow            time.Duration

	// Global limits (across all users/IPs)
	GlobalRequestsPerWindow int
	GlobalWindow            time.Duration

	// Redis configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Key prefixes
	UserPrefix   string
	IPPrefix     string
	GlobalPrefix string

	// Proxy validation for preventing X-Forwarded-For spoofing
	TrustedProxyCIDRs []string
}

// DefaultConfig returns default rate limiting configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultRequestsPerWindow: 100,
		DefaultWindow:            time.Minute,
		UserRequestsPerWindow:    200,
		UserWindow:               time.Minute,
		IPRequestsPerWindow:      50,
		IPWindow:                 time.Minute,
		GlobalRequestsPerWindow:  1000,
		GlobalWindow:             time.Minute,
		RedisAddr:                "localhost:6379",
		RedisPassword:            "",
		RedisDB:                  0,
		UserPrefix:               "rate_limit:user:",
		IPPrefix:                 "rate_limit:ip:",
		GlobalPrefix:             "rate_limit:global:",
	}
}

// RateLimiter provides Redis-backed rate limiting
type RateLimiter struct {
	redisClient *redis.Client
	config      *Config
}

// RequestType represents the type of request being rate limited
type RequestType string

const (
	RequestTypeAuth     RequestType = "auth"
	RequestTypeBet      RequestType = "bet"
	RequestTypeDeposit  RequestType = "deposit"
	RequestTypeWithdraw RequestType = "withdraw"
	RequestTypeOdds     RequestType = "odds"
	RequestTypeProfile  RequestType = "profile"
	RequestTypeGeneric  RequestType = "generic"
)

// LimitResult represents the result of a rate limit check
type LimitResult struct {
	Allowed    bool          `json:"allowed"`
	Remaining  int64         `json:"remaining"`
	ResetTime  time.Time     `json:"reset_time"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`
	LimitType  string        `json:"limit_type"`
	Key        string        `json:"key"`
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	Type       string        `json:"type"`
	Message    string        `json:"message"`
	RetryAfter time.Duration `json:"retry_after"`
	LimitType  string        `json:"limit_type"`
	Key        string        `json:"key"`
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// LimitConfig represents rate limit configuration for a specific type
type LimitConfig struct {
	RequestsPerWindow int           `json:"requests_per_window"`
	Window            time.Duration `json:"window"`
	KeyPrefix         string        `json:"key_prefix"`
}

// UserLimit represents rate limit data for a user
type UserLimit struct {
	UserID       string    `json:"user_id"`
	RequestCount int64     `json:"request_count"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Limit        int64     `json:"limit"`
}

// IPLimit represents rate limit data for an IP
type IPLimit struct {
	IPAddress    string    `json:"ip_address"`
	RequestCount int64     `json:"request_count"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Limit        int64     `json:"limit"`
}

// GlobalLimit represents global rate limit data
type GlobalLimit struct {
	RequestCount int64     `json:"request_count"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	Limit        int64     `json:"limit"`
}

// RateLimitMetrics represents rate limiting metrics
type RateLimitMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	AllowedRequests     int64         `json:"allowed_requests"`
	BlockedRequests     int64         `json:"blocked_requests"`
	UserLimits          int64         `json:"user_limits"`
	IPLimits            int64         `json:"ip_limits"`
	GlobalLimits        int64         `json:"global_limits"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastResetTime       time.Time     `json:"last_reset_time"`
}

// RateLimitInfo represents comprehensive rate limit information
type RateLimitInfo struct {
	UserLimits   []*UserLimit      `json:"user_limits"`
	IPLimits     []*IPLimit        `json:"ip_limits"`
	GlobalLimits []*GlobalLimit    `json:"global_limits"`
	Metrics      *RateLimitMetrics `json:"metrics"`
}

// RateLimitStore interface for rate limiting operations
type RateLimitStore interface {
	// Basic rate limiting
	CheckLimit(ctx context.Context, key string, limit int, window time.Duration) (*LimitResult, error)
	IncrementRequest(ctx context.Context, key string, window time.Duration) (int64, error)
	ResetLimit(ctx context.Context, key string) error

	// User-specific operations
	CheckUserLimit(ctx context.Context, userID string, requestType RequestType) (*LimitResult, error)
	GetUserLimitInfo(ctx context.Context, userID string) (*UserLimit, error)

	// IP-specific operations
	CheckIPLimit(ctx context.Context, ip string, requestType RequestType) (*LimitResult, error)
	GetIPLimitInfo(ctx context.Context, ip string) (*IPLimit, error)

	// Global operations
	CheckGlobalLimit(ctx context.Context, requestType RequestType) (*LimitResult, error)
	GetGlobalLimitInfo(ctx context.Context) (*GlobalLimit, error)

	// Metrics and info
	GetMetrics(ctx context.Context) (*RateLimitMetrics, error)
	GetRateLimitInfo(ctx context.Context) (*RateLimitInfo, error)

	// Cleanup operations
	CleanupExpiredKeys(ctx context.Context) error
	ResetAllLimits(ctx context.Context) error
}

// LimitKey represents a rate limit key with metadata
type LimitKey struct {
	Key          string        `json:"key"`
	Type         RequestType   `json:"type"`
	Limit        int           `json:"limit"`
	Window       time.Duration `json:"window"`
	ExpiresAt    time.Time     `json:"expires_at"`
	RequestCount int64         `json:"request_count"`
}

// RateLimitConfig represents configuration for different request types
type RateLimitConfig struct {
	Auth     *LimitConfig `json:"auth"`
	Bet      *LimitConfig `json:"bet"`
	Deposit  *LimitConfig `json:"deposit"`
	Withdraw *LimitConfig `json:"withdraw"`
	Odds     *LimitConfig `json:"odds"`
	Profile  *LimitConfig `json:"profile"`
	Generic  *LimitConfig `json:"generic"`
}

// GetDefaultRateLimitConfig returns default rate limit configurations
func GetDefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Auth: &LimitConfig{
			RequestsPerWindow: 10,
			Window:            time.Minute,
			KeyPrefix:         "auth:",
		},
		Bet: &LimitConfig{
			RequestsPerWindow: 50,
			Window:            time.Minute,
			KeyPrefix:         "bet:",
		},
		Deposit: &LimitConfig{
			RequestsPerWindow: 20,
			Window:            time.Minute,
			KeyPrefix:         "deposit:",
		},
		Withdraw: &LimitConfig{
			RequestsPerWindow: 15,
			Window:            time.Minute,
			KeyPrefix:         "withdraw:",
		},
		Odds: &LimitConfig{
			RequestsPerWindow: 100,
			Window:            time.Minute,
			KeyPrefix:         "odds:",
		},
		Profile: &LimitConfig{
			RequestsPerWindow: 30,
			Window:            time.Minute,
			KeyPrefix:         "profile:",
		},
		Generic: &LimitConfig{
			RequestsPerWindow: 100,
			Window:            time.Minute,
			KeyPrefix:         "generic:",
		},
	}
}
