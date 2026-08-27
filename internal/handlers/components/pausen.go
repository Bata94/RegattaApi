package components

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func PausenNew(w http.ResponseWriter, r *http.Request) {
	nachRennenUuidStr := webfw.Param(r, "nach_rennen_uuid")
	nachRennenUuid, err := uuid.Parse(nachRennenUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	p := crud.Pause{Pause: sqlc.Pause{ID: int32(-2), NachRennenUuid: nachRennenUuid, Laenge: 45}}

	templ.Handler(regattaleitung.PausenEntry(p)).ServeHTTP(w, r)
}

func PausenPost(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error(fmt.Sprintf("ID: %s - Error: %s", idStr, err.Error()))
		webfw.ErrorToast(w, r, "Invalid ID")
		return
	}
	laengeStr := r.FormValue("laenge")
	laenge, err := strconv.Atoi(laengeStr)
	if err != nil || laenge < 0 || laenge > 120 {
		slog.Error(fmt.Sprintf("Laenge: %s - Error: %s", laengeStr, err.Error()))
		webfw.ErrorToast(w, r, "Invalid laenge")
		return
	}
	nachRennenUuidStr := r.FormValue("nach_rennen_uuid")
	nachRennenUuid, err := uuid.Parse(nachRennenUuidStr)
	if err != nil {
		slog.Error(fmt.Sprintf("UUID: %s - Error: %s", nachRennenUuidStr, err.Error()))
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	if id == -2 {
		_, err = crud.CreatePause(r.Context(), sqlc.CreatePauseParams{
			NachRennenUuid: nachRennenUuid,
			Laenge:         int32(laenge),
		})
		if err != nil {
			webfw.ErrorToast(w, r, "Error while creating pause")
			return
		}

		templ.Handler(regattaleitung.Pausen()).ServeHTTP(w, r)
	} else {
		_, err = crud.UpdatePause(r.Context(), sqlc.UpdatePauseParams{
			ID:     int32(id),
			Laenge: int32(laenge),
		})
		if err != nil {
			webfw.ErrorToast(w, r, "Error while updating pause")
			return
		}

		templ.Handler(regattaleitung.Pausen()).ServeHTTP(w, r)
	}
}

func PausenDelete(w http.ResponseWriter, r *http.Request) {
	idStr := webfw.Param(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error(fmt.Sprintf("ID: %s - Error: %s", idStr, err.Error()))
		webfw.ErrorToast(w, r, "Invalid ID")
		return
	}

	err = crud.DeletePause(r.Context(), int32(id))
	if err != nil {
		webfw.ErrorToast(w, r, "Error while deleting pause")
		return
	}

	templ.Handler(regattaleitung.Pausen()).ServeHTTP(w, r)
}
