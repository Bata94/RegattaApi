package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/bata94/RegattaApi/internal/config"
)

func CORS() func(http.Handler) http.Handler {
	originsEnv := config.C.CORS.AllowedOrigins
	methods := config.C.CORS.AllowedMethods
	headers := config.C.CORS.AllowedHeaders

	allowedOrigins := strings.Split(originsEnv, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	hasWildcard := slices.Contains(allowedOrigins, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			slog.Debug(fmt.Sprintf("CORS DEBUG: Origin=%q Allowed=%v Method=%s", origin, allowedOrigins, r.Method))

			if hasWildcard {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				matchedOrigin := matchOrigin(origin, allowedOrigins)
				slog.Debug(fmt.Sprintf("CORS DEBUG: Matched=%q", matchedOrigin))
				if matchedOrigin != "" {
					w.Header().Set("Access-Control-Allow-Origin", matchedOrigin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func matchOrigin(origin string, allowedOrigins []string) string {
	if origin == "" {
		return ""
	}

	for _, allowed := range allowedOrigins {
		if strings.HasPrefix(allowed, "*.") {
			prefix := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(origin, prefix) {
				return origin
			}
			continue
		}
		if before, ok := strings.CutSuffix(allowed, ":*"); ok {
			hostPattern := before
			if strings.HasPrefix(origin, hostPattern+":") {
				return origin
			}
			continue
		}
		if allowed == origin {
			return origin
		}
	}

	return ""
}
