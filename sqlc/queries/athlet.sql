-- name: GetAthletMinimal :one
SELECT * FROM athlet
WHERE uuid = $1 LIMIT 1;

-- name: GetAllAthlet :many
SELECT * FROM athlet
ORDER BY name ASC;

-- name: GetAllNNAthleten :many
SELECT * FROM athlet
WHERE vorname = 'No' and name = 'Name' and jahrgang = '9999'
ORDER BY verein_uuid ASC;

-- name: GetAllAthletenForVerein :many
SELECT DISTINCT
  athlet.*
FROM
  athlet
WHERE
  athlet.verein_uuid = $1
ORDER BY
  athlet.name, athlet.vorname;

-- name: GetAllAthletenForVereinMissStartber :many
SELECT DISTINCT
  sqlc.embed(athlet),
  sqlc.embed(rennen)
FROM
  athlet
JOIN
  link_meldung_athlet
ON
  athlet.uuid = link_meldung_athlet.athlet_uuid
JOIN
  meldung
ON
  link_meldung_athlet.meldung_uuid = meldung.uuid
JOIN
  rennen
ON
  meldung.rennen_uuid = rennen.uuid
WHERE
  meldung.verein_uuid = $1 AND
  meldung.abgemeldet = false AND
  athlet.startberechtigt = false
ORDER BY
  athlet.name, athlet.vorname;

-- name: GetAllAthletenForVereinWaage :many
SELECT DISTINCT
  sqlc.embed(athlet),
  sqlc.embed(rennen)
FROM
  athlet
JOIN
  link_meldung_athlet
ON
  athlet.uuid = link_meldung_athlet.athlet_uuid
JOIN
  meldung
ON
  link_meldung_athlet.meldung_uuid = meldung.uuid
JOIN
  rennen
ON
  meldung.rennen_uuid = rennen.uuid
WHERE
  meldung.verein_uuid = $1 AND
  rennen.leichtgewicht = true AND
  meldung.abgemeldet = false AND
  link_meldung_athlet.rolle = 'Ruderer' AND
  athlet.gewicht = 0
ORDER BY
  athlet.name, athlet.vorname, rennen.sort_id;

-- name: CreateAthlet :one
INSERT INTO athlet
(uuid, verein_uuid, name, vorname, jahrgang, startberechtigt, geschlecht)
VALUES
($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateAthletAerztlBesch :exec
UPDATE
  athlet
SET
  startberechtigt = $1
WHERE
  uuid = $2;

-- name: UpdateAthletWaage :exec
UPDATE
  athlet
SET
  gewicht = $1
WHERE
  uuid = $2;
