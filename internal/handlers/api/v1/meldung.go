package api_v1

import (
	"github.com/gofiber/fiber/v2"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func GetAllMeldung(c *fiber.Ctx) error {
	mLs, err := crud.GetAllMeldungen()
	if err != nil {
		return err
	}
	if mLs == nil {
		mLs = []sqlc.Meldung{}
	}

	return c.JSON(mLs)
}

func GetMeldung(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	m, err := crud.GetMeldungMinimal(*uuid)
	if err != nil {
		return err
	}

	return c.JSON(m)
}

func UpdateSetzungBatch(c *fiber.Ctx) error {
	params := new(crud.UpdateSetzungBatchParams)
	err := c.BodyParser(params)
	if err != nil {
		return err
	}

	err = crud.UpdateSetzungBatch(*params)
	if err != nil {
		return err
	}

	return c.JSON("Setzung erfolgreich aktualisiert!")
}

func GetAllMeldungForVerein(c *fiber.Ctx) error {
	vereinUuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	meldungen, err := crud.GetAllMeldungForVerein(*vereinUuid)
	if err != nil {
		return err
	}

	return c.JSON(meldungen)
}
