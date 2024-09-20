// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: zeitnahme_ziel.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createZeitnahmeZiel = `-- name: CreateZeitnahmeZiel :one
INSERT INTO zeitnahme_ziel (
  rennen_nummer,
  start_nummer,
  time_client,
  time_server,
  measured_latency
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
) RETURNING id, rennen_nummer, start_nummer, time_client, time_server, measured_latency, verarbeitet
`

type CreateZeitnahmeZielParams struct {
	RennenNummer    pgtype.Text      `json:"rennen_nummer"`
	StartNummer     pgtype.Text      `json:"start_nummer"`
	TimeClient      pgtype.Timestamp `json:"time_client"`
	TimeServer      pgtype.Timestamp `json:"time_server"`
	MeasuredLatency pgtype.Int4      `json:"measured_latency"`
}

func (q *Queries) CreateZeitnahmeZiel(ctx context.Context, arg CreateZeitnahmeZielParams) (ZeitnahmeZiel, error) {
	row := q.db.QueryRow(ctx, createZeitnahmeZiel,
		arg.RennenNummer,
		arg.StartNummer,
		arg.TimeClient,
		arg.TimeServer,
		arg.MeasuredLatency,
	)
	var i ZeitnahmeZiel
	err := row.Scan(
		&i.ID,
		&i.RennenNummer,
		&i.StartNummer,
		&i.TimeClient,
		&i.TimeServer,
		&i.MeasuredLatency,
		&i.Verarbeitet,
	)
	return i, err
}

const deleteZeitnahmeZiel = `-- name: DeleteZeitnahmeZiel :one
DELETE FROM zeitnahme_ziel
WHERE id = $1
RETURNING id, rennen_nummer, start_nummer, time_client, time_server, measured_latency, verarbeitet
`

func (q *Queries) DeleteZeitnahmeZiel(ctx context.Context, id int32) (ZeitnahmeZiel, error) {
	row := q.db.QueryRow(ctx, deleteZeitnahmeZiel, id)
	var i ZeitnahmeZiel
	err := row.Scan(
		&i.ID,
		&i.RennenNummer,
		&i.StartNummer,
		&i.TimeClient,
		&i.TimeServer,
		&i.MeasuredLatency,
		&i.Verarbeitet,
	)
	return i, err
}

const getAllOpenZeitnahmeZiel = `-- name: GetAllOpenZeitnahmeZiel :many
SELECT id, rennen_nummer, start_nummer, time_client, time_server, measured_latency, verarbeitet FROM zeitnahme_ziel
WHERE verarbeitet = false
ORDER BY id DESC
`

func (q *Queries) GetAllOpenZeitnahmeZiel(ctx context.Context) ([]ZeitnahmeZiel, error) {
	rows, err := q.db.Query(ctx, getAllOpenZeitnahmeZiel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ZeitnahmeZiel{}
	for rows.Next() {
		var i ZeitnahmeZiel
		if err := rows.Scan(
			&i.ID,
			&i.RennenNummer,
			&i.StartNummer,
			&i.TimeClient,
			&i.TimeServer,
			&i.MeasuredLatency,
			&i.Verarbeitet,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllZeitnahmeZiel = `-- name: GetAllZeitnahmeZiel :many
SELECT id, rennen_nummer, start_nummer, time_client, time_server, measured_latency, verarbeitet FROM zeitnahme_ziel
ORDER BY id ASC
`

func (q *Queries) GetAllZeitnahmeZiel(ctx context.Context) ([]ZeitnahmeZiel, error) {
	rows, err := q.db.Query(ctx, getAllZeitnahmeZiel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ZeitnahmeZiel{}
	for rows.Next() {
		var i ZeitnahmeZiel
		if err := rows.Scan(
			&i.ID,
			&i.RennenNummer,
			&i.StartNummer,
			&i.TimeClient,
			&i.TimeServer,
			&i.MeasuredLatency,
			&i.Verarbeitet,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getZeitnahmeZiel = `-- name: GetZeitnahmeZiel :one
SELECT id, rennen_nummer, start_nummer, time_client, time_server, measured_latency, verarbeitet FROM zeitnahme_ziel
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetZeitnahmeZiel(ctx context.Context, id int32) (ZeitnahmeZiel, error) {
	row := q.db.QueryRow(ctx, getZeitnahmeZiel, id)
	var i ZeitnahmeZiel
	err := row.Scan(
		&i.ID,
		&i.RennenNummer,
		&i.StartNummer,
		&i.TimeClient,
		&i.TimeServer,
		&i.MeasuredLatency,
		&i.Verarbeitet,
	)
	return i, err
}

const updateZeitnahmeZiel = `-- name: UpdateZeitnahmeZiel :one
UPDATE
  zeitnahme_ziel
SET
  rennen_nummer = $2,
  start_nummer = $3
WHERE 
  id = $1
RETURNING 
  id, rennen_nummer, start_nummer, time_client, time_server, measured_latency, verarbeitet
`

type UpdateZeitnahmeZielParams struct {
	ID           int32       `json:"id"`
	RennenNummer pgtype.Text `json:"rennen_nummer"`
	StartNummer  pgtype.Text `json:"start_nummer"`
}

func (q *Queries) UpdateZeitnahmeZiel(ctx context.Context, arg UpdateZeitnahmeZielParams) (ZeitnahmeZiel, error) {
	row := q.db.QueryRow(ctx, updateZeitnahmeZiel, arg.ID, arg.RennenNummer, arg.StartNummer)
	var i ZeitnahmeZiel
	err := row.Scan(
		&i.ID,
		&i.RennenNummer,
		&i.StartNummer,
		&i.TimeClient,
		&i.TimeServer,
		&i.MeasuredLatency,
		&i.Verarbeitet,
	)
	return i, err
}
