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

	return api.JSON(c, vLs)
}

func GetVerein(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	v, err := crud.GetVerein(*uuid)
	if err != nil {
		return err
	}

	return api.JSON(c, v)
}

func GetAllAthletenForVerein(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	aLs, err := crud.GetAllAthletenForVerein(*uuid)
	if err != nil {
		return err
	}

	return api.JSON(c, aLs)
}

func GetAllAthletenForVereinMissStartber(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	aLs, err := crud.GetAllAthletenForVereinMissStartber(*uuid)
	if err != nil {
		return err
	}

	return api.JSON(c, aLs)
}

func GetAllAthletenForVereinWaage(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	aLs, err := crud.GetAllAthletenForVereinWaage(*uuid)
	if err != nil {
		return err
	}

	return api.JSON(c, aLs)
}
