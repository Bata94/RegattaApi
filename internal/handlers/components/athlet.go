package components

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	regattabuero "github.com/bata94/RegattaApi/internal/templates/pages/regattabuero"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func NewAthletPost(w http.ResponseWriter, r *http.Request) {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}

	vorname := r.FormValue("vorname")
	name := r.FormValue("name")
	jahrgang := r.FormValue("jahrgang")
	geschlecht := r.FormValue("geschlecht")
	startberechtigt := r.FormValue("startberechtigt") == "on"

	fieldErrors := make(map[string]string)
	if vorname == "" {
		fieldErrors["vorname"] = "Vorname erforderlich"
	}
	if name == "" {
		fieldErrors["name"] = "Name erforderlich"
	}
	if jahrgang == "" {
		fieldErrors["jahrgang"] = "Jahrgang erforderlich"
	}
	if geschlecht != "m" && geschlecht != "w" && geschlecht != "x" {
		fieldErrors["geschlecht"] = "Geschlecht erforderlich"
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattabuero.NewAthlet(verein, "", fieldErrors), "Bitte alle Pflichtfelder ausfüllen")
		return
	}

	athletUuid := uuid.NewV7()

	a, err := crud.CreateAthlet(r.Context(), sqlc.CreateAthletParams{
		Uuid:            athletUuid,
		VereinUuid:      vereinUuid,
		Name:            name,
		Vorname:         vorname,
		Jahrgang:        jahrgang,
		Startberechtigt: startberechtigt,
		Geschlecht:      sqlc.Geschlecht(geschlecht),
	})
	if err != nil {
		webfw.ErrorWithForm(w, r, regattabuero.NewAthlet(verein, "", nil), "Fehler beim Anlegen des Athleten")
		return
	}

	a.Verein = &verein
	if err := regattabuero.NewAthletSuccess(a).Render(context.Background(), w); err != nil {
		slog.Warn("NewAthletSuccess render error", "err", err)
	}
}

func StartberechtigungPost(w http.ResponseWriter, r *http.Request) {
	slog.Debug("StartberechtigungPost", "formVal", r.FormValue("startberechtigt"))

	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	athletUuidStr := webfw.Param(r, "a_uuid")
	if athletUuidStr != r.FormValue("uuid") {
		webfw.ErrorToast(w, r, "UUIDs stimmen nicht überein")
		return
	}
	athletUuid, err := uuid.Parse(athletUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	_, err = crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading verein")
		return
	}
	athlet, err := crud.GetAthlet(r.Context(), athletUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading athlet")
		return
	}

	if athlet.VereinUuid != vereinUuid {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	formVal := r.FormValue("startberechtigt")
	formVal = strings.ToLower(formVal)
	if formVal != "on" && formVal != "true" {
		webfw.ErrorToast(w, r, "Bitte aktivieren Sie die Ärztliche Bescheinigung")
		return
	}

	err = athlet.UpdateStartberechtigung(r.Context(), true)
	if err != nil {
		slog.Error("UpdateStartberechtigung error", "err", err)
		webfw.ErrorToast(w, r, "Error while updating startberechtigung")
		return
	}

	webfw.SetRedirect(w, fmt.Sprintf("/internal/regattabuero/%s/startberechtigung", vereinUuidStr))
	w.WriteHeader(http.StatusOK)
}
