package crud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Zeitnahme struct {
	ID              int32      `json:"id"`
	RennenNummer    *string    `json:"rennen_nummer,omitempty"`
	StartNummer     *string    `json:"start_nummer,omitempty"`
	TimeClient      *time.Time `json:"time_client,omitempty"`
	TimeServer      *time.Time `json:"time_server,omitempty"`
	MeasuredLatency *int       `json:"measured_latency,omitempty"`
	Verarbeitet     bool       `json:"verarbeitet"`
}

func (z Zeitnahme) MarshalJSON() ([]byte, error) {
	type alias Zeitnahme
	return json.Marshal(alias(z))
}

func ZeitnahmeFromSqlcStart(z sqlc.ZeitnahmeStart) Zeitnahme {
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

func ZeitnahmeFromSqlcZiel(z sqlc.ZeitnahmeZiel) Zeitnahme {
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

func GetOpenZeitnahmeStart(ctx context.Context) ([]Zeitnahme, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()
	retLs := []Zeitnahme{}

	q, err := DB.Queries.GetOpenStarts(ctx)
	if err != nil {
		return retLs, err
	}

	for _, z := range q {
		retLs = append(retLs, ZeitnahmeFromSqlcStart(z))
	}

	return retLs, nil
}

func GetZeitnahmeZiel(ctx context.Context, id int) (Zeitnahme, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	idI32 := int32(id)

	q, err := DB.Queries.GetZeitnahmeZiel(ctx, idI32)
	if err != nil {
		return Zeitnahme{}, err
	}

	return ZeitnahmeFromSqlcZiel(q), nil
}

func GetOpenZeitnahmeZiel(ctx context.Context) ([]Zeitnahme, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.Queries.GetAllOpenZeitnahmeZiel(ctx)
	if err != nil {
		return []Zeitnahme{}, err
	}

	retLs := []Zeitnahme{}
	for _, z := range q {
		retLs = append(retLs, ZeitnahmeFromSqlcZiel(z))
	}

	return retLs, nil
}

func CreateZeitnahmeStart(ctx context.Context, rennNr *string, startNummern []string, timeClient time.Time, measuredLatency int) ([]Zeitnahme, error) {
	now := time.Now()
	retLs := []Zeitnahme{}

	ctx, cancel := getCtx(ctx)
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
		retLs = append(retLs, ZeitnahmeFromSqlcStart(q))
	}

	return retLs, nil
}

func CreateZeitnahmeZiel(ctx context.Context, rennNr, startNr *string, timeClient time.Time, measuredLatency int) (Zeitnahme, error) {
	now := time.Now()

	ctx, cancel := getCtx(ctx)
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

	return ZeitnahmeFromSqlcZiel(q), nil
}

func CreateZeitnahmeErgebnis(ctx context.Context, s, z Zeitnahme, meld Meldung) error {
	ctx, cancel := getCtx(ctx)
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

	slog.Debug(fmt.Sprintf("%v", q))

	return nil
}

func GetZeitnahmeErgebnisByMeld(ctx context.Context, meldUuid uuid.UUID) (sqlc.ZeitnahmeErgebni, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.Queries.GetZeitnahmeErgebnisByMeld(ctx, meldUuid)
}
