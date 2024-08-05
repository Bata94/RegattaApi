-- name: GetVereinMinimal :one
SELECT * FROM verein
WHERE uuid = $1 LIMIT 1;

-- name: GetAllVerein :many
SELECT * FROM verein
ORDER BY name ASC;

-- name: GetVereinRechnungsnummern :many
SELECT DISTINCT
   meldung.rechnungs_nummer
FROM
  meldung
INNER JOIN
  verein
ON
  meldung.verein_uuid = verein.uuid
WHERE
  verein.uuid = $1;

-- name: CreateVerein :one
INSERT INTO verein (
  uuid,
  name,
  kurzform,
  kuerzel
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;
