package api_v1

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/google/uuid"
)

type WSZnMsg struct {
	Status *string         `json:"status"`
	Method string          `json:"method"`
	Data   *crud.Zeitnahme `json:"data"`
}

func WsZeitnahmeZiel(c *handler.Context) error {
	defer func() {
		handlers.Unregister <- nil
		c.Writer.WriteHeader(200)
	}()

	handlers.Register <- nil

	q, err := crud.GetOpenZeitnahmeZiel(c.Request.Context())
	if err != nil {
		errStr := fmt.Sprint("Error getting open ZnZiel... ", err)
		slog.Error(errStr)
		c.Writer.Write([]byte(errStr))
		return nil
	}

	qJson, err := json.Marshal(map[string]interface{}{"list": q})
	if err != nil {
		errStr := fmt.Sprint("Error getting open ZnZiel... ", err)
		slog.Error(errStr)
		c.Writer.Write([]byte(errStr))
		return nil
	}
	c.Writer.Write(qJson)
	return nil
}

type PostStartParams struct {
	RennenNummer    *string   `json:"renn_nummer"`
	StartNummern    []string  `json:"start_nummern"`
	TimeClient      time.Time `json:"time_client"`
	MeasuredLatency *int      `json:"measured_latency"`
}

func PostZeitnahmeStart(c *handler.Context) error {
	p := new(PostStartParams)
	err := c.BodyParser(p)
	if err != nil {
		return err
	}

	q, err := crud.CreateZeitnahmeStart(c.Request.Context(), p.RennenNummer, p.StartNummern, p.TimeClient, *p.MeasuredLatency)
	if err != nil {
		return err
	}

	return c.JSON(q)
}

func GetOpenStarts(c *handler.Context) error {
	q, err := crud.GetOpenZeitnahmeStart(c.Request.Context())
	if err != nil {
		return err
	}

	return c.JSON(q)
}

func GenerateEndZeit(c *handler.Context) error {
	starts, err := crud.GetOpenZeitnahmeStart(c.Request.Context())
	if err != nil {
		slog.Debug("GetOpenZeitnahmeStart")
		return err
	}

	ziels, err := crud.GetOpenZeitnahmeZiel(c.Request.Context())
	if err != nil {
		slog.Debug("GetOpenZeitnahmeZiel")
		return err
	}

	if len(ziels) == 0 {
		slog.Debug("0 Ziels")
		return handler.BadRequest("No ziels")
	}
	if len(starts) == 0 {
		slog.Debug("0 Starts")
		return handler.BadRequest("No starts")
	}

	for _, z := range ziels {
		if z.StartNummer == nil || *z.StartNummer == "" {
			continue
		}
		for _, s := range starts {
			if *z.StartNummer == *s.StartNummer {
				startNummerInt, err := strconv.Atoi(*s.StartNummer)
				if err != nil {
					slog.Error("Error StartNummerStr to int")
					return err
				}
				meld, err := crud.GetMeldungByStartNrUndTag(c.Request.Context(), startNummerInt, crud.TagSa)
				if err != nil {
					slog.Debug("GetMeldungByStartNrUndTag")
					return err
				}
				if meld.Uuid == uuid.Nil {
					slog.Warn("GetMeldungByStartNrUndTag meld.Uuid is nil")
					return handler.BadRequest("Meldung not found")
				}

				err = crud.CreateZeitnahmeErgebnis(c.Request.Context(), s, z, meld)
				if err != nil {
					slog.Debug("CreateZeitnahmeErgebnis")
					return err
				}
			}
		}
	}

	return c.JSON("success")
}
