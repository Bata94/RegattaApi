package crud

import (
	"context"

	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func EnqueueEmail(ctx context.Context, params sqlc.CreateEmailQueueEntryParams) (sqlc.EmailQueue, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).CreateEmailQueueEntry(ctx, params)
}

func ClaimNextEmail(ctx context.Context) (sqlc.EmailQueue, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).ClaimNextEmailQueueEntry(ctx)
}

func MarkEmailSent(ctx context.Context, uuid uuid.UUID) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).MarkEmailQueueSent(ctx, uuid)
}

func MarkEmailFailed(ctx context.Context, uuid uuid.UUID, backoffSecs float64, lastError string) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).MarkEmailQueueFailed(ctx, sqlc.MarkEmailQueueFailedParams{
		Uuid:      uuid,
		Secs:      backoffSecs,
		LastError: pgtype.Text{String: lastError, Valid: true},
	})
}

func GetAllEmailQueue(ctx context.Context) ([]sqlc.EmailQueue, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).GetEmailQueueEntries(ctx)
}

func ResetEmailQueue(ctx context.Context, uuid uuid.UUID) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).ResetEmailQueueEntry(ctx, uuid)
}

func DeleteEmailQueue(ctx context.Context, uuid uuid.UUID) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).DeleteEmailQueueEntry(ctx, uuid)
}
