package crud

import (
	"errors"
	"time"

	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Zeitnahme struct {
	ID              int32      `json:"id"`
	RennenNummer    *string    `json:"rennen_nummer"`
	StartNummer     *string    `json:"start_nummer"`
	TimeClient      *time.Time `json:"time_client"`
	TimeServer      *time.Time `json:"time_server"`
	MeasuredLatency *int       `json:"measured_latency"`
	Verarbeitet     bool       `json:"verarbeitet"`
}

func SqlcZeitnahmeStartToZeitnahme(z sqlc.ZeitnahmeStart) Zeitnahme {
	var rennenNummer, startNummer *string
	var timeClient, timeServer *time.Time
	var measuredLatency *int

	if z.RennenNummer.Valid {
		rennenNummer = &z.RennenNummer.String
	}
	if z.StartNummer.Valid {
		startNummer = &z.StartNummer.String
	}

	if z.TimeClient.Valid {
		timeClient = &z.TimeClient.Time
	}
	if z.TimeServer.Valid {
		timeServer = &z.TimeServer.Time
	}

	if z.MeasuredLatency.Valid {
		measuredLatencyVal := int(z.MeasuredLatency.Int32)
		measuredLatency = &measuredLatencyVal
	}

	return Zeitnahme{
		ID:              z.ID,
		RennenNummer:    rennenNummer,
		StartNummer:     startNummer,
		TimeClient:      timeClient,
		TimeServer:      timeServer,
		MeasuredLatency: measuredLatency,
		Verarbeitet:     false,
	}
}

func SqlcZeitnahmeZielToZeitnahme(z sqlc.ZeitnahmeZiel) Zeitnahme {
	var rennenNummer, startNummer *string
	var timeClient, timeServer *time.Time
	var measuredLatency *int

	if z.RennenNummer.Valid {
		rennenNummer = &z.RennenNummer.String
	}
	if z.StartNummer.Valid {
		startNummer = &z.StartNummer.String
	}

	if z.TimeClient.Valid {
		timeClient = &z.TimeClient.Time
	}
	if z.TimeServer.Valid {
		timeServer = &z.TimeServer.Time
	}

	if z.MeasuredLatency.Valid {
		measuredLatencyVal := int(z.MeasuredLatency.Int32)
		measuredLatency = &measuredLatencyVal
	}

	return Zeitnahme{
		ID:              z.ID,
		RennenNummer:    rennenNummer,
		StartNummer:     startNummer,
		TimeClient:      timeClient,
		TimeServer:      timeServer,
		MeasuredLatency: measuredLatency,
		Verarbeitet:     false,
	}
}

func GetOpenZeitnahmeStart() ([]Zeitnahme, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()
	retLs := []Zeitnahme{}

	q, err := DB.Queries.GetOpenStarts(ctx)
	if err != nil {
		return retLs, err
	}

	for _, z := range q {
		retLs = append(retLs, SqlcZeitnahmeStartToZeitnahme(z))
	}

	return retLs, nil
}

func GetZeitnahmeZiel(id int) (Zeitnahme, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	idI32 := int32(id)

	q, err := DB.Queries.GetZeitnahmeZiel(ctx, idI32)
	if err != nil {
		return Zeitnahme{}, err
	}

	return SqlcZeitnahmeZielToZeitnahme(q), nil
}

func GetOpenZeitnahmeZiel() ([]Zeitnahme, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.GetAllOpenZeitnahmeZiel(ctx)
	if err != nil {
		return []Zeitnahme{}, err
	}

	retLs := []Zeitnahme{}
	for _, z := range q {
		retLs = append(retLs, SqlcZeitnahmeZielToZeitnahme(z))
	}

	return retLs, nil
}

func CreateZeitnahmeStart(rennNr *string, startNummern []string, timeClient time.Time, measuredLatency int) ([]Zeitnahme, error) {
	now := time.Now()
	retLs := []Zeitnahme{}

	ctx, cancel := getCtxWithTo()
	defer cancel()

	var rennenNummer, startNummer pgtype.Text
	if rennNr != nil {
		rennenNummer = pgtype.Text{String: *rennNr, Valid: true}
	} else {
		rennenNummer = pgtype.Text{Valid: false}
	}

	if startNummern == nil {
		return retLs, errors.New("startNummern is nil")
	}

	for _, startNr := range startNummern {
		startNummer = pgtype.Text{String: startNr, Valid: true}

		p := sqlc.CreateZeitnahmeStartParams{
			RennenNummer: rennenNummer,
			StartNummer:  startNummer,
			TimeClient: pgtype.Timestamp{
				Valid: true,
				Time:  timeClient,
			},
			TimeServer: pgtype.Timestamp{
				Valid: true,
				Time:  now,
			},
			MeasuredLatency: pgtype.Int4{
				Valid: true,
				Int32: int32(measuredLatency),
			},
		}

		q, err := DB.Queries.CreateZeitnahmeStart(ctx, p)
		if err != nil {
			return retLs, err
		}
		retLs = append(retLs, SqlcZeitnahmeStartToZeitnahme(q))
	}

	return retLs, nil
}

func CreateZeitnahmeZiel(rennNr, startNr *string, timeClient time.Time, measuredLatency int) (Zeitnahme, error) {
	now := time.Now()

	ctx, cancel := getCtxWithTo()
	defer cancel()

	var rennenNummer, startNummer pgtype.Text
	if rennNr != nil {
		rennenNummer = pgtype.Text{String: *rennNr, Valid: true}
	} else {
		rennenNummer = pgtype.Text{Valid: false}
	}

	if startNr != nil {
		startNummer = pgtype.Text{String: *startNr, Valid: true}
	} else {
		startNummer = pgtype.Text{Valid: false}
	}

	p := sqlc.CreateZeitnahmeZielParams{
		RennenNummer: rennenNummer,
		StartNummer:  startNummer,
		TimeClient: pgtype.Timestamp{
			Valid: true,
			Time:  timeClient,
		},
		TimeServer: pgtype.Timestamp{
			Valid: true,
			Time:  now,
		},
		MeasuredLatency: pgtype.Int4{
			Valid: true,
			Int32: int32(measuredLatency),
		},
	}

	q, err := DB.Queries.CreateZeitnahmeZiel(ctx, p)
	if err != nil {
		return Zeitnahme{}, err
	}

	return SqlcZeitnahmeZielToZeitnahme(q), nil
}

func UpdateZeitnahmeZiel(z Zeitnahme, rennNr, startNr *string) (Zeitnahme, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	rennenNummer, startNummer := pgtype.Text{Valid: false}, pgtype.Text{Valid: false}
	if rennNr != nil {
		rennenNummer = pgtype.Text{String: *rennNr, Valid: true}
	} else if z.RennenNummer != nil {
		rennenNummer = pgtype.Text{String: *z.RennenNummer, Valid: true}
	}

	if startNr != nil {
		startNummer = pgtype.Text{String: *startNr, Valid: true}
	} else if z.StartNummer != nil {
		startNummer = pgtype.Text{String: *z.StartNummer, Valid: true}
	}

	p := sqlc.UpdateZeitnahmeZielParams{
		ID:           z.ID,
		RennenNummer: rennenNummer,
		StartNummer:  startNummer,
	}

	q, err := DB.Queries.UpdateZeitnahmeZiel(ctx, p)
	if err != nil {
		return Zeitnahme{}, err
	}

	return SqlcZeitnahmeZielToZeitnahme(q), nil
}

func DeleteZeitnahmeZiel(z Zeitnahme) (Zeitnahme, error) {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	q, err := DB.Queries.DeleteZeitnahmeZiel(ctx, z.ID)
	if err != nil {
		return Zeitnahme{}, err
	}

	return SqlcZeitnahmeZielToZeitnahme(q), nil
}

func CreateZeitnahmeErgebnis(s, z Zeitnahme, meld Meldung) error {
	ctx, cancel := getCtxWithTo()
	defer cancel()

	endZeit := z.TimeClient.Sub(*s.TimeClient)

	params := sqlc.CreateZeitnahmeErgebnisParams{
		Endzeit:          endZeit.Seconds(),
		ZeitnahmeStartID: s.ID,
		ZeitnahmeZielID:  z.ID,
		MeldungUuid:      meld.Uuid,
	}

	q, err := DB.Queries.CreateZeitnahmeErgebnis(ctx, params)
	if err != nil {
		return err
	}

	err = DB.Queries.SetZeitnahmeStartVerarbeitet(ctx, s.ID)
	if err != nil {
		return err
	}
	err = DB.Queries.SetZeitnahmeZielVerarbeitet(ctx, z.ID)
	if err != nil {
		return err
	}

	log.Debug(q)

	return nil
}

func GetZeitnahmeErgebnisByMeld(meldUuid uuid.UUID) (sqlc.ZeitnahmeErgebni, error) {
  ctx, cancel := getCtxWithTo()
  defer cancel()
  
  return DB.Queries.GetZeitnahmeErgebnisByMeld(ctx, meldUuid)
}
