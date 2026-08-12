package api_v1

import (
	"strings"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

type NewAthletParams struct {
	VereinUUID      string `json:"verein_uuid"`
	Name            string `json:"name"`
	Vorname         string `json:"vorname"`
	Jahrgang        string `json:"jahrgang"`
	Startberechtigt bool   `json:"startberechtigt"`
	Geschlecht      string `json:"geschlecht"`
}

func GetAthlet(c *handler.Context) error {
	id, err := c.GetUUID("uuid")
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	a, err := crud.GetAthletMinimal(c.Request.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(a)
}

func GetAllAthlet(c *handler.Context) error {
	aLs, err := crud.GetAllAthlet(c.Request.Context())
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}

func CreateAthlet(c *handler.Context) error {
	aParams := new(NewAthletParams)
	err := c.BodyParser(&aParams)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	vereinUuid, err := uuid.Parse(aParams.VereinUUID)
	if err != nil {
		return handler.BadRequest(err.Error())
	}

	var geschlecht sqlc.Geschlecht
	aParams.Geschlecht = strings.ToLower(aParams.Geschlecht)
	switch aParams.Geschlecht {
	case "m":
		geschlecht = sqlc.GeschlechtM
	case "f", "w":
		geschlecht = sqlc.GeschlechtW
	case "x":
		geschlecht = sqlc.GeschlechtX
	}

	a, err := crud.CreateAthlet(c.Request.Context(), sqlc.CreateAthletParams{
		Uuid:            uuid.New(),
		VereinUuid:      vereinUuid,
		Name:            aParams.Name,
		Vorname:         aParams.Vorname,
		Jahrgang:        aParams.Jahrgang,
		Startberechtigt: aParams.Startberechtigt,
		Geschlecht:      geschlecht,
	})
	if err != nil {
		return err
	}

	return c.JSON(a)
}

type UpdateAthletStartberechtigungParams struct {
	Uuid            string `json:"uuid"`
	Startberechtigt bool   `json:"startberechtigt"`
}

func UpdateAthletStartberechtigung(c *handler.Context) error {
	p := new(UpdateAthletStartberechtigungParams)
	err := c.BodyParser(p)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(p.Uuid)
	if err != nil {
		return err
	}

	ath, err := crud.GetAthletMinimal(c.Request.Context(), id)
	if err != nil {
		return err
	}

	err = ath.UpdateStartberechtigung(c.Request.Context(), p.Startberechtigt)
	if err != nil {
		return err
	}

	return c.JSON("Athlet erfolgreich angepasst!")
}

func GetAthletWaage(c *handler.Context) error {
	ls, err := crud.GetForAllVereineMissingAthlet(crud.Waage)
	if err != nil {
		return err
	}
	return c.JSON(ls)
}

func GetAthletStartberechtigung(c *handler.Context) error {
	ls, err := crud.GetForAllVereineMissingAthlet(crud.Startberechtigt)
	if err != nil {
		return err
	}
	return c.JSON(ls)
}

type UpdateAthletWaageParams struct {
	Uuid    string `json:"uuid"`
	Gewicht int    `json:"gewicht"`
}

func UpdateAthletWaage(c *handler.Context) error {
	p := new(UpdateAthletWaageParams)
	err := c.BodyParser(p)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(p.Uuid)
	if err != nil {
		return err
	}

	ath, err := crud.GetAthletMinimal(c.Request.Context(), id)
	if err != nil {
		return err
	}

	err = ath.UpdateGewicht(c.Request.Context(), p.Gewicht)
	if err != nil {
		return err
	}

	return c.JSON("Athlet erfolgreich angepasst!")
}
