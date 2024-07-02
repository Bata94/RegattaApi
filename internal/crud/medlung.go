package crud

import (
	"fmt"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

type CreateMeldungParams struct {
	*sqlc.CreateMeldungParams
	Athleten []MeldungAthlet
}

type MeldungAthlet struct {
	Uuid     uuid.UUID
	Position *int32
	Rolle    sqlc.Rolle
}

func GetAllMeldungen() ([]*sqlc.Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	mLs, err := DB.Queries.GetAllMeldung(ctx)
	if err != nil {
		return nil, err
	}
	if mLs == nil {
		mLs = []*sqlc.Meldung{}
	}

	return mLs, nil
}

func GetMeldungMinimal(uuid uuid.UUID) (*sqlc.Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	m, err := DB.Queries.GetMeldungMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return nil, &api.NOT_FOUND
		}
		return nil, err
	}

	return m, nil
}

func CheckMeldungSetzung() (bool, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	m, err := DB.Queries.CheckMedlungSetzung(ctx)
	if err != nil {
		if isNoRowError(err) {
			return false, nil
		}
		return true, err
	}

	if m != nil {
		return true, nil
	} else {
		return false, nil
	}
}

func CreateMeldung(mParams CreateMeldungParams) (*sqlc.Meldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	// TODO: Implement as Transaction!
	m, err := DB.Queries.CreateMeldung(ctx, *mParams.CreateMeldungParams)
	if err != nil {
		return nil, err
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
			return nil, &retErr
		}
	}

	return m, nil
}

func UpdateMeldungSetzung(p sqlc.UpdateMeldungSetzungParams) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.UpdateMeldungSetzung(ctx, p)

}
