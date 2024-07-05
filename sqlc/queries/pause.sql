-- name: GetAllPause :many
SELECT * FROM pause
ORDER BY id ASC;

-- name: GetPause :one
SELECT * FROM pause
WHERE id = $1 LIMIT 1;

-- name: CreatePause :one
INSERT INTO pause (
  laenge,
  nach_rennen_uuid
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: UpdatePause :one
UPDATE pause
SET laenge = $2
WHERE id = $1
RETURNING *;

-- name: DeletePause :exec
DELETE FROM pause
WHERE id = $1;
