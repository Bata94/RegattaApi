-- name: GetZeitnahmeZiel :one
SELECT * FROM zeitnahme_ziel
WHERE id = $1 LIMIT 1;

-- name: GetAllZeitnahmeZiel :many
SELECT * FROM zeitnahme_ziel
ORDER BY id ASC;

-- name: GetAllOpenZeitnahmeZiel :many
SELECT * FROM zeitnahme_ziel
WHERE verarbeitet = false
ORDER BY id DESC;

-- name: CreateZeitnahmeZiel :one
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
) RETURNING *;

-- name: UpdateZeitnahmeZiel :one
UPDATE
  zeitnahme_ziel
SET
  rennen_nummer = $2,
  start_nummer = $3
WHERE 
  id = $1
RETURNING 
  *;

-- name: DeleteZeitnahmeZiel :one
DELETE FROM zeitnahme_ziel
WHERE id = $1
RETURNING *;
