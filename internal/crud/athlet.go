package crud

import (
	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type AthletWithPos struct {
	sqlc.Athlet
	Rolle    sqlc.Rolle `json:"rolle"`
	Position int        `json:"position"`
}

func GetAthletMinimal(uuid uuid.UUID) (sqlc.Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	a, err := DB.Queries.GetAthletMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return sqlc.Athlet{}, &api.NOT_FOUND
		}
		return sqlc.Athlet{}, err
	}

	return a, nil
}

func GetAllAthlet() ([]sqlc.Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	aLs, err := DB.Queries.GetAllAthlet(ctx)
	if err != nil {
		return nil, err
	}
	if aLs == nil {
		aLs = []sqlc.Athlet{}
	}

	return aLs, err
}

func GetAllNNAthleten() ([]sqlc.Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	aLs, err := DB.Queries.GetAllNNAthleten(ctx)
	if err != nil {
		return nil, err
	}
	if aLs == nil {
		aLs = []sqlc.Athlet{}
	}

	return aLs, err
}

func CreateAthlet(aParams sqlc.CreateAthletParams) (sqlc.Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	a, err := DB.Queries.CreateAthlet(ctx, aParams)
	if err != nil {
		log.Error(err.Error())
		return sqlc.Athlet{}, err
	}

	return a, nil
}
