package api_v1

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func GetOpenZeitnahmeZiel(w http.ResponseWriter, r *http.Request) {
	q, err := crud.GetOpenZeitnahmeZiel(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	webfw.JSON(w, r, map[string]any{"list": q})
}

type PostStartParams struct {
	RennenNummer    *string   `json:"renn_nummer"`
	StartNummern    []string  `json:"start_nummern"`
	TimeClient      time.Time `json:"time_client"`
	MeasuredLatency *int      `json:"measured_latency"`
}

func PostZeitnahmeStart(w http.ResponseWriter, r *http.Request) {
	p := new(PostStartParams)
	err := webfw.ParseBody(r, p)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	q, err := crud.CreateZeitnahmeStart(r.Context(), p.RennenNummer, p.StartNummern, p.TimeClient, *p.MeasuredLatency, uuid.New().String(), uuid.New().String())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	handlers.GetHub().BroadcastJSON(map[string]any{
		"type": "new_start",
		"data": q,
	})

	webfw.JSON(w, r, q)
}

func GetOpenStarts(w http.ResponseWriter, r *http.Request) {
	q, err := crud.GetOpenZeitnahmeStart(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, q)
}

func GenerateEndZeit(w http.ResponseWriter, r *http.Request) {
	starts, err := crud.GetOpenZeitnahmeStart(r.Context())
	if err != nil {
		slog.Debug("GetOpenZeitnahmeStart")
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	ziels, err := crud.GetOpenZeitnahmeZiel(r.Context())
	if err != nil {
		slog.Debug("GetOpenZeitnahmeZiel")
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	if len(ziels) == 0 {
		slog.Debug("0 Ziels")
		webfw.APIError(w, webfw.BadRequest("No ziels"))
		return
	}
	if len(starts) == 0 {
		slog.Debug("0 Starts")
		webfw.APIError(w, webfw.BadRequest("No starts"))
		return
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
					webfw.APIError(w, webfw.InternalError(err.Error()))
					return
				}
				meld, err := crud.GetMeldungByStartNrUndTag(r.Context(), startNummerInt, crud.TagSa)
				if err != nil {
					slog.Debug("GetMeldungByStartNrUndTag")
					webfw.APIError(w, webfw.InternalError(err.Error()))
					return
				}
				if meld.Uuid == uuid.Nil {
					slog.Warn("GetMeldungByStartNrUndTag meld.Uuid is nil")
					webfw.APIError(w, webfw.BadRequest("Meldung not found"))
					return
				}

				err = crud.CreateZeitnahmeErgebnis(r.Context(), s, z, meld)
				if err != nil {
					slog.Debug("CreateZeitnahmeErgebnis")
					webfw.APIError(w, webfw.InternalError(err.Error()))
					return
				}

				handlers.GetHub().BroadcastJSON(map[string]any{
					"type": "finish_confirmed",
					"data": map[string]any{
						"startNummer": *s.StartNummer,
						"rennNummer":  s.RennenNummer,
						"endzeit":     z.TimeClient.Sub(*s.TimeClient).Seconds(),
					},
				})
			}
		}
	}

	webfw.JSON(w, r, "success")
}
