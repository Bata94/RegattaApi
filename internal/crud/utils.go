package crud

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func isNoRowError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func getCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 60*time.Second)
}