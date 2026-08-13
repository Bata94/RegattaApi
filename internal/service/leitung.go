package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type SetZeitplanParams struct {
	SaStartStunde int `json:"sa_start_stunde"`
	SoStartStunde int `json:"so_start_stunde"`
}

func SetZeitplan(ctx context.Context, param SetZeitplanParams) error {
	rLs, err := crud.GetAllRennen(ctx, crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	pLs, err := crud.GetAllPausen(ctx)
	if err != nil {
		return err
	}

	curStartTimeSa, err := time.Parse("15:04", fmt.Sprintf("%d:00", param.SaStartStunde))
	if err != nil {
		return err
	}

	curStartTimeSo, err := time.Parse("15:04", fmt.Sprintf("%d:00", param.SoStartStunde))
	if err != nil {
		return err
	}

	slog.Debug("Zeitplan start time", "sa", curStartTimeSa, "so", curStartTimeSo)

	for _, r := range rLs {
		rennAbstand := 10
		if v := r.GetRennabstand(); v != nil {
			rennAbstand = *v
		}

		switch r.Tag {
		case crud.TagSa:
			saTimeStr := curStartTimeSa.Format("15:04")
			slog.Debug("Setting rennen start time", "nummer", r.Nummer, "time", saTimeStr)
			err := crud.UpdateStartZeit(ctx, sqlc.UpdateStartZeitParams{
				Uuid:      r.Uuid,
				Startzeit: pgtype.Text{String: saTimeStr, Valid: true},
			})
			if err != nil {
				return err
			}
			if r.Wettkampf == sqlc.WettkampfLangstrecke {
				curStartTimeSa = curStartTimeSa.Add(time.Minute * time.Duration(rennAbstand**r.NumMeldungen))
			} else {
				curStartTimeSa = curStartTimeSa.Add(time.Minute * time.Duration(rennAbstand**r.NumAbteilungen))
			}

			for _, p := range pLs {
				if p.NachRennenUuid == r.Uuid {
					curStartTimeSa = curStartTimeSa.Add(time.Minute * time.Duration(p.Laenge))
				}
			}
		case crud.TagSo:
			soTimeStr := curStartTimeSo.Format("15:04")
			slog.Debug("Setting rennen start time", "nummer", r.Nummer, "time", soTimeStr)
			err := crud.UpdateStartZeit(ctx, sqlc.UpdateStartZeitParams{
				Uuid:      r.Uuid,
				Startzeit: pgtype.Text{String: soTimeStr, Valid: true},
			})
			if err != nil {
				return err
			}
			curStartTimeSo = curStartTimeSo.Add(time.Minute * time.Duration(rennAbstand**r.NumAbteilungen))

			for _, p := range pLs {
				if p.NachRennenUuid == r.Uuid {
					curStartTimeSo = curStartTimeSo.Add(time.Minute * time.Duration(p.Laenge))
				}
			}
		}
	}

	return nil
}

func SetStartnummern(ctx context.Context) error {
	check, err := crud.CheckMeldungSetzung(ctx)
	if err != nil {
		return err
	}
	if !check {
		return handler.BadRequest("Setzung not done!")
	}

	check2, err := crud.CheckMeldungStartnummern(ctx)
	if err != nil {
		return err
	}
	if check2 {
		return handler.BadRequest("Startnummern not done!")
	}

	rLs, err := crud.GetAllRennen(ctx, crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		return err
	}

	startNummerMap := make(map[crud.Tag]int32)
	startNummerMap[crud.TagSa] = 1
	startNummerMap[crud.TagSo] = 1

	for _, r := range rLs {
		meldungen, err := r.GetMeldungen(ctx)
		if err != nil {
			return err
		}
		for _, m := range meldungen {
			if m.Abgemeldet {
				continue
			}
			err = crud.UpdateStartNummer(ctx, sqlc.UpdateStartNummerParams{
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

func ResetStartnummern(ctx context.Context) error {
	mLs, err := crud.GetAllMeldungen(ctx)
	if err != nil {
		return err
	}

	for _, m := range mLs {
		err = crud.UpdateStartNummer(ctx, sqlc.UpdateStartNummerParams{
			Uuid:        m.Uuid,
			StartNummer: 0,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
