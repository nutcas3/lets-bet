package middleware

import (
	"net/http"
	"strconv"

	"github.com/betting-platform/internal/infrastructure/logging"
	"github.com/betting-platform/internal/infrastructure/ratelimit"
	"github.com/google/uuid"
)

// RateLimitMiddleware creates HTTP middleware using the existing ratelimit.RedisLimiter
func RateLimitMiddleware(redisLimiter *ratelimit.RedisLimiter, config *ratelimit.Config, proxyValidator *ProxyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := logging.FromContext(ctx)

			// Get client IP
			clientIP := clientIP(r, proxyValidator)

			// Check IP rate limit
			result, err := redisLimiter.CheckIPLimit(ctx, clientIP)
			if err != nil {
				// Log error but allow request to proceed (fail open)
				logger.Error("rate limit check failed", "error", err, "ip", clientIP)
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.IPRequestsPerWindow))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				writeJSON(w, http.StatusTooManyRequests, map[string]string{
					"error":       "rate limit exceeded",
					"retry_after": result.RetryAfter.String(),
					"limit_type":  result.LimitType,
				})
				logger.Warn("rate limit exceeded",
					"ip", clientIP,
					"limit_type", result.LimitType,
					"retry_after", result.RetryAfter,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// UserRateLimitMiddleware creates HTTP middleware for per-user rate limiting
func UserRateLimitMiddleware(redisLimiter *ratelimit.RedisLimiter, config *ratelimit.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := logging.FromContext(ctx)

			// Get user ID from context (should be set by JWT middleware)
			userID, ok := ctx.Value("user_id").(string)
			if !ok {
				// No user ID, proceed without rate limiting
				next.ServeHTTP(w, r)
				return
			}

			// Convert string to UUID
			userUUID, err := uuid.Parse(userID)
			if err != nil {
				logger.Error("invalid user ID in context", "user_id", userID, "error", err)
				next.ServeHTTP(w, r)
				return
			}

			// Check user rate limit
			result, err := redisLimiter.CheckUserLimit(ctx, userUUID)
			if err != nil {
				// Log error but allow request to proceed (fail open)
				logger.Error("user rate limit check failed", "error", err, "user_id", userID)
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.UserRequestsPerWindow))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

			if !result.Allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
				writeJSON(w, http.StatusTooManyRequests, map[string]string{
					"error":       "user rate limit exceeded",
					"retry_after": result.RetryAfter.String(),
					"limit_type":  result.LimitType,
				})
				logger.Warn("user rate limit exceeded",
					"user_id", userID,
					"limit_type", result.LimitType,
					"retry_after", result.RetryAfter,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CombinedRateLimitMiddleware creates middleware that checks both IP and user limits
func CombinedRateLimitMiddleware(redisLimiter *ratelimit.RedisLimiter, ipConfig, userConfig *ratelimit.Config, proxyValidator *ProxyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := logging.FromContext(ctx)

			// First check IP rate limit
			clientIP := clientIP(r, proxyValidator)
			ipResult, err := redisLimiter.CheckIPLimit(ctx, clientIP)
			if err != nil {
				logger.Error("IP rate limit check failed", "error", err, "ip", clientIP)
			} else if !ipResult.Allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(ipResult.RetryAfter.Seconds()), 10))
				writeJSON(w, http.StatusTooManyRequests, map[string]string{
					"error":       "IP rate limit exceeded",
					"retry_after": ipResult.RetryAfter.String(),
					"limit_type":  ipResult.LimitType,
				})
				logger.Warn("IP rate limit exceeded",
					"ip", clientIP,
					"limit_type", ipResult.LimitType,
					"retry_after", ipResult.RetryAfter,
				)
				return
			}

			// Set IP rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(ipConfig.IPRequestsPerWindow))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(ipResult.Remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(ipResult.ResetTime.Unix(), 10))

			// Then check user rate limit if user is authenticated
			userID, ok := ctx.Value("user_id").(string)
			if ok {
				userUUID, err := uuid.Parse(userID)
				if err == nil {
					userResult, err := redisLimiter.CheckUserLimit(ctx, userUUID)
					if err != nil {
						logger.Error("user rate limit check failed", "error", err, "user_id", userID)
					} else if !userResult.Allowed {
						w.Header().Set("Retry-After", strconv.FormatInt(int64(userResult.RetryAfter.Seconds()), 10))
						writeJSON(w, http.StatusTooManyRequests, map[string]string{
							"error":       "user rate limit exceeded",
							"retry_after": userResult.RetryAfter.String(),
							"limit_type":  userResult.LimitType,
						})
						logger.Warn("user rate limit exceeded",
							"user_id", userID,
							"limit_type", userResult.LimitType,
							"retry_after", userResult.RetryAfter,
						)
						return
					}

					// Override with user rate limit headers (more restrictive)
					w.Header().Set("X-RateLimit-Limit", strconv.Itoa(userConfig.UserRequestsPerWindow))
					w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(userResult.Remaining, 10))
					w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(userResult.ResetTime.Unix(), 10))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
