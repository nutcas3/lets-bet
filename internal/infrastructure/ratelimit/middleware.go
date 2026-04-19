package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
)

// HTTPMiddleware creates HTTP middleware for rate limiting
func (rl *RedisLimiter) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get client IP
			clientIP := getClientIP(r)

			// Get user ID from context (if authenticated)
			var userID uuid.UUID
			if userIDStr, ok := ctx.Value("user_id").(string); ok {
				if parsedID, err := uuid.Parse(userIDStr); err == nil {
					userID = parsedID
				}
			}

			// Check rate limits
			var result *LimitResult
			var err error

			if userID != (uuid.UUID{}) {
				// User is authenticated, check user limit
				result, err = rl.CheckUserLimit(ctx, userID)
			} else {
				// User is not authenticated, only check IP limit
				result, err = rl.CheckIPLimit(ctx, clientIP)
			}

			if err != nil {
				// Log error but allow request (fail open)
				fmt.Printf("Rate limit check failed: %v\n", err)
				next.ServeHTTP(w, r)
				return
			}

			// Add rate limit headers
			rl.addRateLimitHeaders(w, result)

			// Check if request is allowed
			if !result.Allowed {
				// Return 429 Too Many Requests
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", time.Until(result.ResetTime).Seconds()))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded","retry_after":` + fmt.Sprintf("%.0f", time.Until(result.ResetTime).Seconds()) + `}`))
				return
			}

			// Add rate limit info to context
			ctx = context.WithValue(ctx, "rate_limit_result", result)

			// Continue with request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// addRateLimitHeaders adds rate limit headers to the response
func (rl *RedisLimiter) addRateLimitHeaders(w http.ResponseWriter, result *LimitResult) {
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.getLimitForType(result.LimitType)))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%.0f", float64(result.ResetTime.Unix())))
	w.Header().Set("X-RateLimit-Type", result.LimitType)
}

// getLimitForType returns the limit for a given limit type
func (rl *RedisLimiter) getLimitForType(limitType string) int {
	switch limitType {
	case "user":
		return rl.config.UserRequestsPerWindow
	case "ip":
		return rl.config.IPRequestsPerWindow
	case "global":
		return rl.config.GlobalRequestsPerWindow
	default:
		return rl.config.DefaultRequestsPerWindow
	}
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (when behind proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		// Simple split for first IP
		for i, c := range xff {
			if c == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// RateLimitMiddlewareConfig holds configuration for rate limiting middleware
type RateLimitMiddlewareConfig struct {
	// Skip paths that should not be rate limited
	SkipPaths []string

	// Custom user ID extractor function
	UserIDExtractor func(*http.Request) (uuid.UUID, bool)

	// Custom IP extractor function
	IPExtractor func(*http.Request) string

	// Custom error handler
	ErrorHandler func(http.ResponseWriter, *http.Request, error)
}

// HTTPMiddlewareWithConfig creates HTTP middleware with custom configuration
func (rl *RedisLimiter) HTTPMiddlewareWithConfig(config RateLimitMiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path should be skipped
			if slices.Contains(config.SkipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			// Get client IP
			clientIP := r.RemoteAddr
			if config.IPExtractor != nil {
				clientIP = config.IPExtractor(r)
			} else {
				clientIP = getClientIP(r)
			}

			// Get user ID
			var userID uuid.UUID
			var hasUser bool

			if config.UserIDExtractor != nil {
				userID, hasUser = config.UserIDExtractor(r)
			} else {
				// Default user ID extraction from context
				if userIDStr, ok := ctx.Value("user_id").(string); ok {
					if parsedID, err := uuid.Parse(userIDStr); err == nil {
						userID = parsedID
						hasUser = true
					}
				}
			}

			// Check rate limits
			var result *LimitResult
			var err error

			if hasUser {
				// User is authenticated, check user limit
				result, err = rl.CheckUserLimit(ctx, userID)
			} else {
				// User is not authenticated, only check IP limit
				result, err = rl.CheckIPLimit(ctx, clientIP)
			}

			if err != nil {
				// Handle error
				if config.ErrorHandler != nil {
					config.ErrorHandler(w, r, err)
					return
				}

				// Default error handling: log and allow
				fmt.Printf("Rate limit check failed: %v\n", err)
				next.ServeHTTP(w, r)
				return
			}

			// Add rate limit headers
			rl.addRateLimitHeaders(w, result)

			// Check if request is allowed
			if !result.Allowed {
				// Handle rate limit exceeded
				if config.ErrorHandler != nil {
					config.ErrorHandler(w, r, fmt.Errorf("rate limit exceeded"))
					return
				}

				// Default rate limit response
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", time.Until(result.ResetTime).Seconds()))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded","retry_after":` + fmt.Sprintf("%.0f", time.Until(result.ResetTime).Seconds()) + `}`))
				return
			}

			// Add rate limit info to context
			ctx = context.WithValue(ctx, "rate_limit_result", result)

			// Continue with request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
