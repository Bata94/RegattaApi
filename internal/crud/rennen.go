package crud

import (
	"cmp"
	"errors"
	"slices"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

func GetAllRennen() ([]*sqlc.Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	rLs, err := DB.Queries.GetAllRennen(ctx)
	if err != nil {
		return nil, err
	}
	if rLs == nil {
		rLs = []*sqlc.Rennen{}
	}

	return rLs, nil
}

type Rennen struct {
	Uuid             uuid.UUID           `json:"uuid"`
	SortID           int32               `json:"sort_id"`
	Nummer           string              `json:"nummer"`
	Bezeichnung      *string             `json:"bezeichnung"`
	BezeichnungLang  *string             `json:"bezeichnung_lang"`
	Zusatz           *string             `json:"zusatz"`
	Leichtgewicht    *bool               `json:"leichtgewicht"`
	Geschlecht       sqlc.NullGeschlecht `json:"geschlecht"`
	Bootsklasse      *string             `json:"bootsklasse"`
	BootsklasseLang  *string             `json:"bootsklasse_lang"`
	Altersklasse     *string             `json:"altersklasse"`
	AltersklasseLang *string             `json:"altersklasse_lang"`
	Tag              sqlc.Tag            `json:"tag"`
	Wettkampf        sqlc.Wettkampf      `json:"wettkampf"`
	KostenEur        *int32              `json:"kosten_eur"`
	Rennabstand      *int32              `json:"rennabstand"`
	Startzeit        *string             `json:"startzeit"`
	NumMeldungen     int                 `json:"num_meldungen"`
	NumAbteilungen   int                 `json:"num_abteilungen"`
}

type RennenWithMeldung struct {
	*Rennen
	Meldungen []*sqlc.Meldung
}

func sqlcRennenToCrudRennen(q []*sqlc.GetAllRennenWithMeldRow, getEmptyRennen bool) ([]*RennenWithMeldung, error) {
	var (
		rLs           []*RennenWithMeldung
		currentRennen *RennenWithMeldung
	)

	for _, m := range q {
		rennenStruct := &Rennen{
			Uuid:             m.Uuid,
			SortID:           *m.SortID,
			Nummer:           *m.Nummer,
			Bezeichnung:      m.Bezeichnung,
			BezeichnungLang:  m.BezeichnungLang,
			Zusatz:           m.Zusatz,
			Leichtgewicht:    m.Leichtgewicht,
			Geschlecht:       m.Geschlecht,
			Bootsklasse:      m.Bootsklasse,
			BootsklasseLang:  m.BootsklasseLang,
			Altersklasse:     m.Altersklasse,
			AltersklasseLang: m.AltersklasseLang,
			Tag:              sqlc.Tag(m.Tag.Tag),
			Wettkampf:        sqlc.Wettkampf(m.Wettkampf.Wettkampf),
			KostenEur:        m.KostenEur,
			Rennabstand:      m.Rennabstand,
			Startzeit:        m.Startzeit,
		}
		meldungStruct := &sqlc.Meldung{}
		if m.Uuid_2 != uuid.Nil {
			meldungStruct = &sqlc.Meldung{
				Uuid:               m.Uuid_2,
				DrvRevisionUuid:    m.DrvRevisionUuid,
				Typ:                *m.Typ,
				Bemerkung:          m.Bemerkung,
				Abgemeldet:         m.Abgemeldet,
				Dns:                m.Dns,
				Dnf:                m.Dnf,
				Dsq:                m.Dsq,
				ZeitnahmeBemerkung: m.ZeitnahmeBemerkung,
				StartNummer:        m.StartNummer,
				Abteilung:          m.Abteilung,
				Bahn:               m.Bahn,
				Kosten:             *m.Kosten,
				VereinUuid:         m.VereinUuid,
				RennenUuid:         m.RennenUuid,
			}
		}

		if len(rLs) == 0 && currentRennen == nil {
			if m.Uuid_2 != uuid.Nil {
				currentRennen = &RennenWithMeldung{
					Rennen:    rennenStruct,
					Meldungen: []*sqlc.Meldung{meldungStruct},
				}
			} else {
				currentRennen = &RennenWithMeldung{
					Rennen:    rennenStruct,
					Meldungen: []*sqlc.Meldung{},
				}
			}
		}

		if currentRennen.Rennen.Uuid != rennenStruct.Uuid {
			if len(currentRennen.Meldungen) > 0 || getEmptyRennen {
				rLs = append(rLs, currentRennen)
			}

			if m.Uuid_2 != uuid.Nil {
				currentRennen = &RennenWithMeldung{
					Rennen:    rennenStruct,
					Meldungen: []*sqlc.Meldung{meldungStruct},
				}
			} else {
				currentRennen = &RennenWithMeldung{
					Rennen:    rennenStruct,
					Meldungen: []*sqlc.Meldung{},
				}
			}
		} else if currentRennen.Rennen.Uuid == m.Uuid {
			if m.Uuid_2 != uuid.Nil {
				currentRennen.Meldungen = append(currentRennen.Meldungen, meldungStruct)
			}
		} else {
			return nil, errors.New("This error should be happening!")
		}
	}

	if rLs == nil {
		rLs = []*RennenWithMeldung{}
	}

	for i, r := range rLs {
		rLs[i].NumMeldungen = len(r.Meldungen)

		maxAbt := 0
		if len(r.Meldungen) != 0 {
			for _, m := range r.Meldungen {
				if maxAbt < int(*m.Abteilung) {
					maxAbt = int(*m.Abteilung)
				}
			}
		}
		rLs[i].NumAbteilungen = maxAbt

		slices.SortFunc(r.Meldungen, func(a, b *sqlc.Meldung) int {
			return cmp.Or(
				cmp.Compare(*a.Abteilung, *b.Abteilung),
				cmp.Compare(*a.Bahn, *b.Bahn),
			)
		})

	}

	return rLs, nil
}

func GetAllRennenWithMeld(getEmptyRennen bool) ([]*RennenWithMeldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetAllRennenWithMeld(ctx)
	if err != nil {
		return nil, err
	}

	rLs, err := sqlcRennenToCrudRennen(q, getEmptyRennen)
	if err != nil {
		return nil, err
	}

	return rLs, nil
}

func GetAllRennenByWettkampf(wettkampf sqlc.Wettkampf, showStarted, showEmpty bool) ([]*Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetAllRennenByWettkampf(ctx, wettkampf)
	if err != nil {
		return nil, err
	}

	qParsed := []*sqlc.GetAllRennenWithMeldRow{}
	for _, r := range q {
		qParsed = append(qParsed, &sqlc.GetAllRennenWithMeldRow{
			Uuid:               r.Uuid,
			SortID:             r.SortID,
			Nummer:             r.Nummer,
			Bezeichnung:        r.Bezeichnung,
			BezeichnungLang:    r.BezeichnungLang,
			Zusatz:             r.Zusatz,
			Leichtgewicht:      r.Leichtgewicht,
			Geschlecht:         r.Geschlecht,
			Bootsklasse:        r.Bootsklasse,
			BootsklasseLang:    r.BootsklasseLang,
			Altersklasse:       r.Altersklasse,
			AltersklasseLang:   r.AltersklasseLang,
			Tag:                r.Tag,
			Wettkampf:          r.Wettkampf,
			KostenEur:          r.KostenEur,
			Rennabstand:        r.Rennabstand,
			Startzeit:          r.Startzeit,
			Uuid_2:             r.Uuid_2,
			DrvRevisionUuid:    r.DrvRevisionUuid,
			Typ:                r.Typ,
			Bemerkung:          r.Bemerkung,
			Abgemeldet:         r.Abgemeldet,
			Dns:                r.Dns,
			Dnf:                r.Dnf,
			Dsq:                r.Dsq,
			ZeitnahmeBemerkung: r.ZeitnahmeBemerkung,
			StartNummer:        r.StartNummer,
			Abteilung:          r.Abteilung,
			Bahn:               r.Bahn,
			Kosten:             r.Kosten,
			VereinUuid:         r.VereinUuid,
			RennenUuid:         r.RennenUuid,
		})
	}

	rLs, err := sqlcRennenToCrudRennen(qParsed, showEmpty)
	if err != nil {
		return nil, err
	}

	rennenLs := []*Rennen{}
	for _, r := range rLs {
		rennenLs = append(rennenLs, &Rennen{
			Uuid:             r.Uuid,
			SortID:           r.SortID,
			Nummer:           r.Nummer,
			Bezeichnung:      r.Bezeichnung,
			BezeichnungLang:  r.BezeichnungLang,
			Zusatz:           r.Zusatz,
			Leichtgewicht:    r.Leichtgewicht,
			Geschlecht:       r.Geschlecht,
			Bootsklasse:      r.Bootsklasse,
			BootsklasseLang:  r.BootsklasseLang,
			Altersklasse:     r.Altersklasse,
			AltersklasseLang: r.AltersklasseLang,
			Tag:              r.Tag,
			Wettkampf:        r.Wettkampf,
			KostenEur:        r.KostenEur,
			Rennabstand:      r.Rennabstand,
			Startzeit:        r.Startzeit,
			NumMeldungen:     r.NumMeldungen,
			NumAbteilungen:   r.NumAbteilungen,
		})
	}

	return rennenLs, nil
}

func GetRennenMinimal(uuid uuid.UUID) (*sqlc.Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	r, err := DB.Queries.GetRennenMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return nil, &api.NOT_FOUND
		}
		return nil, err
	}

	return r, nil
}

func UpdateStartZeit(params sqlc.UpdateStartZeitParams) error {
  ctx, cancel := getCtxWithTo()
  defer cancel()
  
  return DB.Queries.UpdateStartZeit(ctx, params)
}

func CreateRennen(rParams sqlc.CreateRennenParams) (*sqlc.Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.CreateRennen(ctx, rParams)
	if err != nil {
		return nil, err
	}

	return v, nil
}
