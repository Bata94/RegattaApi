package crud

import (
	"context"

	DB "github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

type Obmann struct {
	*sqlc.Obmann
}

func GetAllObmannForVerein(ctx context.Context, vereinUuid uuid.UUID) ([]Obmann, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	retLs := []Obmann{}
	q, err := DB.Queries.GetAllObmannForVerein(ctx, vereinUuid)
	if err != nil {
		return retLs, err
	}

	for _, o := range q {
		retLs = append(retLs, Obmann{&o})
	}
	return retLs, nil
}

func GetAllObmann(ctx context.Context) ([]Obmann, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	retLs := []Obmann{}
	q, err := DB.Queries.GetAllObmann(ctx)
	if err != nil {
		return retLs, err
	}

	for _, o := range q {
		retLs = append(retLs, Obmann{&o})
	}
	return retLs, nil
}

func GetObmannMinimal(ctx context.Context, uuid uuid.UUID) (Obmann, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	o, err := DB.QueriesFromCtx(ctx).GetObmannMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Obmann{}, apierr.ErrNotFound
		}
		return Obmann{}, err
	}

	return Obmann{&o}, nil
}

func CreateObmann(ctx context.Context, oParams sqlc.CreateObmannParams) (Obmann, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	o, err := DB.QueriesFromCtx(ctx).CreateObmann(ctx, oParams)
	if err != nil {
		return Obmann{}, err
	}

	return Obmann{&o}, nil
}

func UpdateObmann(ctx context.Context, uuid uuid.UUID, oParams sqlc.UpdateObmannParams) (Obmann, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	o, err := DB.QueriesFromCtx(ctx).UpdateObmann(ctx, sqlc.UpdateObmannParams{
		Uuid:       uuid,
		Name:       oParams.Name,
		Email:      oParams.Email,
		Phone:      oParams.Phone,
		VereinUuid: oParams.VereinUuid,
	})
	if err != nil {
		return Obmann{}, err
	}

	return Obmann{&o}, nil
}

func DeleteObmann(ctx context.Context, uuid uuid.UUID) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).DeleteObmann(ctx, uuid)
}
