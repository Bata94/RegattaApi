-- name: CreateLinkMeldungAthlet :one
INSERT INTO link_meldung_athlet (
  athlet_uuid,
  meldung_uuid,
  position
) VALUES (
  $1,
  $2,
  $3
) RETURNING *;
