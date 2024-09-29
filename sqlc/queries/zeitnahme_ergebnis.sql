-- name: GetZeitnahmeErgebnisMinimal :one
SELECT * FROM zeitnahme_ergebnis
WHERE id = $1 LIMIT 1;

-- name: GetZeitnahmeErgebnisByMeld :one
SELECT * FROM zeitnahme_ergebnis
WHERE meldung_uuid = $1 LIMIT 1;

-- name: GetAllZeitnahmeErgebnis :many
SELECT * FROM zeitnahme_ergebnis
ORDER BY id ASC;

-- name: CreateZeitnahmeErgebnis :one
INSERT INTO zeitnahme_ergebnis (
  endzeit,
  zeitnahme_start_id,
  zeitnahme_ziel_id,
  meldung_uuid
) VALUES (
  $1,
  $2,
  $3,
  $4
) RETURNING *;
