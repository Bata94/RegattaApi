-- name: GetMeldungMinimal :one
SELECT * FROM meldung
WHERE uuid = $1 LIMIT 1;

-- name: GetAllMeldung :many
SELECT * FROM meldung
ORDER BY start_nummer ASC;

-- name: CheckMedlungSetzung :one
SELECT uuid, abteilung, bahn FROM meldung
WHERE abteilung != 0 AND bahn != 0 LIMIT 1;

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
