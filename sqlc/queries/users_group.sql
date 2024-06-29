-- name: GetUserGroupMinimal :one
SELECT *
FROM users_group
WHERE users_group.ulid = $1;

-- name: GetUserGroup :many
SELECT sqlc.embed(users_group), sqlc.embed(users)
FROM users_group
JOIN users
ON users_group.ulid = users.group_ulid
WHERE users_group.ulid = $1;

-- name: GetUserGroupUlidByName :one
SELECT ulid
FROM users_group
WHERE name = $1;

-- name: GetAllUserGroup :many
SELECT * FROM users_group
ORDER BY ulid;

-- name: CreateUserGroup :one
INSERT INTO users_group (
  name,
  allowed_admin
) VALUES (
  $1, $2
)
RETURNING *;
