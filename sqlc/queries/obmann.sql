-- name: GetObmannMinimal :one
SELECT * FROM obmann
WHERE uuid = $1 LIMIT 1;

-- name: GetAllObmannForVerein :many
SELECT * FROM obmann
WHERE verein_uuid = $1
ORDER BY name ASC;

-- name: GetAllObmann :many
SELECT * FROM obmann
ORDER BY name ASC;

-- name: CreateObmann :one
INSERT INTO obmann (
  uuid,
  verein_uuid,
  name,
  email,
  phone
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateObmann :one
UPDATE obmann
SET
  name = $2,
  email = $3,
  phone = $4,
  verein_uuid = $5
WHERE uuid = $1
RETURNING *;

-- name: DeleteObmann :exec
DELETE FROM obmann
WHERE uuid = $1;
