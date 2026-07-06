package api_v1

import (
	"net/http"
	"strconv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func GetAllPausen(c *handler.Context) error {
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

	wettLs := []sqlc.Wettkampf{}
	if len(showWettkampfStr) == 0 {
		wettLs = crud.AllWettkampf
	} else {
		wettLs = []sqlc.Wettkampf{showWettkampf.Wettkampf}
	}
	pLs, err := crud.GetPausenByWettkampf(c.Request.Context(), wettLs)
	if err != nil {
		return err
	}
	if pLs == nil {
		pLs = []crud.Pause{}
	}
	return c.JSON(pLs)
}

func GetPause(c *handler.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: "invalid id"}
	}

	p, err := crud.GetPause(c.Request.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(p)
}

func DeletePause(c *handler.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return &handler.Error{StatusCode: http.StatusBadRequest, Message: "invalid id"}
	}

	err = crud.DeletePause(c.Request.Context(), int32(id))
	if err != nil {
		return err
	}
	return c.JSON("Pause erfolgreich gelöscht!")
}

func CreatePause(c *handler.Context) error {
	params := new(sqlc.CreatePauseParams)
	err := c.BodyParser(params)
	if err != nil {
		return err
	}

	p, err := crud.CreatePause(c.Request.Context(), *params)
	if err != nil {
		return err
	}

	return c.JSON(p)
}

func UpdatePause(c *handler.Context) error {
	params := new(sqlc.UpdatePauseParams)
	err := c.BodyParser(params)
	if err != nil {
		return err
	}

	p, err := crud.UpdatePause(c.Request.Context(), *params)
	if err != nil {
		return err
	}

	return c.JSON(p)
}
