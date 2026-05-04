package middleware

import (
	"strings"

	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/golang-jwt/jwt/v5"
)

func Protected() func(handler.Handler) handler.Handler {
	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			authHeader := c.Headers().Get("Authorization")
			if authHeader == "" {
				return &handler.Error{StatusCode: 401, Message: "Missing or malformed JWT"}
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return &handler.Error{StatusCode: 401, Message: "Missing or malformed JWT"}
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte("DO_NOT_USE_IN_PROD"), nil
			})

			if err != nil || !token.Valid {
				return &handler.Error{StatusCode: 401, Message: "Invalid token"}
			}

			c.Locals("user", token)
			return next(c)
		}
	}
}
