package crud

import (
	"fmt"
	"time"

	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type SetZeitplanParams struct {
	SaStartStunde int `json:"sa_start_stunde"`
	SoStartStunde int `json:"so_start_stunde"`
}

func SetZeitplan(param SetZeitplanParams) error {
	rLs, err := GetAllRennen(GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	pLs, err := GetAllPausen()
	if err != nil {
		return err
	}

	curStartTimeSa, err:= time.Parse("15:04", fmt.Sprintf("%d:00", param.SaStartStunde))
	if err != nil {
		return err
	}

	curStartTimeSo, err := time.Parse("15:04", fmt.Sprintf("%d:00", param.SoStartStunde))
	if err != nil {
		return err
	}

	fmt.Println("curStartTimeSa:", curStartTimeSa)
	fmt.Println("curStartTimeSo:", curStartTimeSo)

	for _, r := range rLs {
		switch r.Tag {
		case sqlc.TagSa:
			saTimeStr := curStartTimeSa.Format("15:04")
			fmt.Printf("Setting RennenNr: %s to time %s\n", r.Nummer, saTimeStr)
			err := UpdateStartZeit(sqlc.UpdateStartZeitParams{
				Uuid:      r.Uuid,
				Startzeit: pgtype.Text{String: saTimeStr, Valid: true},
			})
			if err != nil {
				return err
			}
			if r.Wettkampf == sqlc.WettkampfLangstrecke {
				curStartTimeSa = curStartTimeSa.Add(time.Minute * time.Duration(r.Rennabstand * *r.NumMeldungen))
			} else {
				curStartTimeSa = curStartTimeSa.Add(time.Minute * time.Duration(r.Rennabstand * *r.NumAbteilungen))
			}

			for _, p := range pLs {
				if p.NachRennenUuid == r.Uuid {
					curStartTimeSa = curStartTimeSa.Add(time.Minute * time.Duration(p.Laenge))
				}
			}
		case sqlc.TagSo:
			soTimeStr := curStartTimeSo.Format("15:04")
			fmt.Printf("Setting RennenNr: %s to time %s\n", r.Nummer, soTimeStr)
			err := UpdateStartZeit(sqlc.UpdateStartZeitParams{
				Uuid:      r.Uuid,
				Startzeit: pgtype.Text{String: soTimeStr, Valid: true},
			})
			if err != nil {
				return err
			}
			curStartTimeSo = curStartTimeSo.Add(time.Minute * time.Duration(r.Rennabstand * *r.NumAbteilungen))

			for _, p := range pLs {
				if p.NachRennenUuid == r.Uuid {
					curStartTimeSo = curStartTimeSo.Add(time.Minute * time.Duration(p.Laenge))
				}
			}
		}
	}

	return nil
}
