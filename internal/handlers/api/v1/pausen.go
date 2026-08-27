package api_v1

import (
	"net/http"
	"strconv"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetAllPausen(w http.ResponseWriter, r *http.Request) {
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

	var wettLs []sqlc.Wettkampf
	if len(showWettkampfStr) == 0 {
		wettLs = crud.AllWettkampf
	} else {
		wettLs = []sqlc.Wettkampf{showWettkampf.Wettkampf}
	}
	pLs, err := crud.GetPausenByWettkampf(r.Context(), wettLs)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	if pLs == nil {
		pLs = []crud.Pause{}
	}
	webfw.JSON(w, r, pLs)
}

func GetPause(w http.ResponseWriter, r *http.Request) {
	idStr := webfw.Param(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest("invalid id"))
		return
	}

	p, err := crud.GetPause(r.Context(), id)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	webfw.JSON(w, r, p)
}

func DeletePause(w http.ResponseWriter, r *http.Request) {
	idStr := webfw.Param(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest("invalid id"))
		return
	}

	err = crud.DeletePause(r.Context(), int32(id))
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	webfw.JSON(w, r, "Pause erfolgreich gelöscht!")
}

func CreatePause(w http.ResponseWriter, r *http.Request) {
	params := new(sqlc.CreatePauseParams)
	err := webfw.ParseBody(r, params)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	p, err := crud.CreatePause(r.Context(), *params)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, p)
}

func UpdatePause(w http.ResponseWriter, r *http.Request) {
	params := new(sqlc.UpdatePauseParams)
	err := webfw.ParseBody(r, params)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	p, err := crud.UpdatePause(r.Context(), *params)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, p)
}
