package api_v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
)

type AbmeldungsParams struct {
	Uuid string `json:"uuid"`
}

func Abmeldung(c *fiber.Ctx) error {
	params := new(AbmeldungsParams)
	c.BodyParser(params)

	uuid, err := uuid.Parse(params.Uuid)
	if err != nil {
		return err
	}

	err = crud.Abmeldung(uuid)
	if err != nil {
		return err
	}

	return c.JSON("Meldung erfolgreich abgemeldet!")
}

func Ummeldung(c *fiber.Ctx) error {
	return &api.NOT_FOUND
}

func Nachmeldung(c *fiber.Ctx) error {
	return &api.NOT_FOUND
}

func StartnummernAusgabe(c *fiber.Ctx) error {
	return &api.NOT_FOUND
}

func StartnummernWechsel(c *fiber.Ctx) error {
	return &api.NOT_FOUND
}

func KasseEinzahlung(c *fiber.Ctx) error {
	return &api.NOT_FOUND
}
