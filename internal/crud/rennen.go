package crud

import (
	"cmp"
	"slices"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type GetAllRennenParams struct {
	GetMeldungen  bool
	GetAthleten   bool
	ShowEmpty     bool
	ShowStarted   bool
	ShowWettkampf sqlc.NullWettkampf
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
	Tag              sqlc.Tag        `json:"tag"`
	Wettkampf        sqlc.Wettkampf  `json:"wettkampf"`
	KostenEur        int             `json:"kosten_eur"`
	Rennabstand      int             `json:"rennabstand"`
	Startzeit        string          `json:"startzeit"`
	NumMeldungen     *int            `json:"num_meldungen"`
	NumAbteilungen   *int            `json:"num_abteilungen"`
	Meldungen        []Meldung       `json:"meldungen"`
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
		log.Debug("Query error: ", err)
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
	retLs := []Rennen{}
	ctx, cancel := getCtxWithTo()
	defer cancel()

	rLs, err := DB.Queries.GetAllRennen(ctx)
	if err != nil {
		return retLs, err
	}
	qLs, err := DB.Queries.GetAllRennenWithAthlet(
		ctx,
		[]sqlc.Wettkampf{
			sqlc.WettkampfLangstrecke,
			sqlc.WettkampfSlalom,
			sqlc.WettkampfKurzstrecke,
			sqlc.WettkampfStaffel,
		},
	)
	if err != nil {
		return retLs, err
	}

	for _, r := range rLs {
		rennen := RennenFromSqlc(r.Rennen, int(r.NumMeldungen), r.NumAbteilungen)
		retLs = append(retLs, rennen)
	}

	i := 0
	for _, q := range qLs {
		for retLs[i].Uuid != q.Rennen.Uuid {
			i++
			continue
		}

		indexLastMeld := len(retLs[i].Meldungen) - 1
		if indexLastMeld < 0 || retLs[i].Meldungen[indexLastMeld].Uuid != q.Meldung.Uuid {
			position := int(q.Position)
			retLs[i].Meldungen = append(retLs[i].Meldungen, Meldung{
				Meldung: q.Meldung,
				Verein:  &Verein{Verein: q.Verein},
				Athleten: []Athlet{{
					Athlet:   q.Athlet,
					Rolle:    &q.Rolle,
					Position: &position,
				}},
			})
		} else {
			position := int(q.Position)
			retLs[i].Meldungen[indexLastMeld].Athleten = append(retLs[i].Meldungen[indexLastMeld].Athleten, Athlet{
				Athlet:   q.Athlet,
				Rolle:    &q.Rolle,
				Position: &position,
			})
		}
	}

	for _, r := range retLs {
		// Sort Meldungen
		slices.SortFunc(r.Meldungen, func(a, b Meldung) int {
			return cmp.Or(
				cmp.Compare(a.Abteilung, b.Abteilung),
				cmp.Compare(a.Bahn, b.Bahn),
			)
		})
	}
	return retLs, nil
}

func RennenFromSqlc(rennen sqlc.Rennen, numMeld int, numAbt interface{}) Rennen {
	numAbteilungenI32, ok := numAbt.(int32)
	if !ok {
		log.Error("Error converting numAbt to int32 ", numAbt)
		numAbteilungenI32 = 0
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
		Tag:              rennen.Tag,
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
			curRennen.Meldungen = append(curRennen.Meldungen, Meldung{
				Meldung: sqlc.Meldung{
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
				},
				Rennen:   &Rennen{},
				Verein:   &Verein{},
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
