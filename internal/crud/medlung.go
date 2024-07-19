package crud

import (
	"fmt"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"

	"github.com/google/uuid"
)

type MeldungMinimal struct {
	Uuid               uuid.UUID `json:"uuid"`
	DrvRevisionUuid    uuid.UUID `json:"drv_revision_uuid"`
	Typ                string    `json:"typ"`
	Bemerkung          string    `json:"bemerkung"`
	Abgemeldet         bool      `json:"abgemeldet"`
	Dns                bool      `json:"dns"`
	Dnf                bool      `json:"dnf"`
	Dsq                bool      `json:"dsq"`
	ZeitnahmeBemerkung string    `json:"zeitnahme_bemerkung"`
	StartNummer        int       `json:"start_nummer"`
	Abteilung          int       `json:"abteilung"`
	Bahn               int       `json:"bahn"`
	Kosten             int       `json:"kosten"`
	VereinUuid         uuid.UUID `json:"verein_uuid"`
	RennenUuid         uuid.UUID `json:"rennen_uuid"`
}

type Meldung struct {
	MeldungMinimal
	Verein   sqlc.Verein     `json:"verein"`
	Athleten []AthletWithPos `json:"athleten"`
}

func SqlcMeldungMinmalToCrudMeldungMinimal(q sqlc.Meldung) MeldungMinimal {
	return MeldungMinimal{
		Uuid:               q.Uuid,
		DrvRevisionUuid:    q.DrvRevisionUuid,
		Typ:                q.Typ,
		Bemerkung:          q.Bemerkung.String,
		Abgemeldet:         q.Abgemeldet,
		Dns:                q.Dns,
		Dnf:                q.Dnf,
		Dsq:                q.Dsq,
		ZeitnahmeBemerkung: q.ZeitnahmeBemerkung.String,
		StartNummer:        int(q.StartNummer),
		Abteilung:          int(q.Abteilung),
		Bahn:               int(q.Bahn),
		Kosten:             int(q.Kosten),
		VereinUuid:         q.VereinUuid,
		RennenUuid:         q.RennenUuid,
	}
}

type UpdateSetzungBatchParams struct {
	RennenUUID uuid.UUID                         `json:"rennen_uuid"`
	Meldungen  []sqlc.UpdateMeldungSetzungParams `json:"meldungen"`
}

type CreateMeldungParams struct {
	sqlc.CreateMeldungParams
	Athleten []MeldungAthlet
}

type MeldungAthlet struct {
	Uuid     uuid.UUID  `json:"uuid"`
	Position int32      `json:"position"`
	Rolle    sqlc.Rolle `json:"rolle"`
}

func GetAllMeldungen() ([]sqlc.Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	mLs, err := DB.Queries.GetAllMeldung(ctx)
	if err != nil {
		return nil, err
	}
	if mLs == nil {
		mLs = []sqlc.Meldung{}
	}

	return mLs, nil
}

func GetMeldungMinimal(uuid uuid.UUID) (sqlc.Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	m, err := DB.Queries.GetMeldungMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return sqlc.Meldung{}, &api.NOT_FOUND
		}
		return sqlc.Meldung{}, err
	}

	return m, nil
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

func CreateMeldung(mParams CreateMeldungParams) (sqlc.Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	// TODO: Implement as Transaction!
	m, err := DB.Queries.CreateMeldung(ctx, mParams.CreateMeldungParams)
	if err != nil {
		return sqlc.Meldung{}, err
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
			return sqlc.Meldung{}, &retErr
		}
	}

	return m, nil
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
