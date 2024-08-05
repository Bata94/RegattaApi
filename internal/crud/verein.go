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

func GetAllVerein() ([]sqlc.Verein, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	vLs, err := DB.Queries.GetAllVerein(ctx)
	if err != nil {
		return nil, err
	}
	if vLs == nil {
		vLs = []sqlc.Verein{}
	}

	return vLs, err
}

func GetVereinMinimal(uuid uuid.UUID) (sqlc.Verein, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.GetVereinMinimal(ctx, uuid)
	if err != nil {
		if isNoRowError(err) {
			return sqlc.Verein{}, &api.NOT_FOUND
		}
		return sqlc.Verein{}, err
	}

	return v, nil
}

func GetVereinRechnungsnummern(vereinUuid uuid.UUID) ([]string, error) {
  ctx, cancel := getCtxWithTo()
  defer cancel()

  retLs := []string{}

  q, err := DB.Queries.GetVereinRechnungsnummern(ctx, vereinUuid)
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

func GetVereinNextRechnungsnummer(v sqlc.Verein) (string, error) {
	rechnungsNummern, err := GetVereinRechnungsnummern(v.Uuid)
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

	reNr := "2024-" + v.Kuerzel + "-" + fwdNrStr
  return reNr, nil
}

func CreateVerein(vParams sqlc.CreateVereinParams) (sqlc.Verein, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	v, err := DB.Queries.CreateVerein(ctx, vParams)
	if err != nil {
		return sqlc.Verein{}, err
	}

	return v, nil
}
