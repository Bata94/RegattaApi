-- +goose Up
ALTER TABLE zeitnahme_start ADD COLUMN client_id text;
ALTER TABLE zeitnahme_start ADD COLUMN seq text;
ALTER TABLE zeitnahme_ziel ADD COLUMN client_id text;
ALTER TABLE zeitnahme_ziel ADD COLUMN seq text;

CREATE UNIQUE INDEX IF NOT EXISTS uq_zeitnahme_start_client_seq ON zeitnahme_start(client_id, seq) WHERE client_id IS NOT NULL AND seq IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_zeitnahme_ziel_client_seq ON zeitnahme_ziel(client_id, seq) WHERE client_id IS NOT NULL AND seq IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_zeitnahme_start_client_seq;
DROP INDEX IF EXISTS uq_zeitnahme_ziel_client_seq;
ALTER TABLE zeitnahme_start DROP COLUMN IF EXISTS client_id;
ALTER TABLE zeitnahme_start DROP COLUMN IF EXISTS seq;
ALTER TABLE zeitnahme_ziel DROP COLUMN IF EXISTS client_id;
ALTER TABLE zeitnahme_ziel DROP COLUMN IF EXISTS seq;
