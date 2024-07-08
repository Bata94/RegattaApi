package crud

import (
	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

func GetAllVerein() ([]sqlc.Verein, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	vLs, err := DB.Queries.GetAllVerein(ctx)
	if err != nil {
		return nil, err
	}
	if vLs == nil {
		vLs = []sqlc.Verein{}
	}

	return vLs, err
}

func GetVereinMinimal(uuid uuid.UUID) (sqlc.Verein, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.GetVereinMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return sqlc.Verein{}, &api.NOT_FOUND
		}
		return sqlc.Verein{}, err
	}

	return v, nil
}

func CreateVerein(vParams sqlc.CreateVereinParams) (sqlc.Verein, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.CreateVerein(ctx, vParams)
	if err != nil {
		return sqlc.Verein{}, err
	}

	return v, nil
}
