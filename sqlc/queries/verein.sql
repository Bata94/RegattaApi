-- name: GetVereinMinimal :one
SELECT * FROM verein
WHERE uuid = $1 LIMIT 1;

-- name: GetAllVerein :many
SELECT * FROM verein
ORDER BY name ASC;

-- name: GetVerein :one
SELECT
  sqlc.embed(verein),
  (SELECT COALESCE(SUM(meldung.kosten), 0) FROM meldung WHERE verein.uuid = meldung.verein_uuid) as ges_kosten,
  (SELECT COALESCE(SUM(zahlung.amount), 0) FROM zahlung WHERE verein.uuid = zahlung.verein_uuid) as ges_zahlungen
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
  ulid ASC;

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
