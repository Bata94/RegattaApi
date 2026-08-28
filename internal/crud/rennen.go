package crud

import (
	"cmp"
	"context"
	jsonv2 "encoding/json/v2"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/uuid"
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

	return "", apierr.ErrNotFound
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
	sqlc.Rennen
	Tag            Tag       `json:"tag"`
	NumMeldungen   *int      `json:"num_meldungen,omitempty"`
	NumAbteilungen *int      `json:"num_abteilungen,omitempty"`
	Meldungen      []Meldung `json:"meldungen,omitempty"`
}

func (r *Rennen) GetZusatz() *string {
	if r.Zusatz.Valid {
		return &r.Zusatz.String
	}
	return nil
}

func (r *Rennen) GetKostenEur() *int {
	if r.KostenEur.Valid {
		v := int(r.KostenEur.Int32)
		return &v
	}
	return nil
}

func (r *Rennen) KostenEurStr() string {
	if v := r.GetKostenEur(); v != nil {
		return strconv.Itoa(*v)
	}
	return ""
}

func (r *Rennen) GetRennabstand() *int {
	if r.Rennabstand.Valid {
		v := int(r.Rennabstand.Int32)
		return &v
	}
	return nil
}

func (r *Rennen) GetStartzeit() *string {
	if r.Startzeit.Valid {
		return &r.Startzeit.String
	}
	return nil
}

func (r *Rennen) GetMeldungen(ctx context.Context) ([]Meldung, error) {
	if r.Meldungen != nil {
		return r.Meldungen, nil
	}
	slog.Warn("lazy loading Meldungen", "rennen_uuid", r.Uuid.String())
	loaded, err := GetRennenMeldungen(ctx, r.Uuid)
	if err != nil {
		return nil, err
	}
	r.Meldungen = loaded
	numMeld := len(r.Meldungen)
	r.NumMeldungen = &numMeld
	return r.Meldungen, nil
}

func (r Rennen) StartzeitStr() string {
	if v := r.GetStartzeit(); v != nil {
		return *v
	}
	return ""
}

func (r Rennen) ZusatzStr() string {
	if v := r.GetZusatz(); v != nil {
		return *v
	}
	return ""
}

func (r Rennen) RennabstandStr() string {
	if v := r.GetRennabstand(); v != nil {
		return strconv.Itoa(*v)
	}
	return ""
}

func (r Rennen) RennabstandInt() int {
	if v := r.GetRennabstand(); v != nil {
		return *v
	}
	return 0
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

func RennenFromSqlc(rennen sqlc.Rennen, numMeld int, numAbt any) Rennen {
	var numAbteilungenI32 int32
	switch v := numAbt.(type) {
	case int32:
		numAbteilungenI32 = v
	case int:
		numAbteilungenI32 = int32(v)
	default:
		slog.Error(fmt.Sprintf("Error converting numAbt to int32: %v (%T)", numAbt, numAbt))
	}
	numAbteilungen := int(numAbteilungenI32)
	return Rennen{
		Rennen:         rennen,
		Tag:            Tag(rennen.Tag),
		NumMeldungen:   &numMeld,
		NumAbteilungen: &numAbteilungen,
	}
}

func MaxStartLanesForWettkampf(w sqlc.Wettkampf) int {
	switch w {
	case sqlc.WettkampfLangstrecke:
		return 1
	case sqlc.WettkampfKurzstrecke:
		return 4
	case sqlc.WettkampfSlalom:
		return 3
	case sqlc.WettkampfStaffel:
		return 2
	}
	return 1
}

type rennenJSON struct {
	Uuid             uuid.UUID       `json:"uuid"`
	SortID           int32           `json:"sort_id"`
	Nummer           string          `json:"nummer"`
	Bezeichnung      string          `json:"bezeichnung"`
	BezeichnungLang  string          `json:"bezeichnung_lang"`
	Zusatz           *string         `json:"zusatz,omitempty"`
	Leichtgewicht    bool            `json:"leichtgewicht"`
	Geschlecht       sqlc.Geschlecht `json:"geschlecht"`
	Bootsklasse      string          `json:"bootsklasse"`
	BootsklasseLang  string          `json:"bootsklasse_lang"`
	Altersklasse     string          `json:"altersklasse"`
	AltersklasseLang string          `json:"altersklasse_lang"`
	Tag              Tag             `json:"tag"`
	Wettkampf        sqlc.Wettkampf  `json:"wettkampf"`
	KostenEur        *int            `json:"kosten_eur,omitempty"`
	Rennabstand      *int            `json:"rennabstand,omitempty"`
	Startzeit        *string         `json:"startzeit,omitempty"`
	NumMeldungen     *int            `json:"num_meldungen,omitempty"`
	NumAbteilungen   *int            `json:"num_abteilungen,omitempty"`
	MaxLanes         int             `json:"max_lanes"`
	Meldungen        []Meldung       `json:"meldungen,omitempty"`
}

func (r Rennen) MarshalJSON() ([]byte, error) {
	j := rennenJSON{
		Uuid:             r.Uuid,
		SortID:           r.SortID,
		Nummer:           r.Nummer,
		Bezeichnung:      r.Bezeichnung,
		BezeichnungLang:  r.BezeichnungLang,
		Zusatz:           r.GetZusatz(),
		Leichtgewicht:    r.Leichtgewicht,
		Geschlecht:       r.Geschlecht,
		Bootsklasse:      r.Bootsklasse,
		BootsklasseLang:  r.BootsklasseLang,
		Altersklasse:     r.Altersklasse,
		AltersklasseLang: r.AltersklasseLang,
		Tag:              r.Tag,
		Wettkampf:        r.Wettkampf,
		KostenEur:        r.GetKostenEur(),
		Rennabstand:      r.GetRennabstand(),
		Startzeit:        r.GetStartzeit(),
		NumMeldungen:     r.NumMeldungen,
		NumAbteilungen:   r.NumAbteilungen,
		MaxLanes:         MaxStartLanesForWettkampf(r.Wettkampf),
	}
	if r.Meldungen != nil {
		j.Meldungen = r.Meldungen
	}
	return jsonv2.Marshal(j)
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

func GetZeitplanung(ctx context.Context, wettkampf []sqlc.Wettkampf) (Zeitplaung, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.QueriesFromCtx(ctx).GetRennenZeitplan(ctx, wettkampf)
	if err != nil {
		return Zeitplaung{}, err
	}
	if len(q) == 0 {
		return Zeitplaung{}, apierr.ErrNotFound
	}

	return ZeitplaungFromSqlc(q), nil
}

func GetAllRennen(ctx context.Context, p GetAllRennenParams) ([]Rennen, error) {
	ctx, cancel := getCtx(ctx)
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

	q, err = DB.QueriesFromCtx(ctx).GetAllRennenWithMeld(ctx, wettkampfFilterLs)
	if err != nil {
		slog.Error("Query error", "err", err)
		return nil, err
	}

	rLs := sqlcRennenToCrudRennen(q, true)
	retLs := []Rennen{}

	for _, r := range rLs {
		if !p.GetMeldungen {
			r.Meldungen = nil
		}
		if !p.ShowEmpty && *r.NumMeldungen == 0 {
			continue
		}

		retLs = append(retLs, r)
	}

	if !p.ShowStarted {
		started, err := startedRennenNummern(ctx)
		if err != nil {
			slog.Error("Query error", "err", err)
			return nil, err
		}
		retLs = filterStartedRennen(retLs, started)
	}

	return retLs, nil
}

func startedRennenNummern(ctx context.Context) (map[string]struct{}, error) {
	rows, err := DB.QueriesFromCtx(ctx).GetAllZeitnahmeStart(ctx)
	if err != nil {
		return nil, err
	}

	started := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		if r.RennenNummer.Valid {
			started[r.RennenNummer.String] = struct{}{}
		}
	}
	return started, nil
}

func filterStartedRennen(rennen []Rennen, started map[string]struct{}) []Rennen {
	if len(started) == 0 {
		return rennen
	}

	retLs := make([]Rennen, 0, len(rennen))
	for _, r := range rennen {
		if _, ok := started[r.Nummer]; !ok {
			retLs = append(retLs, r)
		}
	}
	return retLs
}

func GetAllRennenWithAthlet(ctx context.Context, p GetAllRennenParams) ([]Rennen, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	var wettkampfFilterLs []sqlc.Wettkampf
	if !p.ShowWettkampf.Valid {
		wettkampfFilterLs = AllWettkampf
	} else {
		wettkampfFilterLs = []sqlc.Wettkampf{p.ShowWettkampf.Wettkampf}
	}

	baseRows, err := DB.QueriesFromCtx(ctx).GetAllRennenWithMeld(ctx, wettkampfFilterLs)
	if err != nil {
		return nil, err
	}
	races := sqlcRennenToCrudRennen(baseRows, p.ShowEmpty)

	if !p.ShowStarted {
		started, err := startedRennenNummern(ctx)
		if err != nil {
			return nil, err
		}
		races = filterStartedRennen(races, started)
	}

	if !p.GetAthleten {
		return races, nil
	}

	athRows, err := DB.QueriesFromCtx(ctx).GetAllRennenAthletRows(ctx, wettkampfFilterLs)
	if err != nil {
		return races, err
	}
	raceIndex := make(map[uuid.UUID]int, len(races))
	for i, r := range races {
		raceIndex[r.Uuid] = i
	}
	for _, ar := range athRows {
		ri, ok := raceIndex[ar.RennenUuid]
		if !ok {
			continue
		}
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
			}
			curRennen = RennenFromSqlc(row.Rennen, int(row.NumMeldungen), row.NumAbteilungen)
		}

		if row.Uuid != uuid.Nil {
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
			verein := Verein{Verein: v}
			curRennen.Meldungen = append(curRennen.Meldungen, Meldung{
				Meldung:  m,
				Rennen:   &Rennen{},
				Verein:   &verein,
				Athleten: []Athlet{},
			})
		}
	}

	if getEmptyRennen || len(curRennen.Meldungen) != 0 {
		rLs = append(rLs, curRennen)
	}

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

func GetRennenMinimal(ctx context.Context, uuid uuid.UUID) (Rennen, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	r, err := DB.QueriesFromCtx(ctx).GetRennenMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Rennen{}, apierr.ErrNotFound
		}
		return Rennen{}, err
	}

	return RennenFromSqlc(r, 0, 0), nil
}

func GetRennen(ctx context.Context, uuidParam uuid.UUID) (Rennen, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	rSqlc, err := DB.QueriesFromCtx(ctx).GetRennenMinimal(ctx, uuidParam)
	if err != nil {
		if isNoRowError(err) {
			return Rennen{}, apierr.ErrNotFound
		}
		return Rennen{}, err
	}

	r := RennenFromSqlc(rSqlc, 0, int32(0))

	meldungen, err := GetRennenMeldungen(ctx, uuidParam)
	if err != nil {
		return Rennen{}, err
	}
	r.Meldungen = meldungen

	numMeldungen := len(r.Meldungen)
	r.NumMeldungen = &numMeldungen
	numAbt := int32(0)
	for _, m := range r.Meldungen {
		if m.Abteilung > numAbt {
			numAbt = m.Abteilung
		}
	}
	numAbteilungen := int(numAbt)
	r.NumAbteilungen = &numAbteilungen

	return r, nil
}

func GetRennenMeldungen(ctx context.Context, rennenUuid uuid.UUID) ([]Meldung, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.QueriesFromCtx(ctx).GetAllRennenMeldungen(ctx, rennenUuid)
	if err != nil {
		if isNoRowError(err) {
			return []Meldung{}, nil
		}
		return nil, err
	}

	meldungen := []Meldung{}
	for _, row := range q {
		if len(meldungen) == 0 || meldungen[len(meldungen)-1].Uuid != row.Meldung.Uuid {
			verein := Verein{Verein: row.Verein}
			meldungen = append(meldungen, Meldung{
				Meldung: row.Meldung,
				Verein:  &verein,
			})
		}
		lastIdx := len(meldungen) - 1
		pos := int(row.Position)
		meldungen[lastIdx].Athleten = append(meldungen[lastIdx].Athleten, Athlet{
			Athlet:   row.Athlet,
			Rolle:    &row.Rolle,
			Position: &pos,
		})
	}
	return meldungen, nil
}

func UpdateStartZeit(ctx context.Context, params sqlc.UpdateStartZeitParams) error {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).UpdateStartZeit(ctx, params)
}

func CreateRennen(ctx context.Context, rParams sqlc.CreateRennenParams) (Rennen, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	r, err := DB.QueriesFromCtx(ctx).CreateRennen(ctx, rParams)
	if err != nil {
		return Rennen{}, err
	}

	return RennenFromSqlc(r, 0, 0), nil
}
