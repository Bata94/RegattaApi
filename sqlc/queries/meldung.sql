-- name: GetMeldungMinimal :one
SELECT * FROM meldung
WHERE uuid = $1 LIMIT 1;

-- name: GetAllMeldung :many
SELECT * FROM meldung
ORDER BY start_nummer ASC;

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
  kosten,
  typ,
  bemerkung
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;
