package crud

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Athlet struct {
	sqlc.Athlet
	Rolle        *sqlc.Rolle `json:"rolle,omitempty"`
	Position     *int        `json:"position,omitempty"`
	Verein       *Verein     `json:"verein,omitempty"`
	Meldungen    []Meldung   `json:"meldungen,omitempty"`
	ErstesRennen *Rennen     `json:"erstes_rennen,omitempty"`
}

func AthletFromSqlc(a sqlc.Athlet) Athlet {
	return Athlet{Athlet: a}
}

func (a *Athlet) GetMeldungen() ([]Meldung, error) {
	if a.Meldungen != nil {
		return a.Meldungen, nil
	}
	return nil, nil
}

func (a *Athlet) GewichtStr() string {
	if a.Gewicht.Valid {
		return fmt.Sprintf("%d", a.Gewicht.Int32)
	}
	return ""
}

func (a *Athlet) GewichtInt() int {
	if a.Gewicht.Valid {
		return int(a.Gewicht.Int32)
	}
	return 0
}

type athletJSON struct {
	Uuid            uuid.UUID `json:"uuid"`
	Vorname         string    `json:"vorname"`
	Name            string    `json:"name"`
	Geschlecht      string    `json:"geschlecht"`
	Jahrgang        string    `json:"jahrgang"`
	Gewicht         *int      `json:"gewicht,omitempty"`
	Startberechtigt bool      `json:"startberechtigt"`
	VereinUuid      uuid.UUID `json:"verein_uuid"`
	Rolle           *string   `json:"rolle,omitempty"`
	Position        *int      `json:"position,omitempty"`
	Verein          *Verein   `json:"verein,omitempty"`
	Meldungen       []Meldung `json:"meldungen,omitempty"`
	ErstesRennen    *Rennen   `json:"erstes_rennen,omitempty"`
}

func (a Athlet) MarshalJSON() ([]byte, error) {
	j := athletJSON{
		Uuid:            a.Uuid,
		Vorname:         a.Vorname,
		Name:            a.Name,
		Geschlecht:      string(a.Geschlecht),
		Jahrgang:        a.Jahrgang,
		Startberechtigt: a.Startberechtigt,
		VereinUuid:      a.VereinUuid,
		Verein:          a.Verein,
		Meldungen:       a.Meldungen,
		ErstesRennen:    a.ErstesRennen,
	}
	if a.Rolle != nil {
		r := string(*a.Rolle)
		j.Rolle = &r
	}
	if a.Position != nil {
		j.Position = a.Position
	}
	if a.Gewicht.Valid {
		g := int(a.Gewicht.Int32)
		j.Gewicht = &g
	}
	return json.Marshal(j)
}

func (ath *Athlet) AthletString() string {
	return fmt.Sprintf("%s %s (%s)", ath.Vorname, ath.Name, ath.Jahrgang)
}

func (ath *Athlet) FullName() string {
	return fmt.Sprintf("%s %s", ath.Vorname, ath.Name)
}

func (ath *Athlet) UpdateStartberechtigung(startberechtigt bool) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	err := DB.Queries.UpdateAthletAerztlBesch(ctx, sqlc.UpdateAthletAerztlBeschParams{
		Startberechtigt: startberechtigt,
		Uuid:            ath.Uuid,
	})
	if err != nil {
		return err
	}

	ath.Startberechtigt = startberechtigt
	return nil
}

func (ath *Athlet) UpdateGewicht(gewichtParam int) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	gewicht := int32(gewichtParam)

	err := DB.Queries.UpdateAthletWaage(ctx, sqlc.UpdateAthletWaageParams{
		Gewicht: pgtype.Int4{Valid: true, Int32: gewicht},
		Uuid:    ath.Uuid,
	})
	if err != nil {
		return err
	}

	ath.Gewicht = pgtype.Int4{Valid: true, Int32: gewicht}
	return nil
}

func GetAthlet(uuid uuid.UUID) (Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	rows, err := DB.Queries.GetAthlet(ctx, uuid)
	if err != nil {
		return Athlet{}, err
	}
	if len(rows) == 0 {
		return Athlet{}, &api.NOT_FOUND
	}

	verein := VereinFromSqlc(rows[0].Verein, 0)
	a := Athlet{
		Athlet:    rows[0].Athlet,
		Verein:    &verein,
		Meldungen: []Meldung{},
	}

	for _, r := range rows {
		rennen := RennenFromSqlc(r.Rennen, 0, 0)
		a.Meldungen = append(a.Meldungen, Meldung{
			Meldung: r.Meldung,
			Rennen:  &rennen,
		})
	}

	return a, nil
}

func GetAthletMinimal(uuid uuid.UUID) (Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	a, err := DB.Queries.GetAthletMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Athlet{}, &api.NOT_FOUND
		}
		return Athlet{}, err
	}

	return Athlet{Athlet: a}, nil
}

func GetAllAthlet() ([]Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	aLs := []Athlet{}
	q, err := DB.Queries.GetAllAthlet(ctx)
	if err != nil {
		return nil, err
	}

	for _, a := range q {
		aLs = append(aLs, Athlet{
			Athlet: a,
		})
	}

	return aLs, err
}

func GetAllNNAthleten() ([]Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	aLs := []Athlet{}
	q, err := DB.Queries.GetAllNNAthleten(ctx)
	if err != nil {
		return nil, err
	}

	for _, a := range q {
		aLs = append(aLs, Athlet{
			Athlet: a,
		})
	}

	return aLs, err
}

func GetAllAthletenForVerein(vUuid uuid.UUID) ([]Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetAllAthletenForVerein(ctx, vUuid)
	if err != nil {
		return nil, err
	}

	retLs := []Athlet{}

	for _, a := range q {
		retLs = append(retLs, Athlet{
			Athlet: a,
		})
	}

	return retLs, nil
}

func GetAllAthletenForVereinWaage(vUuid uuid.UUID) ([]Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	retLs := []Athlet{}
	q, err := DB.Queries.GetAllAthletenForVereinWaage(ctx, vUuid)
	if err != nil {
		return retLs, err
	}

	for i, r := range q {
		if len(retLs) == 0 || (q[i-1].Athlet.Vorname != r.Athlet.Vorname && q[i-1].Athlet.Name != r.Athlet.Name) {
			rennen := RennenFromSqlc(r.Rennen, 0, 0)
			retLs = append(retLs, Athlet{
				Athlet:       r.Athlet,
				ErstesRennen: &rennen,
			})
			continue
		}
	}

	return retLs, nil
}

func GetAllAthletenForVereinMissStartber(vUuid uuid.UUID) ([]Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	retLs := []Athlet{}
	// TODO: klären, brauchen Stm. auch eine Startberechtigung? Brauchen Sie nicht need to change!
	q, err := DB.Queries.GetAllAthletenForVereinMissStartber(ctx, vUuid)
	if err != nil {
		return retLs, err
	}

	for i, r := range q {
		if len(retLs) == 0 || (q[i-1].Athlet.Vorname != r.Athlet.Vorname && q[i-1].Athlet.Name != r.Athlet.Name) {
			rennen := RennenFromSqlc(r.Rennen, 0, 0)
			retLs = append(retLs, Athlet{
				Athlet:       r.Athlet,
				ErstesRennen: &rennen,
			})
			continue
		}
	}

	return retLs, nil
}

func CreateAthlet(aParams sqlc.CreateAthletParams) (Athlet, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	a, err := DB.Queries.CreateAthlet(ctx, aParams)
	if err != nil {
		log.Println(err.Error())
		return Athlet{}, err
	}

	return Athlet{Athlet: a}, nil
}
