package middleware

import (
	"github.com/bata94/RegattaApi/internal/handlers/api"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

// Protected protect routes
func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
    // TODO: RM Secret from Codebase till Prod
		SigningKey: jwtware.SigningKey{Key: []byte("DO_NOT_USE_IN_PROD")},
		ErrorHandler: jwtError,
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
    return &api.TOKEN_GENERATION_ERROR
	}
  return &api.TOKEN_INVALID
}
