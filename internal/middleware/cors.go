package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/handler"
)

func CORS() Middleware {
	originsEnv := config.C.CORS.AllowedOrigins
	methods := config.C.CORS.AllowedMethods
	headers := config.C.CORS.AllowedHeaders

	allowedOrigins := strings.Split(originsEnv, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			origin := c.Request.Header.Get("Origin")
			slog.Debug(fmt.Sprintf("CORS DEBUG: Origin=%q Allowed=%v Method=%s", origin, allowedOrigins, c.Method()))
			matchedOrigin := matchOrigin(origin, allowedOrigins)
			slog.Debug(fmt.Sprintf("CORS DEBUG: Matched=%q", matchedOrigin))

			if matchedOrigin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", matchedOrigin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			} else if len(allowedOrigins) > 0 && allowedOrigins[0] == "*" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}

			c.Writer.Header().Set("Access-Control-Allow-Methods", methods)
			c.Writer.Header().Set("Access-Control-Allow-Headers", headers)

			if c.Method() == "OPTIONS" {
				c.Status(http.StatusOK)
				return nil
			}
			return next(c)
		}
	}
}

func matchOrigin(origin string, allowedOrigins []string) string {
	if origin == "" {
		return ""
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return origin
		}
		if strings.HasPrefix(allowed, "*.") {
			prefix := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(origin, prefix) {
				return origin
			}
			continue
		}
		if strings.HasSuffix(allowed, ":*") {
			hostPattern := strings.TrimSuffix(allowed, ":*")
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