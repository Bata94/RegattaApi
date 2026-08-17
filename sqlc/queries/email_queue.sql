-- name: CreateEmailQueueEntry :one
INSERT INTO email_queue (
  to_addresses,
  cc_addresses,
  bcc_addresses,
  subject,
  body,
  attachments,
  max_attempts
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
) RETURNING *;

-- name: ClaimNextEmailQueueEntry :one
UPDATE email_queue
SET status = 'sending', updated_at = now()
WHERE uuid = (
  SELECT uuid FROM email_queue
  WHERE status IN ('pending', 'failed')
    AND next_attempt_at <= now()
    AND attempts < max_attempts
  ORDER BY created_at ASC
  FOR UPDATE SKIP LOCKED
  LIMIT 1
) RETURNING *;

-- name: MarkEmailQueueSent :exec
UPDATE email_queue
SET status = 'sent', sent_at = now(), updated_at = now(), last_error = NULL
WHERE uuid = $1;

-- name: MarkEmailQueueFailed :exec
UPDATE email_queue
SET status = CASE WHEN attempts + 1 >= max_attempts THEN 'dead' ELSE 'failed' END,
    attempts = attempts + 1,
    next_attempt_at = now() + make_interval(secs => $2),
    last_error = $3,
    updated_at = now()
WHERE uuid = $1;

-- name: GetEmailQueueEntries :many
SELECT * FROM email_queue
ORDER BY created_at DESC
LIMIT 200;

-- name: ResetEmailQueueEntry :exec
UPDATE email_queue
SET status = 'pending', attempts = 0, next_attempt_at = now(),
    last_error = NULL, updated_at = now()
WHERE uuid = $1;

-- name: DeleteEmailQueueEntry :exec
DELETE FROM email_queue
WHERE uuid = $1;
