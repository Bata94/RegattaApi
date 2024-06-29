-- name: GetUserMinimal :one
SELECT * FROM users
WHERE ulid = $1 LIMIT 1;

-- name: GetUser :one
SELECT sqlc.embed(users), sqlc.embed(users_group)
FROM users
JOIN users_group
ON users.group_ulid = users_group.ulid
WHERE users.ulid = $1 LIMIT 1;

-- name: GetUserUlidByName :one
SELECT ulid 
FROM users
WHERE username = $1;

-- name: GetAllUser :many
SELECT * FROM users
ORDER BY ulid;

-- name: CreateUser :one
INSERT INTO users (
  group_ulid,
  username,
  hashed_password
) VALUES (
  $1, $2, $3
)
RETURNING *;
