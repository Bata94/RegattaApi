-- name: GetZeitnahmeStart :one
SELECT * FROM zeitnahme_start
WHERE id = $1 LIMIT 1;

-- name: GetAllZeitnahmeStart :many
SELECT * FROM zeitnahme_start
ORDER BY id ASC;

-- name: GetOpenStarts :many
SELECT * FROM zeitnahme_start
WHERE verarbeitet = false
ORDER BY id ASC;

-- name: CreateZeitnahmeStart :one
INSERT INTO zeitnahme_start (
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
) RETURNING *;

-- name: SetZeitnahmeStartVerarbeitet :exec
UPDATE zeitnahme_start
SET verarbeitet = true
WHERE id = $1;

-- name: DeleteZeitnahmeStart :exec
DELETE FROM zeitnahme_start
WHERE id = $1;
