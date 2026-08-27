package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bata94/RegattaApi/internal/config"
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

func RateLimit() func(http.Handler) http.Handler {
	if globalLimiter == nil {
		InitRateLimiter()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := getRateLimitKey(r)
			limiter := globalLimiter.getLimiter(key)

			if !limiter.Allow() {
				burst := globalLimiter.burst

				remaining := 0
				if limiter.Tokens() > 0 {
					remaining = int(limiter.Tokens())
				}

				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", burst))
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
				w.Header().Set("Retry-After", "1")

				w.WriteHeader(http.StatusTooManyRequests)
				if _, err := w.Write([]byte("Too many requests")); err != nil {
					slog.Error("failed to write rate limit response", "error", err)
				}
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", globalLimiter.burst))
			remaining := int(limiter.Tokens())
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

			next.ServeHTTP(w, r)
		})
	}
}

func getRateLimitKey(r *http.Request) string {
	if userID := r.Context().Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok && uid != "" {
			return "user:" + uid
		}
	}

	ip := IP(r)
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
