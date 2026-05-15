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
  allowed_admin,
  allowed_zeitnahme,
  allowed_startlisten,
  allowed_regattabuero,
  allowed_regattaleitung,
  uuid
) VALUES (
  $1, $2, $3, $4, $5, $6, uuidv7()
)
RETURNING *;

-- name: UpdateUserGroup :exec
UPDATE users_group
SET
  name = $2,
  allowed_admin = $3,
  allowed_zeitnahme = $4,
  allowed_startlisten = $5,
  allowed_regattabuero = $6,
  allowed_regattaleitung = $7
WHERE uuid = $1;
