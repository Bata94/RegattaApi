package api_v1

import (
	"log"
	"strconv"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetAllMeldung(c *handler.Context) error {
	mLs, err := crud.GetAllMeldungen()
	if err != nil {
		return err
	}
	if mLs == nil {
		mLs = []crud.Meldung{}
	}

	return c.JSON(mLs)
}

func GetMeldung(c *handler.Context) error {
	meldungUUID, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	m, err := crud.GetMeldung(meldungUUID)
	if err != nil {
		return err
	}

	return c.JSON(m)
}

func PostAbmeldung(c *handler.Context) error {
	params := new(AbmeldungsParams)
	c.BodyParser(params)

	meldungUUID, err := uuid.Parse(params.Uuid)
	if err != nil {
		return err
	}

	err = crud.Abmeldung(meldungUUID)
	if err != nil {
		return err
	}

	return c.JSON("Meldung erfolgreich abgemeldet!")
}

type PostUmmeldungsParams struct {
	MeldungUuid string                        `json:"meldung_uuid"`
	Athleten    []PostNachmeldungAthletParams `json:"athleten"`
}

func PostUmmeldung(c *handler.Context) error {
	params := new(PostUmmeldungsParams)
	c.BodyParser(params)
	meldungUuid, err := uuid.Parse(params.MeldungUuid)
	if err != nil {
		return err
	}

	for _, a := range params.Athleten {
		athUuid, err := uuid.Parse(a.AthletUuid)
		if err != nil {
			return err
		}

		var (
			rolle    sqlc.Rolle
			position int32
		)

		if a.Position == "stm" {
			rolle = sqlc.RolleStm
			position = 1
		} else {
			rolle = sqlc.RolleRuderer
			positionI64, err := strconv.ParseInt(a.Position, 10, 32)
			if err != nil {
				return err
			}
			position = int32(positionI64)
		}

		err = crud.Ummeldung(sqlc.UmmeldungParams{
			MeldungUuid: meldungUuid,
			Rolle:       rolle,
			Position:    position,
			AthletUuid:  athUuid,
		})
		if err != nil {
			return err
		}
	}

	m, err := crud.GetMeldung(meldungUuid)
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

func PostNachmeldung(c *handler.Context) error {
	params := new(PostNachmeldungParams)
	err := c.BodyParser(params)
	if err != nil {
		log.Println("Parm parse Error: ", params)
		return err
	}

	vereinUuid, err := uuid.Parse(params.VereinUuid)
	if err != nil {
		log.Println("Verein Error: ", vereinUuid)
		return err
	}
	rennenUuid, err := uuid.Parse(params.RennenUuid)
	if err != nil {
		log.Println("Rennen Error: ", rennenUuid)
		return err
	}

	rennen, err := crud.GetRennen(rennenUuid)
	if err != nil {
		return err
	}

	kosten := int32(rennen.KostenEur)
	if !params.DoppeltesMeldentgeldBefreiung {
		kosten = kosten * 2
	}

	lastStrtNr, err := crud.GetStartnummerLast(rennen.Tag)
	if err != nil {
		return err
	}

	abteilung := int32(0)
	bahn := int32(0)
	maxBahn := 3

	if rennen.Wettkampf == sqlc.WettkampfLangstrecke {
		abteilung = int32(1)
		bahn = int32(*rennen.NumMeldungen + 1)
	} else {
		if rennen.Wettkampf == sqlc.WettkampfKurzstrecke {
			maxBahn = 4
		} else if rennen.Wettkampf == sqlc.WettkampfStaffel {
			maxBahn = 2
		} else if rennen.Wettkampf == sqlc.WettkampfSlalom {
			maxBahn = 3
		}
		if *rennen.NumMeldungen < maxBahn {
			abteilung = int32(1)
			bahn = int32(*rennen.NumMeldungen + 1)
		}
		for i, m := range rennen.Meldungen {
			if i == 0 {
				continue
			}
			if m.Bahn == 1 && rennen.Meldungen[i-1].Abteilung != m.Abteilung && rennen.Meldungen[i-1].Bahn < int32(maxBahn) {
				bahn = int32(rennen.Meldungen[i-1].Bahn + 1)
				abteilung = int32(rennen.Meldungen[i-1].Abteilung)
				break
			}
		}
		if rennen.Meldungen[len(rennen.Meldungen)-1].Bahn == int32(maxBahn) {
			bahn = int32(1)
			abteilung = int32(*rennen.NumAbteilungen + 1)
		} else {
			abteilung = int32(rennen.Meldungen[len(rennen.Meldungen)-1].Abteilung)
			bahn = int32(rennen.Meldungen[len(rennen.Meldungen)-1].Bahn + 1)
		}
	}

	mAth := []crud.CreateMeldungAthletParams{}
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

		mAth = append(mAth, crud.CreateMeldungAthletParams{
			Uuid:     athUuid,
			Position: athPostition,
			Rolle:    athRolle,
		})
	}

	m, err := crud.CreateMeldung(crud.CreateMeldungParams{
		CreateMeldungParams: sqlc.CreateMeldungParams{
			Uuid:            uuid.New(),
			VereinUuid:      vereinUuid,
			RennenUuid:      rennen.Uuid,
			DrvRevisionUuid: uuid.New(),
			StartNummer:     lastStrtNr + 1,
			Abteilung:       abteilung,
			Bahn:            bahn,
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

func UpdateSetzungBatch(c *handler.Context) error {
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

func GetAllMeldungForVerein(c *handler.Context) error {
	vereinUuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	meldungen, err := crud.GetAllMeldungForVerein(vereinUuid)
	if err != nil {
		return err
	}

	return c.JSON(meldungen)
}
