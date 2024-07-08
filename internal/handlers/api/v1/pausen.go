package api_v1

import (
	"github.com/gofiber/fiber/v2"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func GetAllPausen(c *fiber.Ctx) error {
	pLs, err := crud.GetAllPausen()
	if err != nil {
		return err
	}
	if pLs == nil {
		pLs = []sqlc.Pause{}
	}
	return c.JSON(pLs)
}

func GetPause(c *fiber.Ctx) error {
	id, err := api.GetIdFromCtx(c)
	if err != nil {
		return err
	}

	p, err := crud.GetPause(id)
	if err != nil {
		return err
	}
	return c.JSON(p)
}

func DeletePause(c *fiber.Ctx) error {
	id, err := api.GetId32FromCtx(c)
	if err != nil {
		return err
	}

	err = crud.DeletePause(id)
	if err != nil {
		return err
	}
	return c.JSON("Pause erfolgreich gelöscht!")
}

func CreatePause(c *fiber.Ctx) error {
	params := new(sqlc.CreatePauseParams)
	err := c.BodyParser(params)
	if err != nil {
		return err
	}

	p, err := crud.CreatePause(*params)
	if err != nil {
		return err
	}

	return c.JSON(p)
}

func UpdatePause(c *fiber.Ctx) error {
	params := new(sqlc.UpdatePauseParams)
	err := c.BodyParser(params)
	if err != nil {
		return err
	}

	p, err := crud.UpdatePause(*params)
	if err != nil {
		return err
	}

	return c.JSON(p)
}
