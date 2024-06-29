package api_v1

import (
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/gofiber/fiber/v2"
)

func Login(c *fiber.Ctx) error {
	loginParams := new(crud.LoginParams)
	if err := c.BodyParser(&loginParams); err != nil {
		retErr := api.BAD_REQUEST
		retErr.Details = err.Error()
		return &retErr
	}

	u, err := crud.AuthLogin(*loginParams)
	if err != nil {
		return err
	}

	return c.JSON(u)
}
