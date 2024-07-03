package api_v1

import (
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/gofiber/fiber/v2"
)

func GetAllVerein(c *fiber.Ctx) error {
	vLs, err := crud.GetAllVerein()
	if err != nil {
		return err
	}

	return c.JSON(vLs)
}

func GetVerein(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	v, err := crud.GetVereinMinimal(*uuid)
	if err != nil {
		return err
	}

	return c.JSON(v)
}
