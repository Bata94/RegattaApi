package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const contextKeyLocals contextKey = "webfw_locals"

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

			if strings.HasPrefix(tokenString, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte("Invalid token format")); err != nil {
					slog.Error("failed to write auth error", "error", err)
				}
				return
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
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

			r = r.WithContext(withAuthData(r.Context(), token))
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
				token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
					return []byte(secret), nil
				})

				if err == nil && token.Valid {
					r = r.WithContext(withAuthData(r.Context(), token))
				}
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

func withAuthData(ctx context.Context, token *jwt.Token) context.Context {
	claims := token.Claims.(jwt.MapClaims)

	locals := make(map[string]any)
	locals["logged_in"] = true

	var capabilities []string
	if capsRaw, ok := claims["capabilities"].([]any); ok {
		for _, c := range capsRaw {
			if s, ok := c.(string); ok {
				capabilities = append(capabilities, s)
			}
		}
	}
	capabilities = append(capabilities, "allowed_logged_in")
	locals["capabilities"] = capabilities
	locals["user"] = token

	return context.WithValue(ctx, contextKeyLocals, locals)
}
