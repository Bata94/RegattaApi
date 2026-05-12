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
			tokenString := c.Cookie("auth_token")
			if tokenString == "" {
				authHeader := c.Headers().Get("Authorization")
				if authHeader != "" {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenString == "" {
				return &handler.Error{StatusCode: 401, Message: "Missing authentication token"}
			}

			if strings.HasPrefix(tokenString, "Bearer ") {
				return &handler.Error{StatusCode: 401, Message: "Invalid token format"}
			}

			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				return &handler.Error{StatusCode: 401, Message: "Invalid or expired token"}
			}

			extractAuthData(c, token)
			return next(c)
		}
	}
}

func OptionalAuth() Middleware {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "DO_NOT_USE_IN_PROD"
	}

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
