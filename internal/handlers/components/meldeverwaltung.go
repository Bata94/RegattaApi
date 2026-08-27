package components

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bata94/RegattaApi/internal/crud"
	api_v1 "github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/sqlc"
	regattabuero "github.com/bata94/RegattaApi/internal/templates/pages/regattabuero"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func AbmeldungDelete(w http.ResponseWriter, r *http.Request) {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	_, err = crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}
	meldung, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}

	if meldung.VereinUuid != vereinUuid {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	err = crud.Abmeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while deleting meldung")
		return
	}

	webfw.SetRedirect(w, fmt.Sprintf("/internal/regattabuero/%s/abmeldung", vereinUuid))
	w.WriteHeader(http.StatusOK)
}

func UmmeldungPost(w http.ResponseWriter, r *http.Request) {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}
	meldung, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}

	if meldung.VereinUuid != verein.Uuid {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	athleten, err := crud.GetAllAthletenForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading athleten")
		return
	}

	if err := r.ParseForm(); err != nil {
		webfw.ErrorToast(w, r, "Error parsing form")
		return
	}

	fieldErrors := make(map[string]string)
	for i := range meldung.Athleten {
		athUuidStr := r.FormValue(fmt.Sprintf("athleten_%d", i))
		if athUuidStr == "" {
			continue
		}
		athUuid, err := uuid.Parse(athUuidStr)
		if err != nil {
			fieldErrors[fmt.Sprintf("athleten_%d", i)] = "Ungültige UUID"
			continue
		}
		if athUuid == meldung.Athleten[i].Uuid {
			continue
		}
		err = crud.Ummeldung(r.Context(), sqlc.UmmeldungParams{
			MeldungUuid: meldungUuid,
			Rolle:       *meldung.Athleten[i].Rolle,
			Position:    int32(*meldung.Athleten[i].Position),
			AthletUuid:  athUuid,
		})
		if err != nil {
			fieldErrors[fmt.Sprintf("athleten_%d", i)] = "Fehler beim Aktualisieren"
		}
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattabuero.UmmeldungMeldung(verein, meldung, athleten, "", fieldErrors), "Fehler bei der Ummeldung")
		return
	}

	meldungen, err := crud.GetAllMeldungForVerein(r.Context(), verein.Uuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading meldungen")
		return
	}

	webfw.SetPushUrl(w, fmt.Sprintf("/internal/regattabuero/%s/ummeldung", verein.Uuid))
	w.WriteHeader(http.StatusOK)
	if err := regattabuero.Ummeldung(verein, meldungen).Render(context.Background(), w); err != nil {
		slog.Warn("Ummeldung render error", "err", err)
	}
}

func NachmeldungPost(w http.ResponseWriter, r *http.Request) {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	rennenUuidStr := webfw.Param(r, "r_uuid")
	rennenUuid, err := uuid.Parse(rennenUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	rennen, err := crud.GetRennen(r.Context(), rennenUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading rennen")
		return
	}

	if err := r.ParseForm(); err != nil {
		webfw.ErrorToast(w, r, "Error parsing form")
		return
	}

	vereinUuid, err := uuid.Parse(r.FormValue("verein_uuid"))
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}

	athleten, err := crud.GetAllAthletenForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading athleten")
		return
	}

	numAthletes, stmRequired := rennen.GetTeilnehmerMeldeParams()

	params := api_v1.PostNachmeldungParams{
		VereinUuid:                    r.FormValue("verein_uuid"),
		RennenUuid:                    r.FormValue("rennen_uuid"),
		DoppeltesMeldentgeldBefreiung: r.FormValue("doppeltes_meldentgeld_befreiung") != "",
		Athleten:                      []api_v1.PostNachmeldungAthletParams{},
	}

	fieldErrors := make(map[string]string)
	hasAthlete := false
	for i := range numAthletes {
		athVal := r.FormValue(fmt.Sprintf("athleten_%d", i))
		if athVal == "" || athVal == "---" {
			continue
		}
		hasAthlete = true
		params.Athleten = append(params.Athleten, api_v1.PostNachmeldungAthletParams{
			AthletUuid: athVal,
			Position:   strconv.Itoa(i),
		})
	}

	if stmRequired {
		stmVal := r.FormValue("stm")
		if stmVal != "" && stmVal != "---" {
			hasAthlete = true
			params.Athleten = append(params.Athleten, api_v1.PostNachmeldungAthletParams{
				AthletUuid: stmVal,
				Position:   "stm",
			})
		}
	}

	if !hasAthlete {
		fieldErrors["athleten_0"] = "Mindestens ein Teilnehmer erforderlich"
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattabuero.NachmeldungMeldung(verein, rennen, athleten, "", fieldErrors), "Bitte wähle mindestens einen Teilnehmer aus")
		return
	}

	m, err := api_v1.CreateNachmeldung(r.Context(), params)
	if err != nil {
		webfw.ErrorToast(w, r, "Error creating nachmeldung: "+err.Error())
		return
	}
	meldung, err := crud.GetMeldung(r.Context(), m.Uuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}

	webfw.SetPushUrl(w, fmt.Sprintf("/internal/regattabuero/%s/nachmeldung/success/%s", vereinUuidStr, m.Uuid.String()))
	w.WriteHeader(http.StatusOK)
	if err := regattabuero.NachmeldungSuccess(meldung).Render(context.Background(), w); err != nil {
		slog.Warn("NachmeldungSuccess render error", "err", err)
	}
}
