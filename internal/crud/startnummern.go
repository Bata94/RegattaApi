package crud

import (
	"context"

	DB "github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/sqlc"
)

type StartnummernBereich struct {
	sqlc.StartnummernBereich
}

func StartnummernBereichFromSqlc(b sqlc.StartnummernBereich) StartnummernBereich {
	return StartnummernBereich{StartnummernBereich: b}
}

func GetStartnummernBereich(ctx context.Context) (StartnummernBereich, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	b, err := DB.QueriesFromCtx(ctx).GetStartnummernBereich(ctx)
	if err != nil {
		if isNoRowError(err) {
			return StartnummernBereich{}, apierr.ErrNotFound
		}
		return StartnummernBereich{}, err
	}

	return StartnummernBereich{StartnummernBereich: b}, nil
}

func SetStartnummernBereich(ctx context.Context, minNummer, maxNummer int32, fehlende []int32) (StartnummernBereich, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	b, err := DB.QueriesFromCtx(ctx).SetStartnummernBereich(ctx, sqlc.SetStartnummernBereichParams{
		MinNummer:       minNummer,
		MaxNummer:       maxNummer,
		FehlendeNummern: fehlende,
	})
	if err != nil {
		return StartnummernBereich{}, err
	}

	return StartnummernBereich{StartnummernBereich: b}, nil
}

func (b StartnummernBereich) InBereich(n int32) bool {
	return n >= b.MinNummer && n <= b.MaxNummer
}

func (b StartnummernBereich) IsFehlend(n int32) bool {
	for _, f := range b.FehlendeNummern {
		if f == n {
			return true
		}
	}
	return false
}
