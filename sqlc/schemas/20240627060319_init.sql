-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE EXTENSION ulid;

CREATE TYPE geschlecht AS ENUM ('m', 'w', 'x');
CREATE TYPE tag AS ENUM('sa', 'so');
CREATE TYPE wettkampf AS ENUM('Langstrecke', 'Kurzstrecke', 'Slalom', 'Staffel');
CREATE TYPE ROLLE AS ENUM('Ruderer', 'Stm.', 'Trainer');

CREATE TABLE verein (
  uuid uuid PRIMARY KEY,
  name text,
  kurzform text,
  kuerzel text
);

CREATE TABLE athlet (
  uuid uuid PRIMARY KEY,
  vorname text NOT NULL,
  name text NOT NULL,
  geschlecht geschlecht NOT NULL,
  jahrgang text NOT NULL,
  gewicht int DEFAULT 0,
  aerztliche_bescheinigung boolean DEFAULT false,
  lg_gemeldet boolean DEFAULT false,
  verein_uuid uuid NOT NULL,
  CONSTRAINT fk_verein FOREIGN KEY (verein_uuid) REFERENCES verein(uuid)
);


CREATE TABLE zeitnahme_start (
  id SERIAL PRIMARY KEY,
  rennen_nummer text,
  start_nummer text,
  time_client timestamp,
  time_server timestamp,
  measured_latency int,
  verarbeitet boolean DEFAULT false
);

CREATE TABLE zeitnahme_ziel (
  id SERIAL PRIMARY KEY,
  rennen_nummer text,
  start_nummer text,
  time_client timestamp,
  time_server timestamp,
  measured_latency int,
  verarbeitet boolean DEFAULT false
);

CREATE TABLE rennen (
  uuid uuid PRIMARY KEY,
  sort_id int unique NOT NULL,
  nummer text NOT NULL,
  bezeichnung text,
  bezeichnung_lang text,
  zusatz text,
  leichtgewicht boolean DEFAULT false,
  geschlecht geschlecht,
  bootsklasse text,
  bootsklasse_lang text,
  altersklasse text,
  altersklasse_lang text,
  tag tag NOT NULL,
  wettkampf wettkampf NOT NULL,
  kosten_eur int,
  rennabstand int,
  startzeit text DEFAULT '00:00'
);

CREATE TABLE pausen (
  id SERIAL PRIMARY KEY,
  tag tag NOT NULL,
  laenge int NOT NULL,
  nach_rennen_nummer text NOT NULL
);

CREATE TABLE meldung(
  uuid uuid PRIMARY KEY,
  drv_revision_uuid uuid NOT NULL,
  typ text NOT NULL,
  bemerkung text,
  abgemeldet boolean DEFAULT false,
  dns boolean DEFAULT false,
  dnf boolean DEFAULT false,
  dsq boolean DEFAULT false,
  zeitnahme_bemerkung text,
  start_nummer int DEFAULT 0,
  abteilung int DEFAULT 0,
  bahn int DEFAULT 0,
  kosten int NOT NULL,
  verein_uuid uuid NOT NULL,
  CONSTRAINT fk_verein FOREIGN KEY (verein_uuid) REFERENCES verein(uuid),
  rennen_uuid uuid NOT NULL,
  CONSTRAINT fk_rennen FOREIGN KEY (rennen_uuid) REFERENCES rennen(uuid)
);

CREATE TABLE link_meldung_athlet (
  id SERIAL PRIMARY KEY,
  rolle rolle NOT NULL,
  position int,
  meldung_uuid uuid NOT NULL,
  CONSTRAINT fk_meldung FOREIGN KEY (meldung_uuid) REFERENCES meldung(uuid),
  athlet_uuid uuid NOT NULL,
  CONSTRAINT fk_athlet FOREIGN KEY (athlet_uuid) REFERENCES athlet(uuid)
);

CREATE TABLE obmann (
  uuid uuid PRIMARY KEY,
  name text,
  email text,
  phone text,
  verein_uuid uuid NOT NULL,
  CONSTRAINT fk_verein FOREIGN KEY (verein_uuid) REFERENCES verein(uuid)
);

CREATE TABLE zeitnahme_ergebnis (
  id SERIAL PRIMARY KEY,
  endzeit float,
  zeitnahme_start_id int NOT NULL,
  CONSTRAINT fk_zeitnahme_start FOREIGN KEY (zeitnahme_start_id) REFERENCES zeitnahme_start(id),
  zeitnahme_ziel_id int NOT NULL,
  CONSTRAINT fk_zeitnahme_ziel FOREIGN KEY (zeitnahme_ziel_id) REFERENCES zeitnahme_ziel(id),
  meldung_uuid uuid NOT NULL,
  CONSTRAINT fk_meldung FOREIGN KEY (meldung_uuid) REFERENCES meldung(uuid)
);

CREATE TABLE users_group (
  ulid ulid PRIMARY KEY DEFAULT gen_monotonic_ulid(),
  name text,
  allowed_admin boolean DEFAULT false,
  allowed_zeitnahme boolean DEFAULT false,
  allowed_startlisten boolean DEFAULT false,
  allowed_regattaleitung boolean DEFAULT false
);

CREATE TABLE users (
  ulid ulid PRIMARY KEY DEFAULT gen_monotonic_ulid(),
  username text UNIQUE,
  hashed_password text,
  is_active boolean DEFAULT false,
  group_ulid ulid NOT NULL,
  CONSTRAINT fk_users_group FOREIGN KEY (group_ulid) REFERENCES users_group(ulid)
);

INSERT INTO pausen (tag, laenge, nach_rennen_nummer) VALUES ('so', 30, '115');
INSERT INTO pausen (tag, laenge, nach_rennen_nummer) VALUES ('so', 30, '130');
INSERT INTO pausen (tag, laenge, nach_rennen_nummer) VALUES ('so', 30, '321');

INSERT INTO users_group (
  ulid,
  name,
  allowed_admin,
  allowed_zeitnahme,
  allowed_startlisten,
  allowed_regattaleitung
) VALUES (
  '01J1HJBTAXD1T2DYVJ6SASKGGV',
  'full_admin',
  true,
  true,
  true,
  true
);
INSERT INTO users_group (
  ulid,
  name,
  allowed_regattaleitung
) VALUES (
  '01J1HJBTAWCF0DVYQ0AFJ8GH9P',
  'regattaleitung',
  true
);
INSERT INTO users_group (
  ulid,
  name,
  allowed_zeitnahme
) VALUES (
  '01J1HJBTAXMGNP5R6PR0WJ0GG1',
  'zeitnahme',
  true
);
INSERT INTO users_group (
  ulid,
  name,
  allowed_startlisten
) VALUES (
  '01J1HJBTAXP10XTYV4SW3D65TV',
  'startlisten',
  true
);

INSERT INTO users (
  username,
  hashed_password,
  group_ulid
) VALUES (
  'admin',
  '$2a$14$HKUH7lzr8gf.rKE/.k2mEessP1cgFLvWrKQ18pg2Bi8QBbwjzkWBu',
  '01J1HJBTAXD1T2DYVJ6SASKGGV'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE users;
DROP TABLE users_group;
DROP TABLE zeitnahme_ergebnis;
DROP TABLE obmann;
DROP TABLE link_meldung_athlet;
DROP TABLE meldung;
DROP TABLE rennen;
DROP TABLE pausen;
DROP TABLE zeitnahme_ziel;
DROP TABLE zeitnahme_start;
DROP TABLE athlet;
DROP TABLE verein;

DROP TYPE tag;
DROP TYPE wettkampf;
DROP TYPE geschlecht;
DROP TYPE rolle;

DROP EXTENSION ulid;
-- +goose StatementEnd
