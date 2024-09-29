-- name: GetMeldungMinimal :one
SELECT * FROM meldung
WHERE uuid = $1 LIMIT 1;

-- name: GetMeldung :many
SELECT
  sqlc.embed(meldung),
  sqlc.embed(rennen),
  sqlc.embed(athlet),
  link_meldung_athlet.rolle, link_meldung_athlet.position,
  sqlc.embed(verein)
FROM
  meldung
JOIN
  rennen
ON
  rennen.uuid = meldung.rennen_uuid
JOIN
  link_meldung_athlet
ON
  link_meldung_athlet.meldung_uuid = meldung.uuid
JOIN
  athlet
ON
  link_meldung_athlet.athlet_uuid = athlet.uuid
JOIN
  verein
ON
  meldung.verein_uuid = verein.uuid
WHERE
  meldung.uuid = $1
ORDER BY
  rennen.sort_id, meldung.abteilung, meldung.bahn, link_meldung_athlet.rolle, link_meldung_athlet.position;

-- name: GetMeldungByStartNrUndTag :many
SELECT meldung.* FROM meldung
JOIN
  rennen
ON
  meldung.rennen_uuid = rennen.uuid
WHERE start_nummer = $1 AND rennen.tag = $2;

-- name: GetAllMeldung :many
SELECT * FROM meldung
ORDER BY start_nummer ASC;

-- name: GetLastStartnummer :one
SELECT
  MAX(meldung.start_nummer)
FROM
  meldung
JOIN
  rennen
ON
  meldung.rennen_uuid = rennen.uuid
WHERE
  rennen.tag = $1;

-- name: GetAllMeldungForVerein :many
SELECT
  sqlc.embed(meldung),
  sqlc.embed(rennen),
  sqlc.embed(athlet),
  link_meldung_athlet.rolle, link_meldung_athlet.position
FROM
  meldung
JOIN
  rennen
ON
  rennen.uuid = meldung.rennen_uuid
JOIN
  link_meldung_athlet
ON
  link_meldung_athlet.meldung_uuid = meldung.uuid
JOIN
  athlet
ON
  link_meldung_athlet.athlet_uuid = athlet.uuid
WHERE
  meldung.verein_uuid = $1
ORDER BY
  rennen.sort_id, meldung.abteilung, meldung.bahn, link_meldung_athlet.rolle, link_meldung_athlet.position;

-- name: CheckMedlungSetzung :one
SELECT uuid, abteilung, bahn FROM meldung
WHERE abteilung != 0 AND bahn != 0 LIMIT 1;

-- name: Ummeldung :exec
UPDATE 
  link_meldung_athlet
SET
  athlet_uuid = $4
WHERE
  meldung_uuid = $1 AND
  rolle = $2 AND
  position = $3;

-- name: Abmeldung :exec
UPDATE
  meldung
SET
  abgemeldet = true,
  abteilung = 0,
  bahn = 0
WHERE
  uuid = $1;

-- name: SetMeldungRechnungsNummer :exec
UPDATE meldung
SET rechnungs_nummer = $2
WHERE uuid = $1;

-- name: UpdateMeldungSetzung :exec
UPDATE meldung
SET abteilung = $2, bahn = $3
WHERE uuid = $1;

-- name: UpdateStartNummer :exec
UPDATE meldung
SET start_nummer = $2
WHERE uuid = $1;

-- name: CreateMeldung :one
INSERT INTO meldung (
  uuid,
  verein_uuid,
  rennen_uuid,
  drv_revision_uuid,
  abgemeldet,
  start_nummer,
  abteilung,
  bahn,
  kosten,
  typ,
  bemerkung
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;
