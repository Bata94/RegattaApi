package crud

import (
	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func GetAllAthletenForVereinWaage(vUuid uuid.UUID) ([]sqlc.Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.GetAllAthletenForVereinWaage(ctx, vUuid)
}

func GetAllAthletenForVereinMissStartber(vUuid uuid.UUID) ([]sqlc.Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.GetAllAthletenForVereinMissStartber(ctx, vUuid)
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

func UpdateAthletStartberechtigung(startberechtigt bool, aUuid uuid.UUID) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.UpdateAthletAerztlBesch(ctx, sqlc.UpdateAthletAerztlBeschParams{
		Startberechtigt: startberechtigt,
		Uuid:            aUuid,
	})
}

func UpdateAthletGewicht(gewicht int32, aUuid uuid.UUID) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.UpdateAthletWaage(ctx, sqlc.UpdateAthletWaageParams{
		Gewicht: pgtype.Int4{Valid: true, Int32: gewicht},
		Uuid:    aUuid,
	})
}
