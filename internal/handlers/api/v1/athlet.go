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
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	a, err := crud.GetAthletMinimal(id)
	if err != nil {
		return err
	}

	return c.JSON(a)
}

func GetAllAthlet(c *handler.Context) error {
	aLs, err := crud.GetAllAthlet()
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}

func CreateAthlet(c *handler.Context) error {
	aParams := new(NewAthletParams)
	err := c.BodyParser(&aParams)
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	vereinUuid, err := uuid.Parse(aParams.VereinUUID)
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	var geschlecht sqlc.Geschlecht
	aParams.Geschlecht = strings.ToLower(aParams.Geschlecht)
	if aParams.Geschlecht == "m" {
		geschlecht = sqlc.GeschlechtM
	} else if aParams.Geschlecht == "f" || aParams.Geschlecht == "w" {
		geschlecht = sqlc.GeschlechtW
	} else if aParams.Geschlecht == "x" {
		geschlecht = sqlc.GeschlechtX
	}

	a, err := crud.CreateAthlet(sqlc.CreateAthletParams{
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

	ath, err := crud.GetAthletMinimal(id)
	if err != nil {
		return err
	}

	err = ath.UpdateStartberechtigung(p.Startberechtigt)
	if err != nil {
		return err
	}

	return c.JSON("Athlet erfolgreich angepasst!")
}

func GetAthletWaage(c *handler.Context) error {
	ls, err := crud.GetForAllVereineMissingAthlet(0)
	if err != nil {
		return err
	}
	return c.JSON(ls)
}

func GetAthletStartberechtigung(c *handler.Context) error {
	ls, err := crud.GetForAllVereineMissingAthlet(1)
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

	ath, err := crud.GetAthletMinimal(id)
	if err != nil {
		return err
	}

	err = ath.UpdateGewicht(p.Gewicht)
	if err != nil {
		return err
	}

	return c.JSON("Athlet erfolgreich angepasst!")
}
