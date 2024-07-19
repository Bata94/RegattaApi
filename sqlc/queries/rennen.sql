-- name: GetRennenMinimal :one
SELECT * FROM rennen
WHERE uuid = $1 LIMIT 1;

-- name: GetRennen :many
SELECT 
  sqlc.embed(rennen),
  sqlc.embed(meldung),
  sqlc.embed(athlet),
  sqlc.embed(verein),
  sqlc.embed(link_meldung_athlet)
FROM
  rennen
FULL JOIN
  meldung
ON
  rennen.uuid = meldung.rennen_uuid
FULL JOIN
  link_meldung_athlet 
ON
  meldung.uuid = link_meldung_athlet.meldung_uuid
FULL JOIN
  athlet
ON
  link_meldung_athlet.athlet_uuid = athlet.uuid
FULL JOIN
  verein
ON
  meldung.verein_uuid = verein.uuid
WHERE
  rennen.uuid = $1
ORDER BY
  meldung.abteilung, meldung.bahn, link_meldung_athlet.rolle, link_meldung_athlet.position;

-- name: GetAllRennen :many
SELECT sqlc.embed(rennen),
(SELECT COUNT(meldung.uuid) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid AND meldung.abgemeldet = false) as num_meldungen,
(SELECT COALESCE(MAX(meldung.abteilung),0) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid) as num_abteilungen
FROM rennen
ORDER BY sort_id ASC;

-- name: GetAllRennenWithMeld :many
SELECT sqlc.embed(rennen), meldung.*, verein.name, verein.kuerzel, verein.kurzform,
(SELECT COUNT(meldung.uuid) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid AND meldung.abgemeldet = false) as num_meldungen,
(SELECT COALESCE(MAX(meldung.abteilung),0) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid) as num_abteilungen
FROM rennen
FULL JOIN meldung
ON rennen.uuid = meldung.rennen_uuid
FULL JOIN verein
ON meldung.verein_uuid = verein.uuid
WHERE wettkampf = ANY($1::wettkampf[])
ORDER BY rennen.sort_id;

-- name: GetAllRennenWithAthlet :many
SELECT sqlc.embed(rennen), sqlc.embed(meldung), sqlc.embed(athlet), sqlc.embed(verein), link_meldung_athlet.position, link_meldung_athlet.rolle,
(SELECT COUNT(meldung.uuid) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid AND meldung.abgemeldet = false) as num_meldungen,
(SELECT COALESCE(MAX(meldung.abteilung),0) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid) as num_abteilungen
FROM rennen
JOIN meldung
ON rennen.uuid = meldung.rennen_uuid
JOIN verein
ON meldung.verein_uuid = verein.uuid
JOIN link_meldung_athlet
ON meldung.uuid = link_meldung_athlet.meldung_uuid
JOIN athlet
ON link_meldung_athlet.athlet_uuid = athlet.uuid
WHERE wettkampf = ANY($1::wettkampf[])
ORDER BY rennen.sort_id, meldung.uuid, link_meldung_athlet.rolle, link_meldung_athlet.position;

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
