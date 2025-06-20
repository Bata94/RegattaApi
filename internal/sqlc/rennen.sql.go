// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: rennen.sql

package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createRennen = `-- name: CreateRennen :one
INSERT INTO rennen (
  uuid,
  sort_id,
  nummer,
  bezeichnung,
  bezeichnung_lang,
  zusatz,
  leichtgewicht,
  geschlecht,
  bootsklasse,
  bootsklasse_lang,
  altersklasse,
  altersklasse_lang,
  tag,
  wettkampf,
  kosten_eur,
  rennabstand
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING uuid, sort_id, nummer, bezeichnung, bezeichnung_lang, zusatz, leichtgewicht, geschlecht, bootsklasse, bootsklasse_lang, altersklasse, altersklasse_lang, tag, wettkampf, kosten_eur, rennabstand, startzeit
`

type CreateRennenParams struct {
	Uuid             uuid.UUID   `json:"uuid"`
	SortID           int32       `json:"sort_id"`
	Nummer           string      `json:"nummer"`
	Bezeichnung      string      `json:"bezeichnung"`
	BezeichnungLang  string      `json:"bezeichnung_lang"`
	Zusatz           pgtype.Text `json:"zusatz"`
	Leichtgewicht    bool        `json:"leichtgewicht"`
	Geschlecht       Geschlecht  `json:"geschlecht"`
	Bootsklasse      string      `json:"bootsklasse"`
	BootsklasseLang  string      `json:"bootsklasse_lang"`
	Altersklasse     string      `json:"altersklasse"`
	AltersklasseLang string      `json:"altersklasse_lang"`
	Tag              Tag         `json:"tag"`
	Wettkampf        Wettkampf   `json:"wettkampf"`
	KostenEur        pgtype.Int4 `json:"kosten_eur"`
	Rennabstand      pgtype.Int4 `json:"rennabstand"`
}

func (q *Queries) CreateRennen(ctx context.Context, arg CreateRennenParams) (Rennen, error) {
	row := q.db.QueryRow(ctx, createRennen,
		arg.Uuid,
		arg.SortID,
		arg.Nummer,
		arg.Bezeichnung,
		arg.BezeichnungLang,
		arg.Zusatz,
		arg.Leichtgewicht,
		arg.Geschlecht,
		arg.Bootsklasse,
		arg.BootsklasseLang,
		arg.Altersklasse,
		arg.AltersklasseLang,
		arg.Tag,
		arg.Wettkampf,
		arg.KostenEur,
		arg.Rennabstand,
	)
	var i Rennen
	err := row.Scan(
		&i.Uuid,
		&i.SortID,
		&i.Nummer,
		&i.Bezeichnung,
		&i.BezeichnungLang,
		&i.Zusatz,
		&i.Leichtgewicht,
		&i.Geschlecht,
		&i.Bootsklasse,
		&i.BootsklasseLang,
		&i.Altersklasse,
		&i.AltersklasseLang,
		&i.Tag,
		&i.Wettkampf,
		&i.KostenEur,
		&i.Rennabstand,
		&i.Startzeit,
	)
	return i, err
}

const getAllRennen = `-- name: GetAllRennen :many
SELECT
  rennen.uuid, rennen.sort_id, rennen.nummer, rennen.bezeichnung, rennen.bezeichnung_lang, rennen.zusatz, rennen.leichtgewicht, rennen.geschlecht, rennen.bootsklasse, rennen.bootsklasse_lang, rennen.altersklasse, rennen.altersklasse_lang, rennen.tag, rennen.wettkampf, rennen.kosten_eur, rennen.rennabstand, rennen.startzeit,
  (SELECT COUNT(meldung.uuid) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid AND meldung.abgemeldet = false) as num_meldungen,
  (SELECT COALESCE(MAX(meldung.abteilung),0) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid) as num_abteilungen
FROM
  rennen
ORDER BY
  sort_id ASC
`

type GetAllRennenRow struct {
	Rennen         Rennen      `json:"rennen"`
	NumMeldungen   int64       `json:"num_meldungen"`
	NumAbteilungen interface{} `json:"num_abteilungen"`
}

func (q *Queries) GetAllRennen(ctx context.Context) ([]GetAllRennenRow, error) {
	rows, err := q.db.Query(ctx, getAllRennen)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllRennenRow{}
	for rows.Next() {
		var i GetAllRennenRow
		if err := rows.Scan(
			&i.Rennen.Uuid,
			&i.Rennen.SortID,
			&i.Rennen.Nummer,
			&i.Rennen.Bezeichnung,
			&i.Rennen.BezeichnungLang,
			&i.Rennen.Zusatz,
			&i.Rennen.Leichtgewicht,
			&i.Rennen.Geschlecht,
			&i.Rennen.Bootsklasse,
			&i.Rennen.BootsklasseLang,
			&i.Rennen.Altersklasse,
			&i.Rennen.AltersklasseLang,
			&i.Rennen.Tag,
			&i.Rennen.Wettkampf,
			&i.Rennen.KostenEur,
			&i.Rennen.Rennabstand,
			&i.Rennen.Startzeit,
			&i.NumMeldungen,
			&i.NumAbteilungen,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllRennenWithAthlet = `-- name: GetAllRennenWithAthlet :many
SELECT 
  rennen.uuid, rennen.sort_id, rennen.nummer, rennen.bezeichnung, rennen.bezeichnung_lang, rennen.zusatz, rennen.leichtgewicht, rennen.geschlecht, rennen.bootsklasse, rennen.bootsklasse_lang, rennen.altersklasse, rennen.altersklasse_lang, rennen.tag, rennen.wettkampf, rennen.kosten_eur, rennen.rennabstand, rennen.startzeit,
  meldung.uuid, meldung.drv_revision_uuid, meldung.typ, meldung.bemerkung, meldung.abgemeldet, meldung.dns, meldung.dnf, meldung.dsq, meldung.zeitnahme_bemerkung, meldung.start_nummer, meldung.abteilung, meldung.bahn, meldung.kosten, meldung.rechnungs_nummer, meldung.verein_uuid, meldung.rennen_uuid,
  athlet.uuid, athlet.vorname, athlet.name, athlet.geschlecht, athlet.jahrgang, athlet.gewicht, athlet.startberechtigt, athlet.verein_uuid,
  verein.uuid, verein.name, verein.kurzform, verein.kuerzel,
  link_meldung_athlet.position, link_meldung_athlet.rolle,
  (SELECT COUNT(meldung.uuid) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid AND meldung.abgemeldet = false) as num_meldungen,
  (SELECT COALESCE(MAX(meldung.abteilung),0) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid) as num_abteilungen
FROM 
  rennen
JOIN 
  meldung
ON 
  rennen.uuid = meldung.rennen_uuid
JOIN 
  verein
ON 
  meldung.verein_uuid = verein.uuid
JOIN 
  link_meldung_athlet
ON 
  meldung.uuid = link_meldung_athlet.meldung_uuid
JOIN 
  athlet
ON 
  link_meldung_athlet.athlet_uuid = athlet.uuid
WHERE 
  wettkampf = ANY($1::wettkampf[])
ORDER BY 
  rennen.sort_id, meldung.uuid, link_meldung_athlet.rolle, link_meldung_athlet.position
`

type GetAllRennenWithAthletRow struct {
	Rennen         Rennen      `json:"rennen"`
	Meldung        Meldung     `json:"meldung"`
	Athlet         Athlet      `json:"athlet"`
	Verein         Verein      `json:"verein"`
	Position       int32       `json:"position"`
	Rolle          Rolle       `json:"rolle"`
	NumMeldungen   int64       `json:"num_meldungen"`
	NumAbteilungen interface{} `json:"num_abteilungen"`
}

func (q *Queries) GetAllRennenWithAthlet(ctx context.Context, dollar_1 []Wettkampf) ([]GetAllRennenWithAthletRow, error) {
	rows, err := q.db.Query(ctx, getAllRennenWithAthlet, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllRennenWithAthletRow{}
	for rows.Next() {
		var i GetAllRennenWithAthletRow
		if err := rows.Scan(
			&i.Rennen.Uuid,
			&i.Rennen.SortID,
			&i.Rennen.Nummer,
			&i.Rennen.Bezeichnung,
			&i.Rennen.BezeichnungLang,
			&i.Rennen.Zusatz,
			&i.Rennen.Leichtgewicht,
			&i.Rennen.Geschlecht,
			&i.Rennen.Bootsklasse,
			&i.Rennen.BootsklasseLang,
			&i.Rennen.Altersklasse,
			&i.Rennen.AltersklasseLang,
			&i.Rennen.Tag,
			&i.Rennen.Wettkampf,
			&i.Rennen.KostenEur,
			&i.Rennen.Rennabstand,
			&i.Rennen.Startzeit,
			&i.Meldung.Uuid,
			&i.Meldung.DrvRevisionUuid,
			&i.Meldung.Typ,
			&i.Meldung.Bemerkung,
			&i.Meldung.Abgemeldet,
			&i.Meldung.Dns,
			&i.Meldung.Dnf,
			&i.Meldung.Dsq,
			&i.Meldung.ZeitnahmeBemerkung,
			&i.Meldung.StartNummer,
			&i.Meldung.Abteilung,
			&i.Meldung.Bahn,
			&i.Meldung.Kosten,
			&i.Meldung.RechnungsNummer,
			&i.Meldung.VereinUuid,
			&i.Meldung.RennenUuid,
			&i.Athlet.Uuid,
			&i.Athlet.Vorname,
			&i.Athlet.Name,
			&i.Athlet.Geschlecht,
			&i.Athlet.Jahrgang,
			&i.Athlet.Gewicht,
			&i.Athlet.Startberechtigt,
			&i.Athlet.VereinUuid,
			&i.Verein.Uuid,
			&i.Verein.Name,
			&i.Verein.Kurzform,
			&i.Verein.Kuerzel,
			&i.Position,
			&i.Rolle,
			&i.NumMeldungen,
			&i.NumAbteilungen,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllRennenWithMeld = `-- name: GetAllRennenWithMeld :many
SELECT
  rennen.uuid, rennen.sort_id, rennen.nummer, rennen.bezeichnung, rennen.bezeichnung_lang, rennen.zusatz, rennen.leichtgewicht, rennen.geschlecht, rennen.bootsklasse, rennen.bootsklasse_lang, rennen.altersklasse, rennen.altersklasse_lang, rennen.tag, rennen.wettkampf, rennen.kosten_eur, rennen.rennabstand, rennen.startzeit,
  meldung.uuid, meldung.drv_revision_uuid, meldung.typ, meldung.bemerkung, meldung.abgemeldet, meldung.dns, meldung.dnf, meldung.dsq, meldung.zeitnahme_bemerkung, meldung.start_nummer, meldung.abteilung, meldung.bahn, meldung.kosten, meldung.rechnungs_nummer, meldung.verein_uuid, meldung.rennen_uuid,
  verein.name, verein.kuerzel, verein.kurzform,
  (SELECT COUNT(meldung.uuid) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid AND meldung.abgemeldet = false) as num_meldungen,
  (SELECT COALESCE(MAX(meldung.abteilung),0) FROM meldung WHERE rennen.uuid = meldung.rennen_uuid) as num_abteilungen
FROM
  rennen
FULL
  JOIN meldung
ON
  rennen.uuid = meldung.rennen_uuid
FULL JOIN
  verein
ON
  meldung.verein_uuid = verein.uuid
WHERE
  wettkampf = ANY($1::wettkampf[])
ORDER BY
  rennen.sort_id
`

type GetAllRennenWithMeldRow struct {
	Rennen             Rennen      `json:"rennen"`
	Uuid               uuid.UUID   `json:"uuid"`
	DrvRevisionUuid    uuid.UUID   `json:"drv_revision_uuid"`
	Typ                pgtype.Text `json:"typ"`
	Bemerkung          pgtype.Text `json:"bemerkung"`
	Abgemeldet         pgtype.Bool `json:"abgemeldet"`
	Dns                pgtype.Bool `json:"dns"`
	Dnf                pgtype.Bool `json:"dnf"`
	Dsq                pgtype.Bool `json:"dsq"`
	ZeitnahmeBemerkung pgtype.Text `json:"zeitnahme_bemerkung"`
	StartNummer        pgtype.Int4 `json:"start_nummer"`
	Abteilung          pgtype.Int4 `json:"abteilung"`
	Bahn               pgtype.Int4 `json:"bahn"`
	Kosten             pgtype.Int4 `json:"kosten"`
	RechnungsNummer    pgtype.Text `json:"rechnungs_nummer"`
	VereinUuid         uuid.UUID   `json:"verein_uuid"`
	RennenUuid         uuid.UUID   `json:"rennen_uuid"`
	Name               pgtype.Text `json:"name"`
	Kuerzel            pgtype.Text `json:"kuerzel"`
	Kurzform           pgtype.Text `json:"kurzform"`
	NumMeldungen       int64       `json:"num_meldungen"`
	NumAbteilungen     interface{} `json:"num_abteilungen"`
}

func (q *Queries) GetAllRennenWithMeld(ctx context.Context, dollar_1 []Wettkampf) ([]GetAllRennenWithMeldRow, error) {
	rows, err := q.db.Query(ctx, getAllRennenWithMeld, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllRennenWithMeldRow{}
	for rows.Next() {
		var i GetAllRennenWithMeldRow
		if err := rows.Scan(
			&i.Rennen.Uuid,
			&i.Rennen.SortID,
			&i.Rennen.Nummer,
			&i.Rennen.Bezeichnung,
			&i.Rennen.BezeichnungLang,
			&i.Rennen.Zusatz,
			&i.Rennen.Leichtgewicht,
			&i.Rennen.Geschlecht,
			&i.Rennen.Bootsklasse,
			&i.Rennen.BootsklasseLang,
			&i.Rennen.Altersklasse,
			&i.Rennen.AltersklasseLang,
			&i.Rennen.Tag,
			&i.Rennen.Wettkampf,
			&i.Rennen.KostenEur,
			&i.Rennen.Rennabstand,
			&i.Rennen.Startzeit,
			&i.Uuid,
			&i.DrvRevisionUuid,
			&i.Typ,
			&i.Bemerkung,
			&i.Abgemeldet,
			&i.Dns,
			&i.Dnf,
			&i.Dsq,
			&i.ZeitnahmeBemerkung,
			&i.StartNummer,
			&i.Abteilung,
			&i.Bahn,
			&i.Kosten,
			&i.RechnungsNummer,
			&i.VereinUuid,
			&i.RennenUuid,
			&i.Name,
			&i.Kuerzel,
			&i.Kurzform,
			&i.NumMeldungen,
			&i.NumAbteilungen,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRennen = `-- name: GetRennen :many
SELECT 
  rennen.uuid, rennen.sort_id, rennen.nummer, rennen.bezeichnung, rennen.bezeichnung_lang, rennen.zusatz, rennen.leichtgewicht, rennen.geschlecht, rennen.bootsklasse, rennen.bootsklasse_lang, rennen.altersklasse, rennen.altersklasse_lang, rennen.tag, rennen.wettkampf, rennen.kosten_eur, rennen.rennabstand, rennen.startzeit,
  meldung.uuid, meldung.drv_revision_uuid, meldung.typ, meldung.bemerkung, meldung.abgemeldet, meldung.dns, meldung.dnf, meldung.dsq, meldung.zeitnahme_bemerkung, meldung.start_nummer, meldung.abteilung, meldung.bahn, meldung.kosten, meldung.rechnungs_nummer, meldung.verein_uuid, meldung.rennen_uuid,
  athlet.uuid, athlet.vorname, athlet.name, athlet.geschlecht, athlet.jahrgang, athlet.gewicht, athlet.startberechtigt, athlet.verein_uuid,
  verein.uuid, verein.name, verein.kurzform, verein.kuerzel,
  link_meldung_athlet.id, link_meldung_athlet.rolle, link_meldung_athlet.position, link_meldung_athlet.meldung_uuid, link_meldung_athlet.athlet_uuid
FROM
  rennen
FULL JOIN
  meldung
ON
  rennen.uuid = meldung.rennen_uuid
FULL JOIN
  link_meldung_athlet 
ON
  meldung.uuid = link_meldung_athlet.meldung_uuid
FULL JOIN
  athlet
ON
  link_meldung_athlet.athlet_uuid = athlet.uuid
FULL JOIN
  verein
ON
  meldung.verein_uuid = verein.uuid
WHERE
  rennen.uuid = $1
ORDER BY
  meldung.abteilung, meldung.bahn, link_meldung_athlet.rolle, link_meldung_athlet.position
`

type GetRennenRow struct {
	Rennen            Rennen            `json:"rennen"`
	Meldung           Meldung           `json:"meldung"`
	Athlet            Athlet            `json:"athlet"`
	Verein            Verein            `json:"verein"`
	LinkMeldungAthlet LinkMeldungAthlet `json:"link_meldung_athlet"`
}

func (q *Queries) GetRennen(ctx context.Context, argUuid uuid.UUID) ([]GetRennenRow, error) {
	rows, err := q.db.Query(ctx, getRennen, argUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetRennenRow{}
	for rows.Next() {
		var i GetRennenRow
		if err := rows.Scan(
			&i.Rennen.Uuid,
			&i.Rennen.SortID,
			&i.Rennen.Nummer,
			&i.Rennen.Bezeichnung,
			&i.Rennen.BezeichnungLang,
			&i.Rennen.Zusatz,
			&i.Rennen.Leichtgewicht,
			&i.Rennen.Geschlecht,
			&i.Rennen.Bootsklasse,
			&i.Rennen.BootsklasseLang,
			&i.Rennen.Altersklasse,
			&i.Rennen.AltersklasseLang,
			&i.Rennen.Tag,
			&i.Rennen.Wettkampf,
			&i.Rennen.KostenEur,
			&i.Rennen.Rennabstand,
			&i.Rennen.Startzeit,
			&i.Meldung.Uuid,
			&i.Meldung.DrvRevisionUuid,
			&i.Meldung.Typ,
			&i.Meldung.Bemerkung,
			&i.Meldung.Abgemeldet,
			&i.Meldung.Dns,
			&i.Meldung.Dnf,
			&i.Meldung.Dsq,
			&i.Meldung.ZeitnahmeBemerkung,
			&i.Meldung.StartNummer,
			&i.Meldung.Abteilung,
			&i.Meldung.Bahn,
			&i.Meldung.Kosten,
			&i.Meldung.RechnungsNummer,
			&i.Meldung.VereinUuid,
			&i.Meldung.RennenUuid,
			&i.Athlet.Uuid,
			&i.Athlet.Vorname,
			&i.Athlet.Name,
			&i.Athlet.Geschlecht,
			&i.Athlet.Jahrgang,
			&i.Athlet.Gewicht,
			&i.Athlet.Startberechtigt,
			&i.Athlet.VereinUuid,
			&i.Verein.Uuid,
			&i.Verein.Name,
			&i.Verein.Kurzform,
			&i.Verein.Kuerzel,
			&i.LinkMeldungAthlet.ID,
			&i.LinkMeldungAthlet.Rolle,
			&i.LinkMeldungAthlet.Position,
			&i.LinkMeldungAthlet.MeldungUuid,
			&i.LinkMeldungAthlet.AthletUuid,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getRennenMinimal = `-- name: GetRennenMinimal :one
SELECT uuid, sort_id, nummer, bezeichnung, bezeichnung_lang, zusatz, leichtgewicht, geschlecht, bootsklasse, bootsklasse_lang, altersklasse, altersklasse_lang, tag, wettkampf, kosten_eur, rennabstand, startzeit FROM rennen
WHERE uuid = $1 LIMIT 1
`

func (q *Queries) GetRennenMinimal(ctx context.Context, argUuid uuid.UUID) (Rennen, error) {
	row := q.db.QueryRow(ctx, getRennenMinimal, argUuid)
	var i Rennen
	err := row.Scan(
		&i.Uuid,
		&i.SortID,
		&i.Nummer,
		&i.Bezeichnung,
		&i.BezeichnungLang,
		&i.Zusatz,
		&i.Leichtgewicht,
		&i.Geschlecht,
		&i.Bootsklasse,
		&i.BootsklasseLang,
		&i.Altersklasse,
		&i.AltersklasseLang,
		&i.Tag,
		&i.Wettkampf,
		&i.KostenEur,
		&i.Rennabstand,
		&i.Startzeit,
	)
	return i, err
}

const updateStartZeit = `-- name: UpdateStartZeit :exec
UPDATE rennen SET startzeit = $1 WHERE uuid = $2
`

type UpdateStartZeitParams struct {
	Startzeit pgtype.Text `json:"startzeit"`
	Uuid      uuid.UUID   `json:"uuid"`
}

func (q *Queries) UpdateStartZeit(ctx context.Context, arg UpdateStartZeitParams) error {
	_, err := q.db.Exec(ctx, updateStartZeit, arg.Startzeit, arg.Uuid)
	return err
}
