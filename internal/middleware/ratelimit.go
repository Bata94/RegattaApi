package middleware

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/handler"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

var (
	globalLimiter *RateLimiter
	once          sync.Once
)

func InitRateLimiter() {
	once.Do(func() {
		globalLimiter = &RateLimiter{
			limiters: make(map[string]*rate.Limiter),
			rate:     rate.Limit(config.C.Rate.RPS),
			burst:    config.C.Rate.Burst,
			cleanup:  5 * time.Minute,
		}

		go globalLimiter.cleanupLoop()
	})
}

func RateLimit() Middleware {
	if globalLimiter == nil {
		InitRateLimiter()
	}

	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			key := getRateLimitKey(c)
			limiter := globalLimiter.getLimiter(key)

			if !limiter.Allow() {
				burst := globalLimiter.burst

				remaining := 0
				if limiter.Tokens() > 0 {
					remaining = int(limiter.Tokens())
				}

				c.Writer.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", burst))
				c.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
				c.Writer.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

				retryAfter := 1
				c.Writer.Header().Set("Retry-After", strconv.Itoa(retryAfter))

				return handler.InternalError("Too many requests")
			}

			c.Writer.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", globalLimiter.burst))
			remaining := int(limiter.Tokens())
			if remaining < 0 {
				remaining = 0
			}
			c.Writer.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Writer.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

			return next(c)
		}
	}
}

func getRateLimitKey(c *handler.Context) string {
	if userID := c.GetLocals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok && uid != "" {
			return "user:" + uid
		}
	}

	ip := c.IP()
	if ip == "" {
		ip = "unknown"
	}
	return "ip:" + ip
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limiter, exists = rl.limiters[key]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters[key] = limiter

	return limiter
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for key, limiter := range rl.limiters {
			if limiter.Tokens() >= float64(rl.burst) {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// TODO: Implement distributed rate limiting with Redis for multi-instance deployments.
// Current in-memory implementation works for single-server deployments.
// For distributed setups, use Redis with sliding window or token bucket algorithms.
// Reference: https://github.com/redis/go-redis rate limiting or Redis Cell (SETNX + EXPIRE).