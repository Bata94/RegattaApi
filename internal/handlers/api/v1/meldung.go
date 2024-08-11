package api_v1

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

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

type PostNachmeldungParams struct {
	VereinUuid                    string                        `json:"verein_uuid"`
	RennenUuid                    string                        `json:"rennen_uuid"`
	DoppeltesMeldentgeldBefreiung bool                          `json:"doppeltes_meldentgeld_befreiung"`
	Athleten                      []PostNachmeldungAthletParams `json:"athleten"`
}

type PostNachmeldungAthletParams struct {
	AthletUuid string `json:"uuid"`
	Position   string `json:"position"`
}

func PostNachmeldung(c *fiber.Ctx) error {
	params := new(PostNachmeldungParams)
	err := c.BodyParser(params)
	if err != nil {
		return err
	}

	vereinUuid, err := uuid.Parse(params.VereinUuid)
	if err != nil {
		return err
	}
	rennenUuid, err := uuid.Parse(params.RennenUuid)
	if err != nil {
		return err
	}

	rennen, err := crud.GetRennenMinimal(rennenUuid)
	if err != nil {
		return err
	}

	kosten := int32(rennen.KostenEur.Int32)
	if !params.DoppeltesMeldentgeldBefreiung {
		kosten = kosten * 2
	}

	mAth := []crud.MeldungAthlet{}
	for _, a := range params.Athleten {
		athUuid, err := uuid.Parse(a.AthletUuid)
		if err != nil {
			return err
		}

		var (
			athPostition int32
			athRolle     sqlc.Rolle
		)

		if a.Position == "stm" {
			athRolle = sqlc.RolleStm
			athPostition = 1
		} else {
			athRolle = sqlc.RolleRuderer
			athPostitionI64, err := strconv.ParseInt(a.Position, 10, 32)
			if err != nil {
				return err
			}
			athPostition = int32(athPostitionI64)
		}

		mAth = append(mAth, crud.MeldungAthlet{
			Uuid:     athUuid,
			Position: athPostition,
			Rolle:    athRolle,
		})
	}

	// TODO: add Startnummer, check athleten for doubles and num of entries, check Jahrgang & Geschlecht
	m, err := crud.CreateMeldung(crud.CreateMeldungParams{
		CreateMeldungParams: sqlc.CreateMeldungParams{
			Uuid:            uuid.New(),
			VereinUuid:      vereinUuid,
			RennenUuid:      rennen.Uuid,
			DrvRevisionUuid: uuid.New(),
			Abgemeldet:      false,
			Kosten:          kosten,
			Typ:             "Nachmeldung",
			Bemerkung:       pgtype.Text{},
		},
		Athleten: mAth,
	})

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
