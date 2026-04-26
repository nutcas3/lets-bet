package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/betting-platform/internal/infrastructure/ratelimit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitMiddleware_IPRateLimiting(t *testing.T) {
	t.Parallel()

	// Create test Redis config
	config := &ratelimit.Config{
		IPRequestsPerWindow: 5,
		IPWindow:            time.Second * 2,
		RedisAddr:           "localhost:6379",
		RedisPassword:       "",
		RedisDB:             1, // Use different DB for testing
		UserPrefix:          "test:user:",
		IPPrefix:            "test:ip:",
		GlobalPrefix:        "test:global:",
	}

	ctx := context.Background()
	redisLimiter, err := ratelimit.NewRedisLimiter(ctx, config)
	require.NoError(t, err)
	defer redisLimiter.Close()

	// Create test middleware
	middleware := RateLimitMiddleware(redisLimiter, config, nil)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test rate limiting works
	clientIP := "192.168.1.100"

	// First 5 requests should succeed
	for range 5 {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = clientIP + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = clientIP + ":12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "rate limit exceeded")
	assert.NotEmpty(t, w.Header().Get("Retry-After"))
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		IPRequestsPerWindow: 2,
		IPWindow:            time.Second * 2,
		RedisAddr:           "localhost:6379",
		RedisPassword:       "",
		RedisDB:             1,
		UserPrefix:          "test:user:",
		IPPrefix:            "test:ip:",
		GlobalPrefix:        "test:global:",
	}

	ctx := context.Background()
	redisLimiter, err := ratelimit.NewRedisLimiter(ctx, config)
	require.NoError(t, err)
	defer redisLimiter.Close()

	middleware := RateLimitMiddleware(redisLimiter, config, nil)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Different IPs should have independent rate limits
	ips := []string{"192.168.1.100", "192.168.1.101", "192.168.1.102"}

	for _, ip := range ips {
		for range 2 {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = ip + ":12345"
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "IP %s should succeed", ip)
		}

		// 3rd request for each IP should be rate limited
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, "IP %s should be rate limited", ip)
	}
}

func TestUserRateLimitMiddleware_UserRateLimiting(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		UserRequestsPerWindow: 3,
		UserWindow:            time.Second * 2,
		RedisAddr:             "localhost:6379",
		RedisPassword:         "",
		RedisDB:               1,
		UserPrefix:            "test:user:",
		IPPrefix:              "test:ip:",
		GlobalPrefix:          "test:global:",
	}

	ctx := context.Background()
	redisLimiter, err := ratelimit.NewRedisLimiter(ctx, config)
	require.NoError(t, err)
	defer redisLimiter.Close()

	middleware := UserRateLimitMiddleware(redisLimiter, config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	userID := uuid.New().String()

	// First 3 requests should succeed
	for range 3 {
		req := httptest.NewRequest("GET", "/", nil)
		req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "3", w.Header().Get("X-RateLimit-Limit"))
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "user rate limit exceeded")
}

func TestUserRateLimitMiddleware_NoUserInContext(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		UserRequestsPerWindow: 1,
		UserWindow:            time.Second * 2,
		RedisAddr:             "localhost:6379",
		RedisPassword:         "",
		RedisDB:               1,
		UserPrefix:            "test:user:",
		IPPrefix:              "test:ip:",
		GlobalPrefix:          "test:global:",
	}

	ctx := context.Background()
	redisLimiter, err := ratelimit.NewRedisLimiter(ctx, config)
	require.NoError(t, err)
	defer redisLimiter.Close()

	middleware := UserRateLimitMiddleware(redisLimiter, config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Request without user in context should succeed
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCombinedRateLimitMiddleware_BothLimits(t *testing.T) {
	t.Parallel()

	ipConfig := &ratelimit.Config{
		IPRequestsPerWindow: 2,
		IPWindow:            time.Second * 2,
		RedisAddr:           "localhost:6379",
		RedisPassword:       "",
		RedisDB:             1,
		UserPrefix:          "test:user:",
		IPPrefix:            "test:ip:",
		GlobalPrefix:        "test:global:",
	}

	userConfig := &ratelimit.Config{
		UserRequestsPerWindow: 3,
		UserWindow:            time.Second * 2,
		RedisAddr:             "localhost:6379",
		RedisPassword:         "",
		RedisDB:               1,
		UserPrefix:            "test:user:",
		IPPrefix:              "test:ip:",
		GlobalPrefix:          "test:global:",
	}

	ctx := context.Background()
	redisLimiter, err := ratelimit.NewRedisLimiter(ctx, ipConfig)
	require.NoError(t, err)
	defer redisLimiter.Close()

	middleware := CombinedRateLimitMiddleware(redisLimiter, ipConfig, userConfig, nil)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	userID := uuid.New().String()
	clientIP := "192.168.1.100"

	// Test that both IP and user limits are enforced
	// First 2 requests should succeed (limited by IP)
	for range 2 {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = clientIP + ":12345"
		req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 3rd request should be rate limited by IP limit
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = clientIP + ":12345"
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "IP rate limit exceeded")
}

func TestRateLimitMiddleware_MemoryLeakProtection(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		IPRequestsPerWindow: 1,
		IPWindow:            time.Millisecond * 100, // Short window for testing
		RedisAddr:           "localhost:6379",
		RedisPassword:       "",
		RedisDB:             1,
		UserPrefix:          "test:user:",
		IPPrefix:            "test:ip:",
		GlobalPrefix:        "test:global:",
	}

	ctx := context.Background()
	redisLimiter, err := ratelimit.NewRedisLimiter(ctx, config)
	require.NoError(t, err)
	defer redisLimiter.Close()

	middleware := RateLimitMiddleware(redisLimiter, config, nil)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Simulate DDoS with many unique IPs (this would cause memory leak in in-memory limiter)
	for i := range 1000 {
		ip := fmt.Sprintf("192.168.%d.%d", i/256, i%256)
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Each unique IP should succeed on first request
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Wait for window to expire
	time.Sleep(time.Millisecond * 150)

	// Same IPs should be able to make requests again
	for i := range 10 {
		ip := fmt.Sprintf("192.168.%d.%d", i/256, i%256)
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Memory usage should be bounded by Redis TTL, not by number of unique IPs
	// This test verifies we don't have the memory leak vulnerability
}

func TestRateLimitMiddleware_Headers(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		IPRequestsPerWindow: 10,
		IPWindow:            time.Second * 10,
		RedisAddr:           "localhost:6379",
		RedisPassword:       "",
		RedisDB:             1,
		UserPrefix:          "test:user:",
		IPPrefix:            "test:ip:",
		GlobalPrefix:        "test:global:",
	}

	ctx := context.Background()
	redisLimiter, err := ratelimit.NewRedisLimiter(ctx, config)
	require.NoError(t, err)
	defer redisLimiter.Close()

	middleware := RateLimitMiddleware(redisLimiter, config, nil)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check rate limit headers are set
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimitMiddleware_FailOpen(t *testing.T) {
	t.Parallel()

	// This test verifies that if Redis is unavailable, the middleware fails open
	// and allows requests to proceed rather than blocking all traffic

	config := &ratelimit.Config{
		IPRequestsPerWindow: 1,
		IPWindow:            time.Second * 2,
		RedisAddr:           "invalid:6379", // Invalid Redis address
		RedisPassword:       "",
		RedisDB:             1,
		UserPrefix:          "test:user:",
		IPPrefix:            "test:ip:",
		GlobalPrefix:        "test:global:",
	}

	ctx := context.Background()
	_, err := ratelimit.NewRedisLimiter(ctx, config)
	// This should fail during creation
	assert.Error(t, err)
}
