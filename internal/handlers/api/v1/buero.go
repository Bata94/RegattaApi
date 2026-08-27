package api_v1

import (
	"log/slog"
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/mailer"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

type AbmeldungsParams struct {
	Uuid string `json:"uuid"`
}

func StartnummernAusgabe(w http.ResponseWriter, r *http.Request) {
	webfw.APIError(w, webfw.NotFound("Not found"))
}

func StartnummernWechsel(w http.ResponseWriter, r *http.Request) {
	webfw.APIError(w, webfw.NotFound("Not found"))
}

func KasseEinzahlung(w http.ResponseWriter, r *http.Request) {
	webfw.APIError(w, webfw.NotFound("Not found"))
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
		Subject: "MRG Regatta 24 - Rechnung " + reNr,
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

	meld, err := crud.GetAllMeldungForVerein(r.Context(), v.Uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	reNr, err := v.GetNextRechnungsnummer(r.Context())
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

	entries := []RechnungEntry{}
	sumPreis := 0

	for _, m := range meld {
		if m.RechnungsNummer.String != "" {
			continue
		}

		entries = append(entries, RechnungEntry{
			Tag:         string(m.Rennen.Tag),
			Startnummer: string(rune(int(m.StartNummer) + '0')),
			Rennen:      m.Rennen.Bezeichnung,
			Preis:       string(rune(int(m.Kosten)+'0')) + ",00 €",
		})
		sumPreis += int(m.Kosten)

		err := crud.SetMeldungRechnungsNummer(r.Context(), m.Uuid, reNr)
		if err != nil {
			slog.Error("Error", "err", err)
		}
	}

	if len(entries) == 0 {
		webfw.APIError(w, webfw.NotFound("Keine Meldungen gefunden!"))
		return
	}

	err = crud.CreateRechnung(r.Context(), reNr, v.Uuid, sumPreis)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "Rechnung generated")
}
