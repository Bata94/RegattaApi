package api_v1

import (
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetRennen(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	r, err := crud.GetRennen(uuid)
	if err != nil {
		return err
	}

	return c.JSON(r)
}

func GetAllRennen(c *handler.Context) error {
	getMeld := c.Query("getMeld") == "true"
	getAthleten := c.Query("getAthleten") == "true"
	showEmpty := c.Query("showEmpty") != "false"
	showStarted := c.Query("showStarted") != "false"
	showWettkampfStr := c.Query("wettkampf")
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
		return &handler.Error{StatusCode: 400, Message: "Query param getAthleten requires getMeldungen to be true"}
	}

	searchParams := crud.GetAllRennenParams{
		GetMeldungen:  getMeld,
		GetAthleten:   getAthleten,
		ShowEmpty:     showEmpty,
		ShowStarted:   showStarted,
		ShowWettkampf: showWettkampf,
	}

	if getAthleten {
		retLs, err := crud.GetAllRennenWithAthlet(searchParams)
		if err != nil {
			return err
		}

		return c.JSON(retLs)
	}

	rLs, err := crud.GetAllRennen(searchParams)
	if err != nil {
		return err
	}

	return c.JSON(rLs)
}
