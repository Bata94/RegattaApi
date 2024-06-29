package crud

import (
	"errors"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
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

type RennenWithMeldung struct {
	*sqlc.Rennen
	Meldungen []*sqlc.Meldung
}

func GetAllRennenWithMeld(getEmptyRennen bool) ([]*RennenWithMeldung, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	log.Debug("Show Empty Renn: ", getEmptyRennen)

	q, err := DB.Queries.GetAllRennenWithMeld(ctx)
	if err != nil {
		return nil, err
	}

	var (
		rLs           []*RennenWithMeldung
		currentRennen *RennenWithMeldung
	)

	for _, m := range q {
		log.Debug(*m.SortID)
		rennenStruct := &sqlc.Rennen{
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

		// WIP!!!

		if len(rLs) == 0 || currentRennen.Rennen.Uuid != rennenStruct.Uuid {
			if currentRennen != nil {
				log.Debug("Append Rennen")
				rLs = append(rLs, currentRennen)
			}
			log.Debug("Set new Rennen")

			if m.Uuid_2 != uuid.Nil {
				currentRennen = &RennenWithMeldung{
					Rennen:    rennenStruct,
					Meldungen: []*sqlc.Meldung{meldungStruct},
				}
			} else {
				if getEmptyRennen {
					currentRennen = &RennenWithMeldung{
						Rennen:    rennenStruct,
						Meldungen: []*sqlc.Meldung{},
					}
				}
			}
		} else if currentRennen.Rennen.Uuid == m.Uuid {
			log.Debug("Append Meld")
			currentRennen.Meldungen = append(currentRennen.Meldungen, meldungStruct)
		} else {
			return nil, errors.New("This error should be happening!")
		}
	}

	return rLs, nil
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

func CreateRennen(rParams sqlc.CreateRennenParams) (*sqlc.Rennen, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.CreateRennen(ctx, rParams)
	if err != nil {
		return nil, err
	}

	return v, nil
}
