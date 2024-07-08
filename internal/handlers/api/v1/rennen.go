package api_v1

import (
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetRennen(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	r, err := crud.GetRennen(*uuid)
	if err != nil {
		return err
	}

	return c.JSON(r)
}

func GetAllRennen(c *fiber.Ctx) error {
	getMeld := api.GetQueryParamBoolFromCtx(c, "getMeld", false)
	getAthleten := api.GetQueryParamBoolFromCtx(c, "getAthleten", false)
	showEmpty := api.GetQueryParamBoolFromCtx(c, "showEmpty", true)
	showStarted := api.GetQueryParamBoolFromCtx(c, "showStarted", true)
	showWettkampfStr := c.Query("wettkampf", "")
	showWettkampf := sqlc.NullWettkampf{}
	if showWettkampfStr != "" {
		caser := cases.Title(language.German)
		showWettkampfStr = caser.String(showWettkampfStr)
		showWettkampf = sqlc.NullWettkampf{
			Wettkampf: sqlc.Wettkampf(showWettkampfStr),
			Valid:     true,
		}
	}

	if getAthleten && !getMeld {
		retErr := &api.BAD_REQUEST
		retErr.Msg = "Query param getAthleten requires getMeldungen to be true"
		return retErr
	}

	rLs, err := crud.GetAllRennen(crud.GetAllRennenParams{
		GetMeldungen:  getMeld,
		GetAthleten:   getAthleten,
		ShowEmpty:     showEmpty,
		ShowStarted:   showStarted,
		ShowWettkampf: showWettkampf,
	})
	if err != nil {
		return err
	}

	return c.JSON(rLs)
}
