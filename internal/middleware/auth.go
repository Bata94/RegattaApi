package middleware

import (
	"os"
	"strings"

	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() Middleware {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "DO_NOT_USE_IN_PROD"
	}

	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			authHeader := c.Headers().Get("Authorization")
			if authHeader == "" {
				return &handler.Error{StatusCode: 401, Message: "Missing Authorization header"}
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return &handler.Error{StatusCode: 401, Message: "Invalid Bearer token format"}
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return &handler.Error{StatusCode: 401, Message: "Invalid or expired token"}
			}

			c.Locals("user", token)
			return next(c)
		}
	}
}