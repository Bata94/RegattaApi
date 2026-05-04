-- name: GetUserMinimal :one
SELECT * FROM users
WHERE uuid = $1 LIMIT 1;

-- name: GetUser :one
SELECT sqlc.embed(users), sqlc.embed(users_group)
FROM users
JOIN users_group
ON users.group_uuid = users_group.uuid
WHERE users.uuid = $1 LIMIT 1;

-- name: GetUserUuidByName :one
SELECT uuid
FROM users
WHERE username = $1;

-- name: GetAllUser :many
SELECT * FROM users
ORDER BY uuid;

-- name: CreateUser :one
INSERT INTO users (
  group_uuid,
  username,
  hashed_password,
  uuid
) VALUES (
  $1, $2, $3, gen_random_uuid()
)
RETURNING *;
