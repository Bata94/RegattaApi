-- name: CreateRechnung :exec
INSERT INTO rechnung(
  nummer,
  verein_uuid,
  cost_sum
)VALUES(
  $1,
  $2,
  $3
);
