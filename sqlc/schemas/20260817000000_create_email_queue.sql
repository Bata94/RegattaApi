-- +goose Up
CREATE TABLE email_queue (
  uuid            uuid PRIMARY KEY DEFAULT uuidv7(),
  to_addresses    text[] NOT NULL,
  cc_addresses    text[] NOT NULL DEFAULT '{}',
  bcc_addresses   text[] NOT NULL DEFAULT '{}',
  subject         text NOT NULL,
  body            text NOT NULL,
  attachments     text[] NOT NULL DEFAULT '{}',
  status          text NOT NULL DEFAULT 'pending',
  attempts        int NOT NULL DEFAULT 0,
  max_attempts    int NOT NULL DEFAULT 5,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  last_error      text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  sent_at         timestamptz
);

CREATE INDEX idx_email_queue_due ON email_queue (status, next_attempt_at)
  WHERE status IN ('pending', 'failed');

-- +goose Down
DROP TABLE email_queue;
