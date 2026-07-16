-- name: GetUserGroupMinimal :one
SELECT *
FROM users_group
WHERE users_group.uuid = $1;

-- name: GetUserGroup :many
SELECT sqlc.embed(users_group), sqlc.embed(users)
FROM users_group
JOIN users
ON users_group.uuid = users.group_uuid
WHERE users_group.uuid = $1;

-- name: GetUserGroupUuidByName :one
SELECT uuid
FROM users_group
WHERE name = $1;

-- name: GetAllUserGroup :many
SELECT * FROM users_group
ORDER BY uuid;

-- name: CreateUserGroup :one
INSERT INTO users_group (
  name,
  capabilities,
  uuid
) VALUES (
  $1, $2, uuidv7()
)
RETURNING *;

-- name: UpdateUserGroup :exec
UPDATE users_group
SET
  name = $2,
  capabilities = $3
WHERE uuid = $1;
