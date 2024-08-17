package crud

import (
	"errors"
	"strconv"

	"github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type Verein struct {
	sqlc.Verein
	Athleten []Athlet `json:"athleten"`
}

func (verein *Verein) GetRechnungsnummern() ([]string, error) {
	ctx, cancel := getCtxWithTo()
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

func (verein *Verein) GetNextRechnungsnummer() (string, error) {
	rechnungsNummern, err := verein.GetRechnungsnummern()
	if err != nil {
		return "", err
	}
	fwdNr := 0

	log.Debug(len(rechnungsNummern))
	if len(rechnungsNummern) != 0 {
		for _, r := range rechnungsNummern {
			l := len(r)
			rNrStr := r[l-3 : l]
			log.Debug(rNrStr)

			rNr, err := strconv.Atoi(rNrStr)
			if err != nil {
				return "", err
			}

			if fwdNr < rNr {
				fwdNr = rNr
			}
		}

		if fwdNr == 0 {
			return "", errors.New("Fehler beim erzeugen der neuen RechnungsNummer...")
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

func GetAllVerein() ([]Verein, error) {
	ctx, cancel := getCtxWithTo()
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

func GetVereinMinimal(uuid uuid.UUID) (Verein, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.GetVereinMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return Verein{}, &api.NOT_FOUND
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
	vLs, err := GetAllVerein()
	if err != nil {
		return vLs, err
	}
	retLs := []Verein{}

	var queryFunc func(uuid.UUID) ([]Athlet, error)
	if athletType == 0 {
		queryFunc = GetAllAthletenForVereinWaage
	} else if athletType == 1 {
		queryFunc = GetAllAthletenForVereinMissStartber
	} else {
		return vLs, errors.New("Unknown athlet type")
	}

	for _, v := range vLs {
		missAthlet, err := queryFunc(v.Uuid)
		if err != nil {
			return vLs, err
		}

		if missAthlet != nil && len(missAthlet) != 0 {
			v.Athleten = missAthlet
			retLs = append(retLs, v)
		}
	}

	return retLs, nil
}

func CreateVerein(vParams sqlc.CreateVereinParams) (Verein, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.CreateVerein(ctx, vParams)
	if err != nil {
		return Verein{}, err
	}

	return Verein{Verein: v}, nil
}
