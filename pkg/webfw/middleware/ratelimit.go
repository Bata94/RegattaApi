package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter *rate.Limiter
	burst   int
}

type RateLimiter struct {
	mu             sync.RWMutex
	limiters       map[string]*limiterEntry
	rate           rate.Limit
	burst          int
	userMultiplier int
	cleanup        time.Duration
}

var (
	globalLimiter *RateLimiter
	once          sync.Once
)

func InitRateLimiter() {
	once.Do(func() {
		multiplier := config.C.Rate.UserMultiplier
		if multiplier < 1 {
			multiplier = 1
		}
		globalLimiter = &RateLimiter{
			limiters:       make(map[string]*limiterEntry),
			rate:           rate.Limit(config.C.Rate.RPS),
			burst:          config.C.Rate.Burst,
			userMultiplier: multiplier,
			cleanup:        5 * time.Minute,
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
			if isStaticAssetPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			key, limit, burst := globalLimiter.limitsForKey(r)
			entry := globalLimiter.getLimiter(key, limit, burst)

			if !entry.limiter.Allow() {
				remaining := 0
				if entry.limiter.Tokens() > 0 {
					remaining = int(entry.limiter.Tokens())
				}

				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", entry.burst))
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
				w.Header().Set("Retry-After", "1")

				w.WriteHeader(http.StatusTooManyRequests)
				if _, err := w.Write([]byte("Too many requests")); err != nil {
					slog.Error("failed to write rate limit response", "error", err)
				}
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", entry.burst))
			remaining := int(entry.limiter.Tokens())
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

			next.ServeHTTP(w, r)
		})
	}
}

func isStaticAssetPath(path string) bool {
	return strings.HasPrefix(path, "/assets/") ||
		strings.HasPrefix(path, "/public/") ||
		strings.HasPrefix(path, "/files/")
}

func (rl *RateLimiter) limitsForKey(r *http.Request) (string, rate.Limit, int) {
	if userID := userIDFromRequest(r); userID != "" {
		mult := rate.Limit(rl.userMultiplier)
		return "user:" + userID, rl.rate * mult, rl.burst * rl.userMultiplier
	}

	ip := IP(r)
	if ip == "" {
		ip = "unknown"
	}
	return "ip:" + ip, rl.rate, rl.burst
}

func userIDFromRequest(r *http.Request) string {
	tokenString := getToken(r)
	if tokenString == "" || strings.HasPrefix(tokenString, "Bearer ") {
		return ""
	}

	secret := config.C.Auth.JWTSecret
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if uid, ok := claims["user_id"].(string); ok {
			return uid
		}
	}
	return ""
}

func (rl *RateLimiter) getLimiter(key string, limit rate.Limit, burst int) *limiterEntry {
	rl.mu.RLock()
	entry, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return entry
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if entry, exists = rl.limiters[key]; exists {
		return entry
	}

	entry = &limiterEntry{
		limiter: rate.NewLimiter(limit, burst),
		burst:   burst,
	}
	rl.limiters[key] = entry

	return entry
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for key, entry := range rl.limiters {
			if entry.limiter.Tokens() >= float64(entry.burst) {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}
