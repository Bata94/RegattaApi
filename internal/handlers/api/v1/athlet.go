package api_v1

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"

	"github.com/bata94/RegattaApi/internal/crud"
	api "github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func GetAthlet(c *fiber.Ctx) error {
	id, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	a, err := crud.GetAthletMinimal(*id)
	if err != nil {
		return err
	}

	return c.JSON(a)
}

func GetAllAthlet(c *fiber.Ctx) error {
	aLs, err := crud.GetAllAthlet()
	if err != nil {
		return err
	}

	return c.JSON(aLs)
}

type NewAthletParams struct {
	VereinUUID      string `json:"verein_uuid"`
	Name            string `json:"name"`
	Vorname         string `json:"vorname"`
	Jahrgang        string `json:"jahrgang"`
	Startberechtigt bool   `json:"startberechtigt"`
	Geschlecht      string `json:"geschlecht"`
}

func CreateAthlet(c *fiber.Ctx) error {
	aParams := new(NewAthletParams)
	err := c.BodyParser(&aParams)
	if err != nil {
		retErr := api.BAD_REQUEST
		retErr.Msg = err.Error()
		return &retErr
	}

	vereinUuid, err := uuid.Parse(aParams.VereinUUID)
	if err != nil {
		retErr := api.BAD_REQUEST
		retErr.Msg = err.Error()
		return &retErr
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

func UpdateAthletStartberechtigung(c *fiber.Ctx) error {
	p := new(UpdateAthletStartberechtigungParams)
	err := c.BodyParser(p)
	if err != nil {
		return err
	}

	uuid, err := uuid.Parse(p.Uuid)
	if err != nil {
		return err
	}

	err = crud.UpdateAthletStartberechtigung(p.Startberechtigt, uuid)

	return c.JSON("Athlet erfolgreich angepasst!")
}

func GetAthletWaage(c *fiber.Ctx) error {
	ls, err := crud.GetForAllVereineMissingAthlet(0)
	if err != nil {
		return err
	}
	return c.JSON(ls)
}

func GetAthletStartberechtigung(c *fiber.Ctx) error {
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

func UpdateAthletWaage(c *fiber.Ctx) error {
	p := new(UpdateAthletWaageParams)
	err := c.BodyParser(p)
	if err != nil {
		return err
	}

	uuid, err := uuid.Parse(p.Uuid)
	if err != nil {
		return err
	}

	err = crud.UpdateAthletGewicht(int32(p.Gewicht), uuid)

	return c.JSON("Athlet erfolgreich angepasst!")
}
