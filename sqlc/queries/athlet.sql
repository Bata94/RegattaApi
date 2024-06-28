-- name: GetAthletMinimal :one
SELECT * FROM athlet
WHERE uuid = $1 LIMIT 1;

-- name: GetAllAthlet :many
SELECT * FROM athlet
ORDER BY name ASC;

-- name: CreateAthlet :one
INSERT INTO athlet
(uuid, verein_uuid, name, vorname, jahrgang, aerztliche_bescheinigung, geschlecht)
VALUES
($1, $2, $3, $4, $5, $6, $7)
RETURNING *;
