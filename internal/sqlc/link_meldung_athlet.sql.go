// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: link_meldung_athlet.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const createLinkMeldungAthlet = `-- name: CreateLinkMeldungAthlet :one
INSERT INTO link_meldung_athlet (
  athlet_uuid,
  meldung_uuid,
  rolle,
  position
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING id, rolle, position, meldung_uuid, athlet_uuid
`

type CreateLinkMeldungAthletParams struct {
	AthletUuid  uuid.UUID `json:"athlet_uuid"`
	MeldungUuid uuid.UUID `json:"meldung_uuid"`
	Rolle       Rolle     `json:"rolle"`
	Position    int32     `json:"position"`
}

func (q *Queries) CreateLinkMeldungAthlet(ctx context.Context, arg CreateLinkMeldungAthletParams) (LinkMeldungAthlet, error) {
	row := q.db.QueryRow(ctx, createLinkMeldungAthlet,
		arg.AthletUuid,
		arg.MeldungUuid,
		arg.Rolle,
		arg.Position,
	)
	var i LinkMeldungAthlet
	err := row.Scan(
		&i.ID,
		&i.Rolle,
		&i.Position,
		&i.MeldungUuid,
		&i.AthletUuid,
	)
	return i, err
}
