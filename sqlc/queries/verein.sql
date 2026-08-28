-- name: GetVereinMinimal :one
SELECT * FROM verein
WHERE uuid = $1 LIMIT 1;

-- name: GetAllVerein :many
SELECT * FROM verein
ORDER BY name ASC;

-- name: GetVerein :one
SELECT
  sqlc.embed(verein),
  (SELECT SUM(meldung.kosten) FROM meldung WHERE verein.uuid = meldung.verein_uuid) as ges_kosten,
  (SELECT SUM(zahlung.amount) FROM zahlung WHERE verein.uuid = zahlung.verein_uuid) as ges_zahlungen
FROM
  verein
WHERE
  verein.uuid = $1;

-- name: GetRechnungungenByVerein :many
SELECT
  *
FROM
  rechnung
WHERE
  verein_uuid = $1
ORDER BY
  uuid ASC;

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

-- name: UpdateVerein :one
UPDATE verein
SET
  name = $2,
  kurzform = $3,
  kuerzel = $4
WHERE uuid = $1
RETURNING *;

-- name: DeleteVerein :exec
DELETE FROM verein
WHERE uuid = $1;

-- name: CountAthletenForVerein :one
SELECT COUNT(*) FROM athlet
WHERE verein_uuid = $1;
