package api_v1

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetAllMeldung(w http.ResponseWriter, r *http.Request) {
	mLs, err := crud.GetAllMeldungen(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	if mLs == nil {
		mLs = []crud.Meldung{}
	}

	webfw.JSON(w, r, mLs)
}

func GetMeldung(w http.ResponseWriter, r *http.Request) {
	meldungUUID, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	m, err := crud.GetMeldung(r.Context(), meldungUUID)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, m)
}

func PostAbmeldung(w http.ResponseWriter, r *http.Request) {
	params := new(AbmeldungsParams)
	if err := webfw.ParseBody(r, params); err != nil {
		webfw.APIError(w, webfw.BadRequest("Invalid request body"))
		return
	}

	meldungUUID, err := uuid.Parse(params.Uuid)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	err = crud.Abmeldung(r.Context(), meldungUUID)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "Meldung erfolgreich abgemeldet!")
}

type PostUmmeldungsParams struct {
	MeldungUuid string                        `json:"meldung_uuid"`
	Athleten    []PostNachmeldungAthletParams `json:"athleten"`
}

func PostUmmeldung(w http.ResponseWriter, r *http.Request) {
	params := new(PostUmmeldungsParams)
	if err := webfw.ParseBody(r, params); err != nil {
		webfw.APIError(w, webfw.BadRequest("Invalid request body"))
		return
	}
	meldungUuid, err := uuid.Parse(params.MeldungUuid)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	for _, a := range params.Athleten {
		athUuid, err := uuid.Parse(a.AthletUuid)
		if err != nil {
			webfw.APIError(w, webfw.BadRequest(err.Error()))
			return
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
				webfw.APIError(w, webfw.BadRequest(err.Error()))
				return
			}
			position = int32(positionI64)
		}

		err = crud.Ummeldung(r.Context(), sqlc.UmmeldungParams{
			MeldungUuid: meldungUuid,
			Rolle:       rolle,
			Position:    position,
			AthletUuid:  athUuid,
		})
		if err != nil {
			webfw.APIError(w, webfw.InternalError(err.Error()))
			return
		}
	}

	m, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, m)
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

func CreateNachmeldung(ctx context.Context, params PostNachmeldungParams) (*crud.Meldung, error) {
	vereinUuid, err := uuid.Parse(params.VereinUuid)
	if err != nil {
		slog.Error("Verein Error", "uuid", vereinUuid)
		return nil, err
	}
	rennenUuid, err := uuid.Parse(params.RennenUuid)
	if err != nil {
		slog.Error("Rennen Error", "uuid", rennenUuid)
		return nil, err
	}

	rennen, err := crud.GetRennen(ctx, rennenUuid)
	if err != nil {
		return nil, err
	}

	kosten := int32(10)
	if v := rennen.GetKostenEur(); v != nil {
		kosten = int32(*v)
	}
	if !params.DoppeltesMeldentgeldBefreiung {
		kosten = kosten * 2
	}

	lastStrtNr, err := crud.GetStartnummerLast(ctx, rennen.Tag)
	if err != nil {
		return nil, err
	}

	abteilung := int32(0)
	bahn := int32(0)
	maxBahn := 3

	if rennen.Wettkampf == sqlc.WettkampfLangstrecke {
		abteilung = int32(1)
		bahn = int32(*rennen.NumMeldungen + 1)
	} else {
		switch rennen.Wettkampf {
		case sqlc.WettkampfKurzstrecke:
			maxBahn = 4
		case sqlc.WettkampfStaffel:
			maxBahn = 2
		case sqlc.WettkampfSlalom:
			maxBahn = 3
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
			return nil, err
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
				return nil, err
			}
			athPostition = int32(athPostitionI64)
		}

		mAth = append(mAth, crud.CreateMeldungAthletParams{
			Uuid:     athUuid,
			Position: athPostition,
			Rolle:    athRolle,
		})
	}

	m, err := crud.CreateMeldung(ctx, crud.CreateMeldungParams{
		Uuid:            uuid.NewV7(),
		VereinUuid:      vereinUuid,
		RennenUuid:      rennen.Uuid,
		DrvRevisionUuid: uuid.NewV7(),
		StartNummer:     lastStrtNr + 1,
		Abteilung:       abteilung,
		Bahn:            bahn,
		Abgemeldet:      false,
		Kosten:          kosten,
		Typ:             "Nachmeldung",
		Bemerkung:       pgtype.Text{},
		Athleten:        mAth,
	})

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func PostNachmeldung(w http.ResponseWriter, r *http.Request) {
	params := new(PostNachmeldungParams)
	err := webfw.ParseBody(r, params)
	if err != nil {
		slog.Error("Param parse Error", "params", params)
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	m, err := CreateNachmeldung(r.Context(), *params)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, m)
}

func UpdateSetzungBatch(w http.ResponseWriter, r *http.Request) {
	params := new(crud.UpdateSetzungBatchParams)
	err := webfw.ParseBody(r, params)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	err = crud.UpdateSetzungBatch(r.Context(), *params)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "Setzung erfolgreich aktualisiert!")
}

func GetAllMeldungForVerein(w http.ResponseWriter, r *http.Request) {
	vereinUuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	meldungen, err := crud.GetAllMeldungForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, meldungen)
}
