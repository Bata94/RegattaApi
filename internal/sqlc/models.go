// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package sqlc

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Geschlecht string

const (
	GeschlechtM Geschlecht = "m"
	GeschlechtW Geschlecht = "w"
	GeschlechtX Geschlecht = "x"
)

func (e *Geschlecht) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Geschlecht(s)
	case string:
		*e = Geschlecht(s)
	default:
		return fmt.Errorf("unsupported scan type for Geschlecht: %T", src)
	}
	return nil
}

type NullGeschlecht struct {
	Geschlecht Geschlecht `json:"geschlecht"`
	Valid      bool       `json:"valid"` // Valid is true if Geschlecht is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullGeschlecht) Scan(value interface{}) error {
	if value == nil {
		ns.Geschlecht, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Geschlecht.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullGeschlecht) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Geschlecht), nil
}

type Rolle string

const (
	RolleRuderer Rolle = "Ruderer"
	RolleStm     Rolle = "Stm."
	RolleTrainer Rolle = "Trainer"
)

func (e *Rolle) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Rolle(s)
	case string:
		*e = Rolle(s)
	default:
		return fmt.Errorf("unsupported scan type for Rolle: %T", src)
	}
	return nil
}

type NullRolle struct {
	Rolle Rolle `json:"rolle"`
	Valid bool  `json:"valid"` // Valid is true if Rolle is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRolle) Scan(value interface{}) error {
	if value == nil {
		ns.Rolle, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Rolle.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRolle) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Rolle), nil
}

type Tag string

const (
	TagSa Tag = "sa"
	TagSo Tag = "so"
)

func (e *Tag) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Tag(s)
	case string:
		*e = Tag(s)
	default:
		return fmt.Errorf("unsupported scan type for Tag: %T", src)
	}
	return nil
}

type NullTag struct {
	Tag   Tag  `json:"tag"`
	Valid bool `json:"valid"` // Valid is true if Tag is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullTag) Scan(value interface{}) error {
	if value == nil {
		ns.Tag, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Tag.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullTag) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Tag), nil
}

type Wettkampf string

const (
	WettkampfLangstrecke Wettkampf = "Langstrecke"
	WettkampfKurzstrecke Wettkampf = "Kurzstrecke"
	WettkampfSlalom      Wettkampf = "Slalom"
	WettkampfStaffel     Wettkampf = "Staffel"
)

func (e *Wettkampf) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = Wettkampf(s)
	case string:
		*e = Wettkampf(s)
	default:
		return fmt.Errorf("unsupported scan type for Wettkampf: %T", src)
	}
	return nil
}

type NullWettkampf struct {
	Wettkampf Wettkampf `json:"wettkampf"`
	Valid     bool      `json:"valid"` // Valid is true if Wettkampf is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullWettkampf) Scan(value interface{}) error {
	if value == nil {
		ns.Wettkampf, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Wettkampf.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullWettkampf) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Wettkampf), nil
}

type Athlet struct {
	Uuid            uuid.UUID   `json:"uuid"`
	Vorname         string      `json:"vorname"`
	Name            string      `json:"name"`
	Geschlecht      Geschlecht  `json:"geschlecht"`
	Jahrgang        string      `json:"jahrgang"`
	Gewicht         pgtype.Int4 `json:"gewicht"`
	Startberechtigt bool        `json:"startberechtigt"`
	VereinUuid      uuid.UUID   `json:"verein_uuid"`
}

type LinkMeldungAthlet struct {
	ID          int32     `json:"id"`
	Rolle       Rolle     `json:"rolle"`
	Position    int32     `json:"position"`
	MeldungUuid uuid.UUID `json:"meldung_uuid"`
	AthletUuid  uuid.UUID `json:"athlet_uuid"`
}

type Meldung struct {
	Uuid               uuid.UUID   `json:"uuid"`
	DrvRevisionUuid    uuid.UUID   `json:"drv_revision_uuid"`
	Typ                string      `json:"typ"`
	Bemerkung          pgtype.Text `json:"bemerkung"`
	Abgemeldet         bool        `json:"abgemeldet"`
	Dns                bool        `json:"dns"`
	Dnf                bool        `json:"dnf"`
	Dsq                bool        `json:"dsq"`
	ZeitnahmeBemerkung pgtype.Text `json:"zeitnahme_bemerkung"`
	StartNummer        int32       `json:"start_nummer"`
	Abteilung          int32       `json:"abteilung"`
	Bahn               int32       `json:"bahn"`
	Kosten             int32       `json:"kosten"`
	RechnungsNummer    pgtype.Text `json:"rechnungs_nummer"`
	VereinUuid         uuid.UUID   `json:"verein_uuid"`
	RennenUuid         uuid.UUID   `json:"rennen_uuid"`
}

type Obmann struct {
	Uuid       uuid.UUID   `json:"uuid"`
	Name       pgtype.Text `json:"name"`
	Email      pgtype.Text `json:"email"`
	Phone      pgtype.Text `json:"phone"`
	VereinUuid uuid.UUID   `json:"verein_uuid"`
}

type Pause struct {
	ID             int32     `json:"id"`
	Laenge         int32     `json:"laenge"`
	NachRennenUuid uuid.UUID `json:"nach_rennen_uuid"`
}

type Rechnung struct {
	Ulid       string      `json:"ulid"`
	Nummer     string      `json:"nummer"`
	Date       pgtype.Date `json:"date"`
	VereinUuid uuid.UUID   `json:"verein_uuid"`
	CostSum    int32       `json:"cost_sum"`
}

type Rennen struct {
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
	Startzeit        pgtype.Text `json:"startzeit"`
}

type StartnummerAusgabe struct {
	Ulid                      string      `json:"ulid"`
	VereinUuid                uuid.UUID   `json:"verein_uuid"`
	Date                      pgtype.Date `json:"date"`
	Pfand                     int32       `json:"pfand"`
	Kosten                    int32       `json:"kosten"`
	StartnummerAusgegeben     string      `json:"startnummer_ausgegeben"`
	StartnummerZurueckgegeben pgtype.Text `json:"startnummer_zurueckgegeben"`
}

type User struct {
	Ulid           string `json:"ulid"`
	Username       string `json:"username"`
	HashedPassword string `json:"hashed_password"`
	IsActive       bool   `json:"is_active"`
	GroupUlid      string `json:"group_ulid"`
}

type UsersGroup struct {
	Ulid                  string `json:"ulid"`
	Name                  string `json:"name"`
	AllowedAdmin          bool   `json:"allowed_admin"`
	AllowedZeitnahme      bool   `json:"allowed_zeitnahme"`
	AllowedStartlisten    bool   `json:"allowed_startlisten"`
	AllowedRegattaleitung bool   `json:"allowed_regattaleitung"`
}

type Verein struct {
	Uuid     uuid.UUID `json:"uuid"`
	Name     string    `json:"name"`
	Kurzform string    `json:"kurzform"`
	Kuerzel  string    `json:"kuerzel"`
}

type Zahlung struct {
	Ulid       string      `json:"ulid"`
	Nummer     string      `json:"nummer"`
	Date       pgtype.Date `json:"date"`
	VereinUuid uuid.UUID   `json:"verein_uuid"`
	Amount     int32       `json:"amount"`
}

type ZeitnahmeErgebni struct {
	ID               int32     `json:"id"`
	Endzeit          float64   `json:"endzeit"`
	ZeitnahmeStartID int32     `json:"zeitnahme_start_id"`
	ZeitnahmeZielID  int32     `json:"zeitnahme_ziel_id"`
	MeldungUuid      uuid.UUID `json:"meldung_uuid"`
}

type ZeitnahmeStart struct {
	ID              int32            `json:"id"`
	RennenNummer    pgtype.Text      `json:"rennen_nummer"`
	StartNummer     pgtype.Text      `json:"start_nummer"`
	TimeClient      pgtype.Timestamp `json:"time_client"`
	TimeServer      pgtype.Timestamp `json:"time_server"`
	MeasuredLatency pgtype.Int4      `json:"measured_latency"`
	Verarbeitet     bool             `json:"verarbeitet"`
}

type ZeitnahmeZiel struct {
	ID              int32            `json:"id"`
	RennenNummer    pgtype.Text      `json:"rennen_nummer"`
	StartNummer     pgtype.Text      `json:"start_nummer"`
	TimeClient      pgtype.Timestamp `json:"time_client"`
	TimeServer      pgtype.Timestamp `json:"time_server"`
	MeasuredLatency pgtype.Int4      `json:"measured_latency"`
	Verarbeitet     bool             `json:"verarbeitet"`
}
