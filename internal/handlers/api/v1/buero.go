package api_v1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	DB "github.com/bata94/RegattaApi/internal/db"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/mailer"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

type AbmeldungsParams struct {
	Uuid string `json:"uuid"`
}

func KasseCreateRechnungPDF(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	v, err := crud.GetVereinMinimal(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	reNr, err := v.GetNextRechnungsnummer(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	filePath, err := utils.SavePDFfromHTML(
		"buero/kasse/rechnung/"+v.Uuid.String(),
		"rechnung/"+v.Kuerzel,
		reNr,
		true,
	)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	slog.Info("Generated", "file", filePath)

	toMail := []string{}
	obleute, err := crud.GetAllObmannForVerein(r.Context(), v.Uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	for _, o := range obleute {
		if o.Name.Valid {
			toMail = append(toMail, o.Email.String)
		}
	}

	err = mailer.Enqueue(r.Context(), mailer.Params{
		To:      toMail,
		CC:      []string{},
		Subject: fmt.Sprintf("MRG Regatta %02d - Rechnung %s", time.Now().Year()%100, reNr),
		Body:    "Anbei finden Sie eine neu erstellte Rechnung für Ihren Verein.",
		Files:   []string{filePath},
	})

	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "success")
}

func KasseCreateRechnungAllVereine(w http.ResponseWriter, r *http.Request) {
	vereine, err := crud.GetAllVerein(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	errLs := []error{}
	for _, v := range vereine {
		reNr, err := v.GetNextRechnungsnummer(r.Context())
		if err != nil {
			errLs = append(errLs, err)
			continue
		}

		filePath, err := utils.SavePDFfromHTML(
			"buero/kasse/rechnung/"+v.Uuid.String(),
			"rechnung/"+v.Kuerzel,
			reNr,
			true,
		)
		if err != nil {
			errLs = append(errLs, err)
			continue
		}
		slog.Info("Generated", "file", filePath)
	}

	if len(errLs) > 0 {
		webfw.JSON(w, r, errLs)
		return
	}
	webfw.JSON(w, r, "success")
}

func KasseCreateRechnungHTML(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	v, err := crud.GetVereinMinimal(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	type RechnungEntry struct {
		Tag         string
		Startnummer string
		Rennen      string
		Preis       string
	}

	var entries []RechnungEntry
	var reNr string
	var sumPreis int

	err = DB.WithTx(r.Context(), func(txCtx context.Context) error {
		meld, err := crud.GetAllMeldungForVerein(txCtx, v.Uuid)
		if err != nil {
			return err
		}

		reNr, err = v.GetNextRechnungsnummer(txCtx)
		if err != nil {
			return err
		}

		entries = nil
		sumPreis = 0

		for _, m := range meld {
			if m.RechnungsNummer.String != "" {
				continue
			}

			entries = append(entries, RechnungEntry{
				Tag:         string(m.Rennen.Tag),
				Startnummer: strconv.Itoa(int(m.StartNummer)),
				Rennen:      m.Rennen.Bezeichnung,
				Preis:       strconv.Itoa(int(m.Kosten)) + ",00 €",
			})
			sumPreis += int(m.Kosten)

			err := crud.SetMeldungRechnungsNummer(txCtx, m.Uuid, reNr)
			if err != nil {
				slog.Error("Error", "err", err)
				return err
			}
		}

		if len(entries) == 0 {
			return apierr.ErrNotFound
		}

		err = crud.CreateRechnung(txCtx, reNr, v.Uuid, sumPreis)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, apierr.ErrNotFound) {
			webfw.APIError(w, webfw.NotFound("Keine Meldungen gefunden!"))
			return
		}
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "Rechnung generated")
}
