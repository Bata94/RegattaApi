package crud

import (
	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

func GetAllPausen() ([]sqlc.Pause, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	pLs, err := DB.Queries.GetAllPause(ctx)
	if err != nil {
		return nil, err
	}
	if pLs == nil {
		pLs = []sqlc.Pause{}
	}

	return pLs, err
}

func GetPause(id int) (sqlc.Pause, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	p, err := DB.Queries.GetPause(ctx, int32(id))
	if err != nil {
		if isNoRowError(err) {
			return sqlc.Pause{}, &api.NOT_FOUND
		}
		return sqlc.Pause{}, err
	}

	return p, err
}

func DeletePause(id int32) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	err := DB.Queries.DeletePause(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func CreatePause(params sqlc.CreatePauseParams) (sqlc.Pause, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	p, err := DB.Queries.CreatePause(ctx, params)
	if err != nil {
		return sqlc.Pause{}, err
	}

	return p, nil
}

func UpdatePause(params sqlc.UpdatePauseParams) (sqlc.Pause, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	p, err := DB.Queries.UpdatePause(ctx, params)
	if err != nil {
		return sqlc.Pause{}, err
	}

	return p, err
}
