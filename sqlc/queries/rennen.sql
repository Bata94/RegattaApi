-- name: GetRennenMinimal :one
SELECT * FROM rennen
WHERE uuid = $1 LIMIT 1;

-- name: GetRennen :many
SELECT *
FROM rennen
FULL JOIN meldung
ON rennen.uuid = meldung.rennen_uuid
WHERE rennen.uuid = $1;

-- name: GetAllRennen :many
SELECT * FROM rennen
ORDER BY sort_id ASC;

-- name: GetAllRennenWithMeld :many
SELECT sqlc.embed(rennen), meldung.*, verein.*
FROM rennen
FULL JOIN meldung
ON rennen.uuid = meldung.rennen_uuid
FULL JOIN verein
ON meldung.verein_uuid = verein.uuid
WHERE wettkampf = ANY($1::wettkampf[])
ORDER BY rennen.sort_id;

-- name: UpdateStartZeit :exec
UPDATE rennen SET startzeit = $1 WHERE uuid = $2;

-- name: CreateRennen :one
INSERT INTO rennen (
  uuid,
  sort_id,
  nummer,
  bezeichnung,
  bezeichnung_lang,
  zusatz,
  leichtgewicht,
  geschlecht,
  bootsklasse,
  bootsklasse_lang,
  altersklasse,
  altersklasse_lang,
  tag,
  wettkampf,
  kosten_eur,
  rennabstand
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING *;
