package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetRennen(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	rennen, err := crud.GetRennen(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, rennen)
}

func GetAllRennen(w http.ResponseWriter, r *http.Request) {
	getMeld := webfw.Query(r, "getMeld") == "true"
	getAthleten := webfw.Query(r, "getAthleten") == "true"
	showEmpty := webfw.Query(r, "showEmpty") != "false"
	showStarted := webfw.Query(r, "showStarted") != "false"
	showWettkampfStr := webfw.Query(r, "wettkampf")
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
		webfw.APIError(w, webfw.BadRequest("Query param getAthleten requires getMeldungen to be true"))
		return
	}

	searchParams := crud.GetAllRennenParams{
		GetMeldungen:  getMeld,
		GetAthleten:   getAthleten,
		ShowEmpty:     showEmpty,
		ShowStarted:   showStarted,
		ShowWettkampf: showWettkampf,
	}

	if getAthleten {
		retLs, err := crud.GetAllRennenWithAthlet(r.Context(), searchParams)
		if err != nil {
			webfw.APIError(w, webfw.InternalError(err.Error()))
			return
		}

		webfw.JSON(w, r, retLs)
		return
	}

	rLs, err := crud.GetAllRennen(r.Context(), searchParams)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, rLs)
}
