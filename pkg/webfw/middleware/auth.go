package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() func(http.Handler) http.Handler {
	secret := config.C.Auth.JWTSecret

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := getToken(r)

			if tokenString == "" {
				slog.Warn(fmt.Sprintf("Auth middleware: missing token for %s %s", r.Method, r.URL.Path))
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("Missing authentication token")); err != nil {
					slog.Error("failed to write auth error", "error", err)
				}
				return
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				slog.Warn(fmt.Sprintf("Auth middleware: invalid token for %s %s: %v", r.Method, r.URL.Path, err))
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("Invalid or expired token")); err != nil {
					slog.Error("failed to write auth error", "error", err)
				}
				return
			}

			r = r.WithContext(webfw.WithAuthData(r.Context(), token))
			next.ServeHTTP(w, r)
		})
	}
}

func OptionalAuth() func(http.Handler) http.Handler {
	secret := config.C.Auth.JWTSecret

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := getToken(r)

			if tokenString != "" && !strings.HasPrefix(tokenString, "Bearer ") {
				token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
					return []byte(secret), nil
				})

				if err == nil && token.Valid {
					r = r.WithContext(webfw.WithAuthData(r.Context(), token))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireCap(capabilities ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !webfw.HasAllCapabilities(r, capabilities...) {
				slog.Warn(fmt.Sprintf("Capability check failed for %s %s. Required: %v", r.Method, r.URL.Path, capabilities))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getToken(r *http.Request) string {
	c, err := r.Cookie("auth_token")
	if err == nil && c != nil {
		return c.Value
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
