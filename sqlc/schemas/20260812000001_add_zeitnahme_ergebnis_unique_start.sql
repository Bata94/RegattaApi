-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS uq_zeitnahme_ergebnis_start ON zeitnahme_ergebnis(zeitnahme_start_id) WHERE zeitnahme_start_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_zeitnahme_ergebnis_start;
