package sportradar

import (
	"context"
	"sync"
	"time"
)

// RateLimiter provides rate limiting for API requests
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	interval   time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		maxTokens:  maxTokens,
		tokens:     maxTokens,
		interval:   interval,
		lastRefill: time.Now(),
	}
}

// Wait waits until a token is available
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		if r.tryConsume() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.interval / time.Duration(r.maxTokens)):
			// Wait for a short period before retrying
		}
	}
}

// tryConsume tries to consume a token
func (r *RateLimiter) tryConsume() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	// Refill tokens based on time elapsed
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed / r.interval)
	if tokensToAdd > 0 {
		r.tokens = min(r.maxTokens, r.tokens+tokensToAdd)
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}
