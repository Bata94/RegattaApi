package crud

import (
	"cmp"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	AllWettkampf = []sqlc.Wettkampf{
		sqlc.WettkampfLangstrecke,
		sqlc.WettkampfSlalom,
		sqlc.WettkampfKurzstrecke,
		sqlc.WettkampfStaffel,
	}
)

func WettkampfFromString(str string) (sqlc.Wettkampf, error) {
	caser := cases.Title(language.German)
	str = caser.String(str)

	for _, w := range AllWettkampf {
		if string(w) == str {
			return w, nil
		}
	}

	return "", &api.NOT_FOUND
}

type GetAllRennenParams struct {
	GetMeldungen  bool
	GetAthleten   bool
	ShowEmpty     bool
	ShowStarted   bool
	ShowWettkampf sqlc.NullWettkampf
}

type Tag sqlc.Tag

const (
	TagSa Tag = "sa"
	TagSo Tag = "so"
)

func (t Tag) String() string {
	switch t {
	case TagSa:
		return "Samstag"
	case TagSo:
		return "Sonntag"
	}
	return "Unbekannter Tag"
}

func (t Tag) StringShort() string {
	switch t {
	case TagSa:
		return "Sa"
	case TagSo:
		return "So"
	}
	return "???"
}

type Rennen struct {
	Uuid             uuid.UUID       `json:"uuid"`
	SortID           int             `json:"sort_id"`
	Nummer           string          `json:"nummer"`
	Bezeichnung      string          `json:"bezeichnung"`
	BezeichnungLang  string          `json:"bezeichnung_lang"`
	Zusatz           string          `json:"zusatz"`
	Leichtgewicht    bool            `json:"leichtgewicht"`
	Geschlecht       sqlc.Geschlecht `json:"geschlecht"`
	Bootsklasse      string          `json:"bootsklasse"`
	BootsklasseLang  string          `json:"bootsklasse_lang"`
	Altersklasse     string          `json:"altersklasse"`
	AltersklasseLang string          `json:"altersklasse_lang"`
	Tag              Tag             `json:"tag"`
	Wettkampf        sqlc.Wettkampf  `json:"wettkampf"`
	KostenEur        int             `json:"kosten_eur"`
	Rennabstand      int             `json:"rennabstand"`
	Startzeit        string          `json:"startzeit"`
	NumMeldungen     *int            `json:"num_meldungen"`
	NumAbteilungen   *int            `json:"num_abteilungen"`
	Meldungen        []Meldung       `json:"meldungen"`
}

// returns number of athletes and if Stm is required
// If the number of athletes is 0, there is an error parsing the string
func (r Rennen) GetTeilnehmerMeldeParams() (int, bool) {
	numAthletes, err := strconv.Atoi(r.Bootsklasse[:1])
	if err != nil {
		numAthletes = 0
	}
	stmRequired := strings.Contains(r.Bootsklasse, "+")

	return numAthletes, stmRequired
}

type Zeitplaung struct {
	Rennen []RennenZeitplaung `json:"rennen"`
}

type RennenZeitplaung struct {
	Uuid             uuid.UUID      `json:"uuid"`
	Sort_id          int            `json:"sort_id"`
	Nummer           string         `json:"nummer"`
	Bezeichnung      string         `json:"bezeichnung"`
	Bezeichnung_lang string         `json:"bezeichnung_lang"`
	Zusatz           string         `json:"zusatz"`
	Wettkampf        sqlc.Wettkampf `json:"wettkampf"`
	Tag              Tag            `json:"tag"`
	Startzeit        string         `json:"startzei"`
}

func RennenZeitplaungFromSqlc(rennen sqlc.GetRennenZeitplanRow) RennenZeitplaung {
	return RennenZeitplaung{
		Uuid:             rennen.Uuid,
		Sort_id:          int(rennen.SortID),
		Nummer:           rennen.Nummer,
		Bezeichnung:      rennen.Bezeichnung,
		Bezeichnung_lang: rennen.BezeichnungLang,
		Zusatz:           rennen.Zusatz.String,
		Wettkampf:        rennen.Wettkampf,
		Tag:              Tag(rennen.Tag),
		Startzeit:        rennen.Startzeit.String,
	}
}

func ZeitplaungFromSqlc(rennen []sqlc.GetRennenZeitplanRow) Zeitplaung {
	var retLs []RennenZeitplaung
	for _, r := range rennen {
		retLs = append(retLs, RennenZeitplaungFromSqlc(r))
	}
	return Zeitplaung{
		Rennen: retLs,
	}
}

func GetZeitplanung(wettkampf []sqlc.Wettkampf) (Zeitplaung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetRennenZeitplan(ctx, wettkampf)
	if err != nil {
		return Zeitplaung{}, err
	}
	if len(q) == 0 {
		return Zeitplaung{}, &api.NOT_FOUND
	}

	return ZeitplaungFromSqlc(q), nil
}

func GetAllRennen(p GetAllRennenParams) ([]Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	var (
		q                 []sqlc.GetAllRennenWithMeldRow
		err               error
		wettkampfFilterLs []sqlc.Wettkampf
	)
	allWettkampf := []sqlc.Wettkampf{
		sqlc.WettkampfLangstrecke,
		sqlc.WettkampfSlalom,
		sqlc.WettkampfKurzstrecke,
		sqlc.WettkampfStaffel,
	}

	if !p.ShowWettkampf.Valid {
		wettkampfFilterLs = allWettkampf
	} else {
		wettkampfFilterLs = []sqlc.Wettkampf{p.ShowWettkampf.Wettkampf}
	}

	q, err = DB.Queries.GetAllRennenWithMeld(ctx, wettkampfFilterLs)
	if err != nil {
		log.Println("Query error: ", err)
		return nil, err
	}

	rLs := sqlcRennenToCrudRennen(q, true)
	retLs := []Rennen{}

	for _, r := range rLs {
		meldungen := []Meldung{}
		if p.GetMeldungen {
			meldungen = r.Meldungen
		}
		if p.ShowStarted == false {
			// TODO: Implement!
		}
		if p.ShowEmpty == false && *r.NumMeldungen == 0 {
			continue
		}

		rennen := r
		rennen.Meldungen = meldungen
		retLs = append(retLs, rennen)
	}
	return retLs, nil
}

func GetAllRennenWithAthlet(p GetAllRennenParams) ([]Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	// Filter by Wettkampf
	var wettkampfFilterLs []sqlc.Wettkampf
	if !p.ShowWettkampf.Valid {
		wettkampfFilterLs = AllWettkampf
	} else {
		wettkampfFilterLs = []sqlc.Wettkampf{p.ShowWettkampf.Wettkampf}
	}

	// Phase 1: fetch races with Meldungen
	baseRows, err := DB.Queries.GetAllRennenWithMeld(ctx, wettkampfFilterLs)
	if err != nil {
		return nil, err
	}
	races := sqlcRennenToCrudRennen(baseRows, p.ShowEmpty)

	if !p.GetAthleten {
		return races, nil
	}

	// Phase 2: fetch athlete rows and merge into races
	athRows, err := DB.Queries.GetAllRennenAthletRows(ctx, wettkampfFilterLs)
	if err != nil {
		return races, err
	}
	// Index races by UUID
	raceIndex := make(map[uuid.UUID]int, len(races))
	for i, r := range races {
		raceIndex[r.Uuid] = i
	}
	// Merge athlete entries
	for _, ar := range athRows {
		ri, ok := raceIndex[ar.RennenUuid]
		if !ok {
			continue
		}
		// find Meldung slot
		mi := -1
		for j, m := range races[ri].Meldungen {
			if m.Uuid == ar.MeldungUuid {
				mi = j
				break
			}
		}
		if mi < 0 {
			continue
		}
		pos := int(ar.Position)
		rolePtr := &ar.Rolle
		races[ri].Meldungen[mi].Athleten = append(
			races[ri].Meldungen[mi].Athleten,
			Athlet{Athlet: ar.Athlet, Rolle: rolePtr, Position: &pos},
		)
	}
	return races, nil
}

func RennenFromSqlc(rennen sqlc.Rennen, numMeld int, numAbt any) Rennen {
	var numAbteilungenI32 int32
	switch v := numAbt.(type) {
	case int32:
		numAbteilungenI32 = v
	case int:
		numAbteilungenI32 = int32(v)
	default:
		log.Printf("Error converting numAbt to int32: %v (%T)", numAbt, numAbt)
	}
	numAbteilungen := int(numAbteilungenI32)
	return Rennen{
		Uuid:             rennen.Uuid,
		SortID:           int(rennen.SortID),
		Nummer:           rennen.Nummer,
		Bezeichnung:      rennen.Bezeichnung,
		BezeichnungLang:  rennen.BezeichnungLang,
		Zusatz:           rennen.Zusatz.String,
		Leichtgewicht:    rennen.Leichtgewicht,
		Geschlecht:       rennen.Geschlecht,
		Bootsklasse:      rennen.Bootsklasse,
		BootsklasseLang:  rennen.BootsklasseLang,
		Altersklasse:     rennen.Altersklasse,
		AltersklasseLang: rennen.AltersklasseLang,
		Tag:              Tag(rennen.Tag),
		Wettkampf:        rennen.Wettkampf,
		KostenEur:        int(rennen.KostenEur.Int32),
		Rennabstand:      int(rennen.Rennabstand.Int32),
		Startzeit:        rennen.Startzeit.String,
		NumMeldungen:     &numMeld,
		NumAbteilungen:   &numAbteilungen,
	}
}

func sqlcRennenToCrudRennen(q []sqlc.GetAllRennenWithMeldRow, getEmptyRennen bool) []Rennen {
	var curRennen Rennen
	rLs := []Rennen{}

	for i, row := range q {
		if i == 0 {
			curRennen = RennenFromSqlc(row.Rennen, int(row.NumMeldungen), row.NumAbteilungen)
		}

		if row.Rennen.Uuid != curRennen.Uuid {
			if getEmptyRennen || len(curRennen.Meldungen) != 0 {
				rLs = append(rLs, curRennen)
				curRennen = RennenFromSqlc(row.Rennen, int(row.NumMeldungen), row.NumAbteilungen)
			}
		}

		if row.Uuid != uuid.Nil {
			// Construct Meldung and associated Verein
			v := sqlc.Verein{
				Uuid:     row.VereinUuid,
				Name:     row.Name.String,
				Kurzform: row.Kurzform.String,
				Kuerzel:  row.Kuerzel.String,
			}
			m := sqlc.Meldung{
				Uuid:               row.Uuid,
				DrvRevisionUuid:    row.DrvRevisionUuid,
				Typ:                row.Typ.String,
				Bemerkung:          row.Bemerkung,
				Abgemeldet:         row.Abgemeldet.Bool,
				Dns:                row.Dns.Bool,
				Dnf:                row.Dnf.Bool,
				Dsq:                row.Dsq.Bool,
				ZeitnahmeBemerkung: row.ZeitnahmeBemerkung,
				StartNummer:        row.StartNummer.Int32,
				Abteilung:          row.Abteilung.Int32,
				Bahn:               row.Bahn.Int32,
				Kosten:             row.Kosten.Int32,
				RechnungsNummer:    row.RechnungsNummer,
				VereinUuid:         row.VereinUuid,
				RennenUuid:         row.RennenUuid,
			}
			curRennen.Meldungen = append(curRennen.Meldungen, Meldung{
				Meldung:  m,
				Rennen:   &Rennen{},
				Verein:   &Verein{Verein: v},
				Athleten: []Athlet{},
			})
		}
	}

	// sort Meldungen
	for _, r := range rLs {
		slices.SortFunc(r.Meldungen, func(a, b Meldung) int {
			return cmp.Or(
				cmp.Compare(a.Abteilung, b.Abteilung),
				cmp.Compare(a.Bahn, b.Bahn),
			)
		})
	}

	return rLs
}

func GetRennenMinimal(uuid uuid.UUID) (Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	r, err := DB.Queries.GetRennenMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Rennen{}, &api.NOT_FOUND
		}
		return Rennen{}, err
	}

	return RennenFromSqlc(r, 0, 0), nil
}

func GetRennen(uuidParam uuid.UUID) (Rennen, error) {
	// TODO: Implement queryParams
	// TODO: Fix NULL scan issue - SQL query uses FULL JOINs that return NULL values
	//       but sqlc generated non-pointer types can't handle NULL
	//       Fix by using COALESCE in SQL or regenerating sqlc with nullable types
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetRennen(ctx, uuidParam)
	if err != nil {
		if isNoRowError(err) {
			return Rennen{}, &api.NOT_FOUND
		}
		return Rennen{}, err
	}
	if len(q) == 0 {
		return Rennen{}, &api.NOT_FOUND
	}

	r := RennenFromSqlc(q[0].Rennen, 0, int32(0))
	r.Meldungen = []Meldung{}

	numAbt := 0
	if q[0].Meldung.Uuid != uuid.Nil {
		for i, row := range q {
			meld := row.Meldung
			if numAbt < int(meld.Abteilung) {
				numAbt = int(meld.Abteilung)
			}
			if i == 0 || meld.Uuid != q[i-1].Meldung.Uuid {
				r.Meldungen = append(
					r.Meldungen,
					Meldung{
						Meldung:  meld,
						Verein:   &Verein{Verein: row.Verein},
						Athleten: []Athlet{},
					},
				)
			}

			athlet := row.Athlet
			if athlet.Uuid != uuid.Nil {
				lastMeldIndex := len(r.Meldungen) - 1
				position := int(row.LinkMeldungAthlet.Position)
				r.Meldungen[lastMeldIndex].Athleten = append(r.Meldungen[lastMeldIndex].Athleten, Athlet{
					Athlet:   athlet,
					Rolle:    &row.LinkMeldungAthlet.Rolle,
					Position: &position,
				})
			}
		}
	}
	numMeldungen := len(r.Meldungen)
	r.NumMeldungen = &numMeldungen
	numAbteilungen := int(numAbt)
	r.NumAbteilungen = &numAbteilungen

	return r, nil
}

func UpdateStartZeit(params sqlc.UpdateStartZeitParams) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.UpdateStartZeit(ctx, params)
}

func CreateRennen(rParams sqlc.CreateRennenParams) (Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	r, err := DB.Queries.CreateRennen(ctx, rParams)
	if err != nil {
		return Rennen{}, err
	}

	return RennenFromSqlc(r, 0, 0), nil
}
