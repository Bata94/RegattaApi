package crud

import (
	"errors"
	"fmt"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"
)

type Meldung struct {
	sqlc.Meldung
	Rennen   *Rennen  `json:"rennen"`
	Verein   *Verein  `json:"verein"`
	Athleten []Athlet `json:"athleten"`
}

type UpdateSetzungBatchParams struct {
	RennenUUID uuid.UUID                         `json:"rennen_uuid"`
	Meldungen  []sqlc.UpdateMeldungSetzungParams `json:"meldungen"`
}

type CreateMeldungParams struct {
	sqlc.CreateMeldungParams
	Athleten []CreateMeldungAthletParams
}

type CreateMeldungAthletParams struct {
	Uuid     uuid.UUID  `json:"uuid"`
	Position int32      `json:"position"`
	Rolle    sqlc.Rolle `json:"rolle"`
}

func GetAllMeldungen() ([]Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	mLs := []Meldung{}
	q, err := DB.Queries.GetAllMeldung(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range q {
		mLs = append(mLs, Meldung{
			Meldung: m,
		})
	}

	return mLs, nil
}

func GetMeldungMinimal(uuid uuid.UUID) (Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	m, err := DB.Queries.GetMeldungMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Meldung{}, &api.NOT_FOUND
		}
		return Meldung{}, err
	}

	return Meldung{Meldung: m}, nil
}

func CheckMeldungSetzung() (bool, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	_, err := DB.Queries.CheckMedlungSetzung(ctx)
	if err != nil {
		if isNoRowError(err) {
			return false, nil
		}
		return true, err
	}

	return true, nil
}

func CreateMeldung(mParams CreateMeldungParams) (Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	// TODO: Implement as Transaction!
	m, err := DB.Queries.CreateMeldung(ctx, mParams.CreateMeldungParams)
	if err != nil {
		return Meldung{}, err
	}

	for _, a := range mParams.Athleten {
		_, err = DB.Queries.CreateLinkMeldungAthlet(ctx, sqlc.CreateLinkMeldungAthletParams{
			MeldungUuid: m.Uuid,
			AthletUuid:  a.Uuid,
			Position:    a.Position,
			Rolle:       a.Rolle,
		})

		if err != nil {
			retErr := api.INTERNAL_SERVER_ERROR
			retErr.Details = fmt.Sprintf("Error linking MeldungAthlet: %s \nMeldung-ID: %s \nAthlet-ID: %s",
				err,
				m.Uuid.String(),
				a.Uuid.String(),
			)
			return Meldung{}, &retErr
		}
	}

	return Meldung{Meldung: m}, nil
}

func UpdateMeldungSetzung(p sqlc.UpdateMeldungSetzungParams) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.UpdateMeldungSetzung(ctx, p)
}

func UpdateSetzungBatch(p UpdateSetzungBatchParams) error {
	if len(p.Meldungen) == 0 {
		return &api.BAD_REQUEST
	}

	for _, m := range p.Meldungen {
		err := UpdateMeldungSetzung(sqlc.UpdateMeldungSetzungParams{
			Uuid:      m.Uuid,
			Abteilung: m.Abteilung,
			Bahn:      m.Bahn,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateStartNummer(p sqlc.UpdateStartNummerParams) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.UpdateStartNummer(ctx, p)
}

func GetAllMeldungForVerein(vereinUuid uuid.UUID) ([]Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	meldungen := []Meldung{}

	rows, err := DB.Queries.GetAllMeldungForVerein(ctx, vereinUuid)
	if err != nil {
		return meldungen, err
	}

	for i, r := range rows {
		if i == 0 || rows[i-1].Meldung.Uuid != r.Meldung.Uuid {
			rennen := RennenFromSqlc(r.Rennen, 0, int32(0))
			meldungen = append(meldungen, Meldung{
				Meldung:  r.Meldung,
				Rennen:   &rennen,
				Athleten: []Athlet{},
			})
		}

		curMeldung := &meldungen[len(meldungen)-1]

		position := int(r.Position)
		curMeldung.Athleten = append(curMeldung.Athleten, Athlet{
			Athlet:   r.Athlet,
			Rolle:    &r.Rolle,
			Position: &position,
		})
	}

	return meldungen, nil
}

func Abmeldung(meldUuid uuid.UUID) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.Abmeldung(ctx, meldUuid)
}

func SetMeldungRechnungsNummer(meldUuid uuid.UUID, rechnungsNummer string) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.SetMeldungRechnungsNummer(ctx, sqlc.SetMeldungRechnungsNummerParams{
		Uuid: meldUuid,
		RechnungsNummer: pgtype.Text{
			Valid:  true,
			String: rechnungsNummer,
		},
	})
}

func GetStartnummerLast(tag sqlc.Tag) (int32, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	lastStartNr, err := DB.Queries.GetLastStartnummer(ctx, tag)
	if err != nil {
		return 0, err
	}

	retInt, ok := lastStartNr.(int32)
	if !ok {
		return 0, errors.New("Last Startnummer nicht umwandelbar!")
	}

	return retInt, nil
}
