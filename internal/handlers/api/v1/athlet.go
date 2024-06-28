package api_v1

import (
	"github.com/gofiber/fiber/v2"

	"github.com/bata94/RegattaApi/internal/crud"
	api "github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func GetAthlet(c *fiber.Ctx) error {
	id, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	a, err := crud.GetAthletMinimal(*id)
	if err != nil {
		return err
	}

	return c.JSON(a)
}

func GetAllAthlet(c *fiber.Ctx) error {
	aLs, err := crud.GetAllAthlet()
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}

func CreateAthlet(c *fiber.Ctx) error {
	aParams := new(sqlc.CreateAthletParams)
	err := c.BodyParser(&aParams)
	if err != nil {
		retErr := api.BAD_REQUEST
		retErr.Msg = err.Error()
		return &retErr
	}

	a, err := crud.CreateAthlet(*aParams)
	if err != nil {
		return err
	}

	return c.JSON(a)
}
