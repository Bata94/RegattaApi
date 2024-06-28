package crud

import (
	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type CreateMeldungParams struct {
  *sqlc.CreateMeldungParams
  Athleten []MeldungAthlet
}

type MeldungAthlet struct {
  Uuid uuid.UUID
  Position *int32
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
    })

    if err != nil {
      log.Errorf("Error linking MeldungAthlet: %s \nMeldung-ID: %s \nAthlet-ID: %s", 
        err,
        m.Uuid.String(),
        a.Uuid.String(),
      )
    }
  }

	return m, nil
}
