package crud

import (
	"errors"

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

type VereineWithAthletMissing struct {
	Verein   sqlc.Verein       `json:"verein"`
	Athleten []AthletenMissing `json:"athleten"`
}

type AthletenMissing struct {
	Athlet       sqlc.Athlet `json:"athlet"`
	ErstesRennen sqlc.Rennen `json:"erstes_rennen"`
}

type MissingAthletType int

const (
	Waage           MissingAthletType = 0
	Startberechtigt MissingAthletType = 1
)

func GetForAllVereineMissingAthlet(athletType MissingAthletType) ([]VereineWithAthletMissing, error) {
	retLs := []VereineWithAthletMissing{}
	vLs, err := GetAllVerein()
	if err != nil {
		return retLs, err
	}

	var queryFunc func(uuid.UUID) ([]AthletenMissing, error)
	if athletType == 0 {
		queryFunc = GetAllAthletenForVereinWaage
	} else if athletType == 1 {
		queryFunc = GetAllAthletenForVereinMissStartber
	} else {
		return retLs, errors.New("Unknown athlet type")
	}

	for _, v := range vLs {
		missAthlet, err := queryFunc(v.Uuid)
		if err != nil {
			return retLs, err
		}

		if len(missAthlet) != 0 {
			retLs = append(retLs, VereineWithAthletMissing{
				Verein:   v,
				Athleten: missAthlet,
			})
		}
	}

	return retLs, nil
}

func GetAllAthletenForVereinWaage(vUuid uuid.UUID) ([]AthletenMissing, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	retLs := []AthletenMissing{}
	q, err := DB.Queries.GetAllAthletenForVereinWaage(ctx, vUuid)
	if err != nil {
		return retLs, err
	}

	for i, r := range q {
		if len(retLs) == 0 || (q[i-1].Athlet.Vorname != r.Athlet.Vorname && q[i-1].Athlet.Name != r.Athlet.Name) {
			retLs = append(retLs, AthletenMissing{
				Athlet:       r.Athlet,
				ErstesRennen: r.Rennen,
			})
			continue
		}
	}

	return retLs, nil
}

func GetAllAthletenForVereinMissStartber(vUuid uuid.UUID) ([]AthletenMissing, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	retLs := []AthletenMissing{}
	// TODO: klären, brauchen Stm. auch eine Startberechtigung?
	q, err := DB.Queries.GetAllAthletenForVereinMissStartber(ctx, vUuid)
	if err != nil {
		return retLs, err
	}

	for i, r := range q {
		if len(retLs) == 0 || (q[i-1].Athlet.Vorname != r.Athlet.Vorname && q[i-1].Athlet.Name != r.Athlet.Name) {
			retLs = append(retLs, AthletenMissing{
				Athlet:       r.Athlet,
				ErstesRennen: r.Rennen,
			})
			continue
		}
	}

	return retLs, nil
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
