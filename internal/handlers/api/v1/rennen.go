package api_v1

import (
	"cmp"
	"context"
	"slices"

	"github.com/bata94/RegattaApi/internal/crud"
	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TODO: Implement queryParams
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

	// TODO: Move into CRUD and add Wettkampf param
	if getAthleten {
		retLs := []crud.RennenWithMeldungAndAthlet{}

		rLs, err := DB.Queries.GetAllRennen(context.Background())
		if err != nil {
			return err
		}
		qLs, err := DB.Queries.GetAllRennenWithAthlet(context.Background(), []sqlc.Wettkampf{sqlc.WettkampfLangstrecke, sqlc.WettkampfSlalom, sqlc.WettkampfKurzstrecke, sqlc.WettkampfStaffel})
		if err != nil {
			return err
		}

		for _, r := range rLs {
			retLs = append(retLs, crud.RennenWithMeldungAndAthlet{
				Rennen:    crud.RennenFromSqlc(r.Rennen, int(r.NumMeldungen), r.NumAbteilungen),
				Meldungen: []crud.Meldung{},
			})
		}

		i := 0
		for _, q := range qLs {
			for retLs[i].Uuid != q.Rennen.Uuid {
				i++
				continue
			}

			indexLastMeld := len(retLs[i].Meldungen) - 1
			if indexLastMeld < 0 || retLs[i].Meldungen[indexLastMeld].Uuid != q.Meldung.Uuid {
				retLs[i].Meldungen = append(retLs[i].Meldungen, crud.Meldung{
					MeldungMinimal: crud.SqlcMeldungMinmalToCrudMeldungMinimal(q.Meldung),
					Verein:         q.Verein,
					Athleten: []crud.AthletWithPos{{
						Athlet:   q.Athlet,
						Rolle:    q.Rolle,
						Position: int(q.Position),
					}},
				})
			} else {
				retLs[i].Meldungen[indexLastMeld].Athleten = append(retLs[i].Meldungen[indexLastMeld].Athleten, crud.AthletWithPos{
					Athlet:   q.Athlet,
					Rolle:    q.Rolle,
					Position: int(q.Position),
				})
			}
		}

		log.Debug("Sorting Meldungen")
		for _, r := range retLs {
			// sort Meldungen
			slices.SortFunc(r.Meldungen, func(a, b crud.Meldung) int {
				return cmp.Or(
					cmp.Compare(a.Abteilung, b.Abteilung),
					cmp.Compare(a.Bahn, b.Bahn),
				)
			})
		}

		return c.JSON(retLs)
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
