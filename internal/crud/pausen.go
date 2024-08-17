package crud

import (
	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

type Pause struct {
	sqlc.Pause
}

func GetAllPausen() ([]Pause, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	pLs := []Pause{}
	q, err := DB.Queries.GetAllPause(ctx)
	if err != nil {
		return nil, err
	}

	for _, p := range q {
		pLs = append(pLs, Pause{
			Pause: p,
		})
	}

	return pLs, err
}

func GetPause(id int) (Pause, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	p, err := DB.Queries.GetPause(ctx, int32(id))
	if err != nil {
		if isNoRowError(err) {
			return Pause{}, &api.NOT_FOUND
		}
		return Pause{}, err
	}

	return Pause{Pause: p}, nil
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

func CreatePause(params sqlc.CreatePauseParams) (Pause, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	p, err := DB.Queries.CreatePause(ctx, params)
	if err != nil {
		return Pause{}, err
	}

	return Pause{Pause: p}, nil
}

func UpdatePause(params sqlc.UpdatePauseParams) (Pause, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	p, err := DB.Queries.UpdatePause(ctx, params)
	if err != nil {
		return Pause{}, err
	}

	return Pause{Pause: p}, nil
}
