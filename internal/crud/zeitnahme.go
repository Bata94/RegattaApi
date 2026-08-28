package crud

import (
	"context"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"time"

	DB "github.com/bata94/RegattaApi/internal/db"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	return jsonv2.Marshal(alias(z))
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

	q, err := DB.QueriesFromCtx(ctx).GetOpenStarts(ctx)
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

	q, err := DB.QueriesFromCtx(ctx).GetZeitnahmeZiel(ctx, idI32)
	if err != nil {
		return Zeitnahme{}, err
	}

	return ZeitnahmeFromSqlcZiel(q), nil
}

func GetOpenZeitnahmeZiel(ctx context.Context) ([]Zeitnahme, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	q, err := DB.QueriesFromCtx(ctx).GetAllOpenZeitnahmeZiel(ctx)
	if err != nil {
		return []Zeitnahme{}, err
	}

	retLs := []Zeitnahme{}
	for _, z := range q {
		retLs = append(retLs, ZeitnahmeFromSqlcZiel(z))
	}

	return retLs, nil
}

func CreateZeitnahmeStart(ctx context.Context, rennNr *string, startNummern []string, timeClient time.Time, measuredLatency int, clientID, seq string) ([]Zeitnahme, error) {
	now := time.Now()
	retLs := []Zeitnahme{}

	ctx, cancel := getCtx(ctx)
	defer cancel()

	var rennenNummer pgtype.Text
	if rennNr != nil {
		rennenNummer = pgtype.Text{String: *rennNr, Valid: true}
	} else {
		rennenNummer = pgtype.Text{Valid: false}
	}

	if startNummern == nil {
		return retLs, errors.New("startNummern is nil")
	}

	clientIDText := pgtype.Text{String: clientID, Valid: true}
	seqText := pgtype.Text{String: seq, Valid: true}

	for _, startNr := range startNummern {
		startNummer := pgtype.Text{String: startNr, Valid: true}

		p := sqlc.CreateZeitnahmeStartParams{
			RennenNummer: rennenNummer,
			StartNummer:  startNummer,
			TimeClient: pgtype.Timestamp{
				Valid: true,
				Time:  timeClient.UTC(),
			},
			TimeServer: pgtype.Timestamp{
				Valid: true,
				Time:  now.UTC(),
			},
			MeasuredLatency: pgtype.Int4{
				Valid: true,
				Int32: int32(measuredLatency),
			},
			ClientID: clientIDText,
			Seq:      seqText,
		}

		q, err := DB.QueriesFromCtx(ctx).CreateZeitnahmeStart(ctx, p)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				existing, findErr := DB.QueriesFromCtx(ctx).FindZeitnahmeStartByClientSeq(ctx, sqlc.FindZeitnahmeStartByClientSeqParams{
					ClientID: clientIDText,
					Seq:      seqText,
				})
				if findErr != nil {
					return retLs, fmt.Errorf("conflict lookup failed: %w", findErr)
				}
				retLs = append(retLs, ZeitnahmeFromSqlcStart(existing))
				continue
			}
			return retLs, err
		}
		retLs = append(retLs, ZeitnahmeFromSqlcStart(q))
	}

	return retLs, nil
}

func CreateZeitnahmeZiel(ctx context.Context, rennNr, startNr *string, timeClient time.Time, measuredLatency int, clientID, seq string) (Zeitnahme, error) {
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

	clientIDText := pgtype.Text{String: clientID, Valid: true}
	seqText := pgtype.Text{String: seq, Valid: true}

	p := sqlc.CreateZeitnahmeZielParams{
		RennenNummer: rennenNummer,
		StartNummer:  startNummer,
		TimeClient: pgtype.Timestamp{
			Valid: true,
			Time:  timeClient.UTC(),
		},
		TimeServer: pgtype.Timestamp{
			Valid: true,
			Time:  now.UTC(),
		},
		MeasuredLatency: pgtype.Int4{
			Valid: true,
			Int32: int32(measuredLatency),
		},
		ClientID: clientIDText,
		Seq:      seqText,
	}

	q, err := DB.QueriesFromCtx(ctx).CreateZeitnahmeZiel(ctx, p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existing, findErr := DB.QueriesFromCtx(ctx).FindZeitnahmeZielByClientSeq(ctx, sqlc.FindZeitnahmeZielByClientSeqParams{
				ClientID: clientIDText,
				Seq:      seqText,
			})
			if findErr != nil {
				return Zeitnahme{}, fmt.Errorf("conflict lookup failed: %w", findErr)
			}
			return ZeitnahmeFromSqlcZiel(existing), nil
		}
		return Zeitnahme{}, err
	}

	return ZeitnahmeFromSqlcZiel(q), nil
}

func UpdateZeitnahmeZiel(ctx context.Context, id int32, rennNr, startNr *string) (Zeitnahme, error) {
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

	p := sqlc.UpdateZeitnahmeZielParams{
		ID:           id,
		RennenNummer: rennenNummer,
		StartNummer:  startNummer,
	}

	q, err := DB.QueriesFromCtx(ctx).UpdateZeitnahmeZiel(ctx, p)
	if err != nil {
		return Zeitnahme{}, err
	}

	return ZeitnahmeFromSqlcZiel(q), nil
}

func CreateZeitnahmeErgebnis(ctx context.Context, s, z Zeitnahme, meld Meldung) error {
	endZeit := z.TimeClient.Sub(*s.TimeClient)

	err := DB.WithTx(ctx, func(txCtx context.Context) error {
		params := sqlc.CreateZeitnahmeErgebnisParams{
			Endzeit:          endZeit.Seconds(),
			ZeitnahmeStartID: s.ID,
			ZeitnahmeZielID:  z.ID,
			MeldungUuid:      meld.Uuid,
		}

		q, err := DB.QueriesFromCtx(txCtx).CreateZeitnahmeErgebnis(txCtx, params)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				slog.Debug("create zeitnahme ergebnis: already exists", "start_id", s.ID, "ziel_id", z.ID)
			} else {
				return err
			}
		}

		err = DB.QueriesFromCtx(txCtx).SetZeitnahmeStartVerarbeitet(txCtx, s.ID)
		if err != nil {
			return err
		}
		err = DB.QueriesFromCtx(txCtx).SetZeitnahmeZielVerarbeitet(txCtx, z.ID)
		if err != nil {
			return err
		}

		slog.Debug("create zeitnahme ergebnis", "start_id", s.ID, "ziel_id", z.ID, "result", q)
		return nil
	})

	return err
}

func GetZeitnahmeErgebnisByMeld(ctx context.Context, meldUuid uuid.UUID) (sqlc.ZeitnahmeErgebni, error) {
	ctx, cancel := getCtx(ctx)
	defer cancel()

	return DB.QueriesFromCtx(ctx).GetZeitnahmeErgebnisByMeld(ctx, meldUuid)
}
