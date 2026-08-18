-- +goose Up
CREATE TABLE startnummern_bereich (
  id int PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  min_nummer int NOT NULL DEFAULT 1,
  max_nummer int NOT NULL DEFAULT 350,
  fehlende_nummern int[] NOT NULL DEFAULT '{}'
);

INSERT INTO startnummern_bereich (id) VALUES (1);

-- +goose Down
DROP TABLE startnummern_bereich;
