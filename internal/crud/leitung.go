package crud

import (
	"fmt"
	"time"

	"github.com/bata94/RegattaApi/internal/handler"
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
		case TagSa:
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
		case TagSo:
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

func SetStartnummern() error {
	check, err := CheckMeldungSetzung()
	if err != nil {
		return err
	}
	if !check {
		return &handler.Error{StatusCode: 400, Message: "Setzung not done!"}
	}

	check2, err := CheckMeldungStartnummern()
	if err != nil {
		return err
	}
	if check2 {
		return &handler.Error{StatusCode: 400, Message: "Startnummern not done!"}
	}

	rLs, err := GetAllRennen(GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	startNummerMap := make(map[Tag]int32)
	startNummerMap[TagSa] = 1
	startNummerMap[TagSo] = 1

	for _, r := range rLs {
		for _, m := range r.Meldungen {
			if m.Abgemeldet {
				continue
			}
			err = UpdateStartNummer(sqlc.UpdateStartNummerParams{
				Uuid:        m.Uuid,
				StartNummer: startNummerMap[r.Tag],
			})
			if err != nil {
				return err
			}
			startNummerMap[r.Tag]++
		}
	}

	return nil
}

func ResetStartnummern() error {
	mLs, err := GetAllMeldungen()
	if err != nil {
		return err
	}

	for _, m := range mLs {
		err = UpdateStartNummer(sqlc.UpdateStartNummerParams{
			Uuid:        m.Uuid,
			StartNummer: 0,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
