-- name: GetStartnummernBereich :one
SELECT * FROM startnummern_bereich
WHERE id = 1 LIMIT 1;

-- name: SetStartnummernBereich :one
INSERT INTO startnummern_bereich (
  id,
  min_nummer,
  max_nummer,
  fehlende_nummern
) VALUES (
  1, $1, $2, $3
)
ON CONFLICT (id) DO UPDATE
SET
  min_nummer = EXCLUDED.min_nummer,
  max_nummer = EXCLUDED.max_nummer,
  fehlende_nummern = EXCLUDED.fehlende_nummern
RETURNING *;
