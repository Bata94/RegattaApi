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
