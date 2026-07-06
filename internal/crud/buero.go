package crud

import (
	"context"
	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

func CreateRechnung(ctx context.Context, nummer string, verein_uuid uuid.UUID, costSum int) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.Queries.CreateRechnung(ctx, sqlc.CreateRechnungParams{
		Nummer:     nummer,
		VereinUuid: verein_uuid,
		CostSum:    int32(costSum),
	})
}