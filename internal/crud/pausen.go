package crud

import (
	"context"
	"github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

type Pause struct {
	sqlc.Pause
}

func GetAllPausen(ctx context.Context) ([]Pause, error) {
	ctx, cancel := getCtx(ctx)
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

func GetPausenByWettkampf(ctx context.Context, w []sqlc.Wettkampf) ([]Pause, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	if len(w) == 0 {
		w = []sqlc.Wettkampf{sqlc.WettkampfLangstrecke, sqlc.WettkampfSlalom, sqlc.WettkampfKurzstrecke, sqlc.WettkampfStaffel}
	}

	pLs := []Pause{}
	q, err := DB.Queries.GetPausenByWettkampf(ctx, w)
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

func GetPause(ctx context.Context, id int) (Pause, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	p, err := DB.Queries.GetPause(ctx, int32(id))
	if err != nil {
		if isNoRowError(err) {
			return Pause{}, apierr.ErrNotFound
		}
		return Pause{}, err
	}

	return Pause{Pause: p}, nil
}

func DeletePause(ctx context.Context, id int32) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	err := DB.Queries.DeletePause(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func CreatePause(ctx context.Context, params sqlc.CreatePauseParams) (Pause, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	p, err := DB.Queries.CreatePause(ctx, params)
	if err != nil {
		return Pause{}, err
	}

	return Pause{Pause: p}, nil
}

func UpdatePause(ctx context.Context, params sqlc.UpdatePauseParams) (Pause, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	p, err := DB.Queries.UpdatePause(ctx, params)
	if err != nil {
		return Pause{}, err
	}

	return Pause{Pause: p}, nil
}
