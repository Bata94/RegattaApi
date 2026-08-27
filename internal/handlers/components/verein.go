package components

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func VereinEditNew(w http.ResponseWriter, r *http.Request) {
	var v crud.Verein
	switch webfw.Param(r, "uuid") {
	case "":
		webfw.ErrorToast(w, r, "Verein not found")
		return
	case "new":
		v = crud.Verein{}
	default:
		vereinUuid, err := uuid.Parse(webfw.Param(r, "uuid"))
		if err != nil {
			webfw.ErrorToast(w, r, "Invalid UUID")
			return
		}
		v, err = crud.GetVereinMinimal(r.Context(), vereinUuid)
		if err != nil {
			webfw.ErrorToast(w, r, "Verein not found")
			return
		}
	}

	templ.Handler(regattaleitung.VereinEdit(v, "", nil)).ServeHTTP(w, r)
}

func VereinEditNewPost(w http.ResponseWriter, r *http.Request) {
	uuidStr := webfw.Param(r, "uuid")
	isNew := uuidStr == "new"

	var vereinUuid uuid.UUID
	if isNew {
		var err error
		vereinUuid, err = uuid.NewV7()
		if err != nil {
			webfw.ErrorToast(w, r, "Bad Request")
			return
		}
	} else {
		var err error
		vereinUuid, err = uuid.Parse(uuidStr)
		if err != nil {
			webfw.ErrorToast(w, r, "Bad Request")
			return
		}
	}

	name := r.FormValue("name")
	kurzform := r.FormValue("kurzform")
	kuerzel := r.FormValue("kuerzel")

	v := crud.Verein{Verein: sqlc.Verein{
		Uuid:     vereinUuid,
		Name:     name,
		Kurzform: kurzform,
		Kuerzel:  kuerzel,
	}}

	fieldErrors := make(map[string]string)
	if name == "" {
		fieldErrors["name"] = "Name erforderlich"
	}
	if kurzform == "" {
		fieldErrors["kurzform"] = "Kurzform erforderlich"
	}
	if kuerzel == "" {
		fieldErrors["kuerzel"] = "Kürzel erforderlich"
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattaleitung.VereinEdit(v, "", fieldErrors), "Bitte alle Pflichtfelder ausfüllen")
		return
	}

	if isNew {
		_, err := crud.CreateVerein(r.Context(), sqlc.CreateVereinParams{
			Uuid:     vereinUuid,
			Name:     name,
			Kurzform: kurzform,
			Kuerzel:  kuerzel,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				fieldErrors["kuerzel"] = "Kürzel bereits vergeben"
				webfw.ErrorWithForm(w, r, regattaleitung.VereinEdit(v, "", fieldErrors), "Kürzel bereits vergeben")
				return
			}
			webfw.ErrorWithForm(w, r, regattaleitung.VereinEdit(v, "", nil), "Fehler beim Erstellen des Vereins")
			return
		}
	} else {
		_, err := crud.UpdateVerein(r.Context(), vereinUuid, sqlc.UpdateVereinParams{
			Uuid:     vereinUuid,
			Name:     name,
			Kurzform: kurzform,
			Kuerzel:  kuerzel,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				fieldErrors["kuerzel"] = "Kürzel bereits vergeben"
				webfw.ErrorWithForm(w, r, regattaleitung.VereinEdit(v, "", fieldErrors), "Kürzel bereits vergeben")
				return
			}
			webfw.ErrorWithForm(w, r, regattaleitung.VereinEdit(v, "", nil), "Fehler beim Aktualisieren des Vereins")
			return
		}
	}

	webfw.SetRedirect(w, "/internal/regattaleitung/vereine")
	w.WriteHeader(http.StatusOK)
}

func VereinDelete(w http.ResponseWriter, r *http.Request) {
	vereinUuid, err := uuid.Parse(webfw.Param(r, "uuid"))
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	athletenCount, err := crud.CountAthletenForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while checking verein")
		return
	}

	if athletenCount > 0 {
		webfw.ErrorToast(w, r, "Verein kann nicht gelöscht werden, da noch Athleten zugeordnet sind")
		return
	}

	if err := crud.DeleteVerein(r.Context(), vereinUuid); err != nil {
		webfw.ErrorToast(w, r, "Error while deleting verein")
		return
	}

	templ.Handler(regattaleitung.Vereinsverwaltung()).ServeHTTP(w, r)
}
