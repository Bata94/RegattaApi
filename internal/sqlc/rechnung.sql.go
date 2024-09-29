// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: rechnung.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const createRechnung = `-- name: CreateRechnung :exec
INSERT INTO rechnung(
  nummer,
  verein_uuid,
  cost_sum
)VALUES(
  $1,
  $2,
  $3
)
`

type CreateRechnungParams struct {
	Nummer     string    `json:"nummer"`
	VereinUuid uuid.UUID `json:"verein_uuid"`
	CostSum    int32     `json:"cost_sum"`
}

func (q *Queries) CreateRechnung(ctx context.Context, arg CreateRechnungParams) error {
	_, err := q.db.Exec(ctx, createRechnung, arg.Nummer, arg.VereinUuid, arg.CostSum)
	return err
}