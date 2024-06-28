-- name: GetObmannMinimal :one
SELECT * FROM obmann
WHERE uuid = $1 LIMIT 1;

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
