package mailer

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/jackc/pgx/v5"
)

const (
	pollInterval     = 10 * time.Second
	maxAttempts      = int32(5)
	baseBackoff      = 30 * time.Second
	maxBackoff       = 15 * time.Minute
	claimBaseBackoff = time.Second
)

type Params struct {
	To      []string
	CC      []string
	Bcc     []string
	Subject string
	Body    string
	Files   []string
}

func Enqueue(ctx context.Context, p Params) error {
	if len(p.To) == 0 {
		return errors.New("no recipients provided")
	}
	if p.CC == nil {
		p.CC = []string{}
	}
	if p.Bcc == nil {
		p.Bcc = []string{}
	}
	if p.Files == nil {
		p.Files = []string{}
	}

	_, err := crud.EnqueueEmail(ctx, sqlc.CreateEmailQueueEntryParams{
		ToAddresses:  p.To,
		CcAddresses:  p.CC,
		BccAddresses: p.Bcc,
		Subject:      p.Subject,
		Body:         p.Body,
		Attachments:  p.Files,
		MaxAttempts:  maxAttempts,
	})
	return err
}

func RunWorker(ctx context.Context) {
	slog.Info("Starting email worker")
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	claimFailures := 0

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping email worker")
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				slog.Info("Stopping email worker")
				return
			}
			if err := processQueue(ctx); err != nil {
				if ctx.Err() != nil {
					slog.Info("Stopping email worker")
					return
				}
				claimFailures++
				backoff := claimBackoff(claimFailures)
				if claimFailures <= 3 {
					slog.Warn("Failed to claim next email, backing off", "err", err, "consecutive_failures", claimFailures, "backoff", backoff)
				} else {
					slog.Error("Failed to claim next email, backing off", "err", err, "consecutive_failures", claimFailures, "backoff", backoff)
				}
				if err := waitFor(ctx, backoff); err != nil {
					slog.Info("Stopping email worker")
					return
				}
				continue
			}
			claimFailures = 0
		}
	}
}

func processQueue(ctx context.Context) error {
	for {
		entry, err := crud.ClaimNextEmail(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		}

		sendErr := utils.SendMail(ctx, utils.SendMailParams{
			To:      entry.ToAddresses,
			CC:      entry.CcAddresses,
			Bcc:     entry.BccAddresses,
			Subject: entry.Subject,
			Body:    entry.Body,
			Files:   entry.Attachments,
		})

		if sendErr != nil {
			backoff := nextBackoff(entry.Attempts)
			if err := crud.MarkEmailFailed(ctx, entry.Uuid, backoff.Seconds(), sendErr.Error()); err != nil {
				if ctx.Err() == nil {
					slog.Error("Failed to mark email as failed", "uuid", entry.Uuid, "err", err)
				}
			} else {
				slog.Warn("Email send failed, will retry", "uuid", entry.Uuid, "attempts", entry.Attempts+1, "backoff", backoff)
			}
			continue
		}

		if err := crud.MarkEmailSent(ctx, entry.Uuid); err != nil {
			if ctx.Err() == nil {
				slog.Error("Failed to mark email as sent", "uuid", entry.Uuid, "err", err)
			}
		} else {
			slog.Info("Email sent", "uuid", entry.Uuid, "to", entry.ToAddresses)
		}
	}
}

func nextBackoff(attempts int32) time.Duration {
	backoff := baseBackoff * time.Duration(math.Pow(2, float64(attempts)))
	if backoff > maxBackoff {
		return maxBackoff
	}
	return backoff
}

func claimBackoff(failures int) time.Duration {
	backoff := claimBaseBackoff * time.Duration(math.Pow(2, float64(failures-1)))
	if backoff > maxBackoff {
		return maxBackoff
	}
	return backoff
}

func waitFor(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
