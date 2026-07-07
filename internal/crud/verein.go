package crud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
)

type Verein struct {
	sqlc.Verein
	Athleten     []Athlet `json:"athleten,omitempty"`
	GesKosten    *int     `json:"ges_kosten,omitempty"`
	GesZahlungen *int     `json:"ges_zahlungen,omitempty"`
	Saldo        *int     `json:"saldo,omitempty"`
}

type vereinJSON struct {
	Uuid         uuid.UUID `json:"uuid"`
	Kuerzel      string    `json:"kuerzel"`
	Name         string    `json:"name"`
	Kurzform     string    `json:"kurzform"`
	Athleten     []Athlet  `json:"athleten,omitempty"`
	GesKosten    *int      `json:"ges_kosten,omitempty"`
	GesZahlungen *int      `json:"ges_zahlungen,omitempty"`
	Saldo        *int      `json:"saldo,omitempty"`
}

func (v Verein) MarshalJSON() ([]byte, error) {
	j := vereinJSON{
		Uuid:         v.Uuid,
		Kuerzel:      v.Kuerzel,
		Name:         v.Name,
		Kurzform:     v.Kurzform,
		Athleten:     v.Athleten,
		GesKosten:    v.GesKosten,
		GesZahlungen: v.GesZahlungen,
		Saldo:        v.Saldo,
	}
	return json.Marshal(j)
}

func (v *Verein) GetAthleten() ([]Athlet, error) {
	if v.Athleten != nil {
		return v.Athleten, nil
	}
	return v.loadAthleten(context.Background())
}

func (v *Verein) loadAthleten(ctx context.Context) ([]Athlet, error) {
	loaded, err := GetAllAthletenForVerein(ctx, v.Uuid)
	if err != nil {
		return nil, err
	}
	v.Athleten = loaded
	return v.Athleten, nil
}

func VereinFromSqlc(v sqlc.Verein, gesKosten int) Verein {
	return Verein{
		Verein:       v,
		Athleten:     make([]Athlet, 0),
		GesKosten:    &gesKosten,
		GesZahlungen: &gesKosten,
		Saldo:        &gesKosten,
	}
}

func (verein *Verein) GetRechnungsnummern(ctx context.Context) ([]string, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	retLs := []string{}

	q, err := DB.Queries.GetVereinRechnungsnummern(ctx, verein.Uuid)
	if err != nil || len(q) == 0 {
		return retLs, err
	}

	for _, r := range q {
		if r.Valid {
			retLs = append(retLs, r.String)
		}
	}

	return retLs, nil
}

func (verein *Verein) GetNextRechnungsnummer(ctx context.Context) (string, error) {
	rechnungsNummern, err := verein.GetRechnungsnummern(ctx)
	if err != nil {
		return "", err
	}
	fwdNr := 0

	slog.Debug(fmt.Sprintf("rechnungsNummern count: %d", len(rechnungsNummern)))
	if len(rechnungsNummern) != 0 {
		for _, r := range rechnungsNummern {
			l := len(r)
			rNrStr := r[l-3 : l]
			slog.Debug(fmt.Sprintf("rechnungsNummer: %s", rNrStr))

			rNr, err := strconv.Atoi(rNrStr)
			if err != nil {
				return "", err
			}

			if fwdNr < rNr {
				fwdNr = rNr
			}
		}

		if fwdNr == 0 {
			return "", errors.New("fehler beim erzeugen der neuen rechnungsnummer")
		}

		fwdNr += 1
	} else {
		fwdNr = 1
	}

	fwdNrStr := strconv.Itoa(fwdNr)

	for len(fwdNrStr) < 3 {
		fwdNrStr = "0" + fwdNrStr
	}

	reNr := "2024-" + verein.Kuerzel + "-" + fwdNrStr
	return reNr, nil
}

func GetAllVerein(ctx context.Context) ([]Verein, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.Queries.GetAllVerein(ctx)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return []Verein{}, nil
	}

	vLs := []Verein{}
	for _, i := range q {
		vLs = append(vLs, Verein{
			Verein: i,
		})
	}

	return vLs, err
}

func GetVerein(ctx context.Context, uuid uuid.UUID) (Verein, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.Queries.GetVerein(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Verein{}, apierr.ErrNotFound
		}
		return Verein{}, err
	}

	gesKostenI64, ok := q.GesKosten.(int64)
	if !ok {
		return Verein{}, errors.New("error while converting gesamt kosten to int")
	}
	gesZahlungenI64, ok := q.GesZahlungen.(int64)
	if !ok {
		return Verein{}, errors.New("error while converting gesamt zahlung to int")
	}
	gesKosten := int(gesKostenI64)
	gesZahlungen := int(gesZahlungenI64)
	saldo := gesZahlungen - gesKosten

	return Verein{
		Verein:       q.Verein,
		GesKosten:    &gesKosten,
		GesZahlungen: &gesZahlungen,
		Saldo:        &saldo,
	}, nil
}

func GetVereinMinimal(ctx context.Context, uuid uuid.UUID) (Verein, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	v, err := DB.Queries.GetVereinMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Verein{}, apierr.ErrNotFound
		}
		return Verein{}, err
	}

	return Verein{Verein: v}, nil
}

type MissingAthletType int

const (
	Waage           MissingAthletType = 0
	Startberechtigt MissingAthletType = 1
)

func GetForAllVereineMissingAthlet(athletType MissingAthletType) ([]Verein, error) {
	vLs, err := GetAllVerein(context.Background())
	if err != nil {
		return vLs, err
	}
	retLs := []Verein{}

	var queryFunc func(context.Context, uuid.UUID) ([]Athlet, error)
	switch athletType {
	case Waage:
		queryFunc = GetAllAthletenForVereinWaage
	case Startberechtigt:
		queryFunc = GetAllAthletenForVereinMissStartber
	default:
		return vLs, errors.New("unknown athlet type")
	}

	for _, v := range vLs {
		missAthlet, err := queryFunc(context.Background(), v.Uuid)
		if err != nil {
			return vLs, err
		}

		if len(missAthlet) != 0 {
			v.Athleten = missAthlet
			retLs = append(retLs, v)
		}
	}

	return retLs, nil
}

func CreateVerein(ctx context.Context, vParams sqlc.CreateVereinParams) (Verein, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	v, err := DB.Queries.CreateVerein(ctx, vParams)
	if err != nil {
		return Verein{}, err
	}

	return Verein{Verein: v}, nil
}
