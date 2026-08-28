package components

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bata94/RegattaApi/internal/crud"
	regattabuero "github.com/bata94/RegattaApi/internal/templates/pages/regattabuero"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func WaagePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.Error("ParseForm error", "err", err)
		webfw.ErrorToast(w, r, "Fehler beim Verarbeiten der Anfrage")
		return
	}

	idStr := r.FormValue("uuid")
	gewichtStr := r.FormValue("gewicht")

	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.Error("Parse UUID error", "err", err)
		webfw.ErrorToast(w, r, "Ungültige UUID")
		return
	}

	ath, err := crud.GetAthletMinimal(r.Context(), id)
	if err != nil {
		slog.Error("GetAthletMinimal error", "err", err)
		webfw.ErrorToast(w, r, err.Error())
		return
	}

	fieldErrors := make(map[string]string)
	gewichtFloat, err := strconv.ParseFloat(gewichtStr, 32)
	if err != nil {
		fieldErrors["gewicht"] = "Ungültiges Gewicht"
	}
	gewicht := int(gewichtFloat * 10)

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattabuero.Waage(ath, "", fieldErrors), "Ungültiges Gewicht")
		return
	}

	err = ath.UpdateGewicht(r.Context(), gewicht)
	if err != nil {
		slog.Error("UpdateGewicht error", "err", err)
		webfw.ErrorToast(w, r, "Fehler beim Aktualisieren des Gewichts")
		return
	}

	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Verein nicht gefunden")
		return
	}
	athleten, err := crud.GetAllAthletenForVereinWaage(r.Context(), verein.Uuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading athleten")
		return
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	webfw.SetPushUrl(w, fmt.Sprintf("/internal/regattabuero/%s/waage", vereinUuidStr))
	w.WriteHeader(http.StatusOK)
	if err := regattabuero.WaageWahl(verein, athleten).Render(context.Background(), w); err != nil {
		slog.Warn("WaageWahl render error", "err", err)
	}
}
