// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: athlet.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

const createAthlet = `-- name: CreateAthlet :one
INSERT INTO athlet
(uuid, verein_uuid, name, vorname, jahrgang, startberechtigt, geschlecht)
VALUES
($1, $2, $3, $4, $5, $6, $7)
RETURNING uuid, vorname, name, geschlecht, jahrgang, gewicht, startberechtigt, verein_uuid
`

type CreateAthletParams struct {
	Uuid            uuid.UUID  `json:"uuid"`
	VereinUuid      uuid.UUID  `json:"verein_uuid"`
	Name            string     `json:"name"`
	Vorname         string     `json:"vorname"`
	Jahrgang        string     `json:"jahrgang"`
	Startberechtigt *bool      `json:"startberechtigt"`
	Geschlecht      Geschlecht `json:"geschlecht"`
}

func (q *Queries) CreateAthlet(ctx context.Context, arg CreateAthletParams) (*Athlet, error) {
	row := q.db.QueryRow(ctx, createAthlet,
		arg.Uuid,
		arg.VereinUuid,
		arg.Name,
		arg.Vorname,
		arg.Jahrgang,
		arg.Startberechtigt,
		arg.Geschlecht,
	)
	var i Athlet
	err := row.Scan(
		&i.Uuid,
		&i.Vorname,
		&i.Name,
		&i.Geschlecht,
		&i.Jahrgang,
		&i.Gewicht,
		&i.Startberechtigt,
		&i.VereinUuid,
	)
	return &i, err
}

const getAllAthlet = `-- name: GetAllAthlet :many
SELECT uuid, vorname, name, geschlecht, jahrgang, gewicht, startberechtigt, verein_uuid FROM athlet
ORDER BY name ASC
`

func (q *Queries) GetAllAthlet(ctx context.Context) ([]*Athlet, error) {
	rows, err := q.db.Query(ctx, getAllAthlet)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Athlet
	for rows.Next() {
		var i Athlet
		if err := rows.Scan(
			&i.Uuid,
			&i.Vorname,
			&i.Name,
			&i.Geschlecht,
			&i.Jahrgang,
			&i.Gewicht,
			&i.Startberechtigt,
			&i.VereinUuid,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllNNAthleten = `-- name: GetAllNNAthleten :many
SELECT uuid, vorname, name, geschlecht, jahrgang, gewicht, startberechtigt, verein_uuid FROM athlet
WHERE vorname = 'No' and name = 'Name' and jahrgang = '9999'
ORDER BY verein_uuid ASC
`

func (q *Queries) GetAllNNAthleten(ctx context.Context) ([]*Athlet, error) {
	rows, err := q.db.Query(ctx, getAllNNAthleten)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*Athlet
	for rows.Next() {
		var i Athlet
		if err := rows.Scan(
			&i.Uuid,
			&i.Vorname,
			&i.Name,
			&i.Geschlecht,
			&i.Jahrgang,
			&i.Gewicht,
			&i.Startberechtigt,
			&i.VereinUuid,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAthletMinimal = `-- name: GetAthletMinimal :one
SELECT uuid, vorname, name, geschlecht, jahrgang, gewicht, startberechtigt, verein_uuid FROM athlet
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) GetAthletMinimal(ctx context.Context, argUuid uuid.UUID) (*Athlet, error) {
	row := q.db.QueryRow(ctx, getAthletMinimal, argUuid)
	var i Athlet
	err := row.Scan(
		&i.Uuid,
		&i.Vorname,
		&i.Name,
		&i.Geschlecht,
		&i.Jahrgang,
		&i.Gewicht,
		&i.Startberechtigt,
		&i.VereinUuid,
	)
	return &i, err
}
