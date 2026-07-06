package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() Middleware {
	secret := config.C.Auth.JWTSecret

	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			tokenString := c.Cookie("auth_token")
			if tokenString == "" {
				authHeader := c.Headers().Get("Authorization")
				if authHeader != "" {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenString == "" {
				slog.Warn(fmt.Sprintf("Auth middleware: missing token for %s %s", c.Method(), c.Path()))
				return &handler.Error{StatusCode: http.StatusUnauthorized, Message: "Missing authentication token"}
			}

			if strings.HasPrefix(tokenString, "Bearer ") {
				return &handler.Error{StatusCode: http.StatusUnauthorized, Message: "Invalid token format"}
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				slog.Warn(fmt.Sprintf("Auth middleware: invalid token for %s %s: %v", c.Method(), c.Path(), err))
				return &handler.Error{StatusCode: http.StatusUnauthorized, Message: "Invalid or expired token"}
			}

			extractAuthData(c, token)
			return next(c)
		}
	}
}

func OptionalAuth() Middleware {
	secret := config.C.Auth.JWTSecret

	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			tokenString := c.Cookie("auth_token")
			if tokenString == "" {
				authHeader := c.Headers().Get("Authorization")
				if authHeader != "" {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenString != "" && !strings.HasPrefix(tokenString, "Bearer ") {
				token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
					return []byte(secret), nil
				})

				if err == nil && token.Valid {
					extractAuthData(c, token)
				}
			}

			return next(c)
		}
	}
}

func extractAuthData(c *handler.Context, token *jwt.Token) {
	claims := token.Claims.(jwt.MapClaims)
	c.Locals("logged_in", true)

	var capabilities []string
	capFields := []string{"allowed_logged_in", "allowed_admin", "allowed_zeitnahme", "allowed_startlisten", "allowed_regattabuero", "allowed_regattaleitung"}
	for _, field := range capFields {
		if val, exists := claims[field]; exists && val == true {
			capabilities = append(capabilities, field)
		}
	}
	c.Locals("capabilities", capabilities)
	c.Locals("user", token)
}
