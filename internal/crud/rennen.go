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

func GetAllRennen(p GetAllRennenParams) ([]RennenWithMeldung, error) {
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
		return nil, err
	}

	rLs := sqlcRennenToCrudRennen(q, true)
	retLs := []RennenWithMeldung{}

	for _, r := range rLs {
		meldungen := []MeldungMinimal{}
		if p.GetMeldungen {
			meldungen = r.Meldungen
		}
		if p.ShowStarted == false {
			// TODO: Implement!
		}
		if p.ShowEmpty == false && r.NumMeldungen == 0 {
			continue
		}

		retLs = append(retLs, RennenWithMeldung{
			Rennen:    r.Rennen,
			Meldungen: meldungen,
		})
	}
	return retLs, nil
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
	NumMeldungen     int             `json:"num_meldungen"`
	NumAbteilungen   int             `json:"num_abteilungen"`
}
type RennenWithMeldungAndAthlet struct {
	Rennen
	Meldungen []Meldung
}

func RennenFromSqlc(rennen sqlc.Rennen, numMeld int, numAbt interface{}) Rennen {
	numAbteilungen, err := numAbt.(int)
	if err {
		log.Error("Error converting numAbt to int", numAbt)
		numAbteilungen = 0
	}
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
		NumMeldungen:     numMeld,
		NumAbteilungen:   numAbteilungen,
	}
}

type RennenWithMeldung struct {
	Rennen
	Meldungen []MeldungMinimal `json:"meldungen"`
}

func sqlcRennenToCrudRennen(q []sqlc.GetAllRennenWithMeldRow, getEmptyRennen bool) []RennenWithMeldung {
	var curRennen RennenWithMeldung
	rLs := []RennenWithMeldung{}

	for i, row := range q {
		if i == 0 {
			curRennen = RennenWithMeldung{
				Rennen:    RennenFromSqlc(row.Rennen, int(row.NumMeldungen), row.NumAbteilungen),
				Meldungen: []MeldungMinimal{},
			}
		}

		if row.Rennen.Uuid != curRennen.Uuid {
			if getEmptyRennen || len(curRennen.Meldungen) != 0 {
				rLs = append(rLs, curRennen)
				curRennen = RennenWithMeldung{
					Rennen:    RennenFromSqlc(row.Rennen, int(row.NumMeldungen), row.NumAbteilungen),
					Meldungen: []MeldungMinimal{},
				}
			}
		}

		if row.Uuid != uuid.Nil {
			curRennen.Meldungen = append(curRennen.Meldungen, SqlcMeldungMinmalToCrudMeldungMinimal(sqlc.Meldung{
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
				VereinUuid:         row.VereinUuid,
				RennenUuid:         row.RennenUuid,
			}))
		}
	}
	// Make sure last rennen is added
	if len(q) > 0 && rLs[len(rLs)-1].Rennen.Uuid != curRennen.Uuid {
		if getEmptyRennen || len(curRennen.Meldungen) != 0 {
			rLs = append(rLs, curRennen)
		}
	}

	// sort Meldungen
	for _, r := range rLs {
		slices.SortFunc(r.Meldungen, func(a, b MeldungMinimal) int {
			return cmp.Or(
				cmp.Compare(a.Abteilung, b.Abteilung),
				cmp.Compare(a.Bahn, b.Bahn),
			)
		})
	}

	return rLs
}

func GetRennenMinimal(uuid uuid.UUID) (sqlc.Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	r, err := DB.Queries.GetRennenMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return sqlc.Rennen{}, &api.NOT_FOUND
		}
		return sqlc.Rennen{}, err
	}

	return r, nil
}

func GetRennen(uuidParam uuid.UUID) (RennenWithMeldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetRennen(ctx, uuidParam)
	if err != nil {
		if isNoRowError(err) {
			return RennenWithMeldung{}, &api.NOT_FOUND
		}
		return RennenWithMeldung{}, err
	}
	if len(q) == 0 {
		return RennenWithMeldung{}, &api.NOT_FOUND
	}

	r := RennenWithMeldung{
		Rennen: Rennen{
			Uuid:             q[0].Uuid,
			SortID:           int(q[0].SortID.Int32),
			Nummer:           q[0].Nummer.String,
			Bezeichnung:      q[0].Bezeichnung.String,
			BezeichnungLang:  q[0].BezeichnungLang.String,
			Zusatz:           q[0].Zusatz.String,
			Leichtgewicht:    q[0].Leichtgewicht.Bool,
			Geschlecht:       q[0].Geschlecht.Geschlecht,
			Bootsklasse:      q[0].Bootsklasse.String,
			BootsklasseLang:  q[0].BootsklasseLang.String,
			Altersklasse:     q[0].Altersklasse.String,
			AltersklasseLang: q[0].AltersklasseLang.String,
			Tag:              q[0].Tag.Tag,
			Wettkampf:        q[0].Wettkampf.Wettkampf,
			KostenEur:        int(q[0].KostenEur.Int32),
			Rennabstand:      int(q[0].Rennabstand.Int32),
			Startzeit:        q[0].Startzeit.String,
			NumMeldungen:     0,
			NumAbteilungen:   0,
		},
		Meldungen: []MeldungMinimal{},
	}

	numAbt := 0
	if q[0].Uuid_2 != uuid.Nil {
		for _, row := range q {
			if numAbt < int(row.Abteilung.Int32) {
				numAbt = int(row.Abteilung.Int32)
			}

			r.Meldungen = append(r.Meldungen, SqlcMeldungMinmalToCrudMeldungMinimal(sqlc.Meldung{
				Uuid:               row.Uuid_2,
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
				VereinUuid:         row.VereinUuid,
				RennenUuid:         row.RennenUuid,
			}))
		}
	}
	r.NumMeldungen = len(r.Meldungen)
	r.NumAbteilungen = int(numAbt)

	return r, nil
}

func UpdateStartZeit(params sqlc.UpdateStartZeitParams) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	return DB.Queries.UpdateStartZeit(ctx, params)
}

func CreateRennen(rParams sqlc.CreateRennenParams) (sqlc.Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.CreateRennen(ctx, rParams)
	if err != nil {
		return sqlc.Rennen{}, err
	}

	return v, nil
}
