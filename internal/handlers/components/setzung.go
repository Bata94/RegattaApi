package components

import (
	"encoding/json"
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	api_v1 "github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func SetzungsVerwaltungLosungPost(w http.ResponseWriter, r *http.Request) {
	api_v1.SetzungsLosung(w, r)
	webfw.SuccessToast(w, r, "Losung erfolgreich!")
}

func SetzungsVerwaltungLosungDelete(w http.ResponseWriter, r *http.Request) {
	api_v1.ResetSetzung(w, r)
	webfw.SuccessToast(w, r, "Setzung erfolgreich zurückgesetzt!")
}

func SetzungsVerwaltungAenderungRennenPost(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		rUuid  uuid.UUID
		rennen crud.Rennen
	)

	rUuid, err = uuid.Parse(webfw.Param(r, "param"))
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	rennen, err = crud.GetRennen(r.Context(), rUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Rennen nicht gefunden")
		return
	}

	payloadStr := r.FormValue("params")
	payload := make(map[string]any)
	err = json.Unmarshal([]byte(payloadStr), &payload)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid JSON")
		return
	}

	meldOrderLs, ok := payload["order"].([]any)
	if !ok {
		webfw.ErrorToast(w, r, "Order nicht gefunden")
		return
	}
	abteilungParam := payload["abteilung"]
	if abteilungParam == nil {
		webfw.ErrorToast(w, r, "Abteilung nicht gefunden")
		return
	}
	targetAbteilung := int32(abteilungParam.(float64))

	for i, m := range meldOrderLs {
		mUuid, err := uuid.Parse(m.(string))
		if err != nil {
			webfw.ErrorToast(w, r, "Invalid UUID")
			return
		}

		for _, meldung := range rennen.Meldungen {
			if meldung.Uuid == mUuid {
				bahn := int32(i) + 1

				err = crud.UpdateMeldungSetzung(r.Context(), sqlc.UpdateMeldungSetzungParams{
					Uuid:      meldung.Uuid,
					Abteilung: targetAbteilung,
					Bahn:      bahn,
				})
				if err != nil {
					webfw.ErrorToast(w, r, "Error while updating meldung setzung")
					return
				}
				continue
			}
		}
	}

	webfw.SuccessToast(w, r, "Setzung erfolgreich!")
}
