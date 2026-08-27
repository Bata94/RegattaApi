package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ObmannEditNew(w http.ResponseWriter, r *http.Request) {
	var o crud.Obmann
	switch webfw.Param(r, "uuid") {
	case "":
		webfw.ErrorToast(w, r, "Obmann not found")
		return
	case "new":
		o = crud.Obmann{Obmann: &sqlc.Obmann{}}
	default:
		obmannUuid, err := uuid.Parse(webfw.Param(r, "uuid"))
		if err != nil {
			webfw.ErrorToast(w, r, "Invalid UUID")
			return
		}
		o, err = crud.GetObmannMinimal(r.Context(), obmannUuid)
		if err != nil {
			webfw.ErrorToast(w, r, "Obmann not found")
			return
		}
	}

	templ.Handler(regattaleitung.ObmannEdit(o, "", nil)).ServeHTTP(w, r)
}

func ObmannEditNewPost(w http.ResponseWriter, r *http.Request) {
	uuidStr := webfw.Param(r, "uuid")
	isNew := uuidStr == "new"

	var obmannUuid uuid.UUID
	if isNew {
		var err error
		obmannUuid, err = uuid.NewV7()
		if err != nil {
			webfw.ErrorToast(w, r, "Bad Request")
			return
		}
	} else {
		var err error
		obmannUuid, err = uuid.Parse(uuidStr)
		if err != nil {
			webfw.ErrorToast(w, r, "Bad Request")
			return
		}
	}

	nameStr := r.FormValue("name")
	emailStr := r.FormValue("email")
	phoneStr := r.FormValue("phone")
	vereinUuidStr := r.FormValue("verein_uuid")

	vereinUuid := uuid.Nil
	if vereinUuidStr != "" {
		parsed, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			fieldErrors := map[string]string{"verein_uuid": "Ungültiger Verein"}
			webfw.ErrorWithForm(w, r, regattaleitung.ObmannEdit(crud.Obmann{Obmann: &sqlc.Obmann{
				Uuid:       obmannUuid,
				Name:       pgtype.Text{String: nameStr, Valid: nameStr != ""},
				Email:      pgtype.Text{String: emailStr, Valid: emailStr != ""},
				Phone:      pgtype.Text{String: phoneStr, Valid: phoneStr != ""},
				VereinUuid: vereinUuid,
			}}, "", fieldErrors), "Ungültiger Verein")
			return
		}
		vereinUuid = parsed
	}

	o := crud.Obmann{Obmann: &sqlc.Obmann{
		Uuid:       obmannUuid,
		Name:       pgtype.Text{String: nameStr, Valid: nameStr != ""},
		Email:      pgtype.Text{String: emailStr, Valid: emailStr != ""},
		Phone:      pgtype.Text{String: phoneStr, Valid: phoneStr != ""},
		VereinUuid: vereinUuid,
	}}

	fieldErrors := make(map[string]string)
	if nameStr == "" {
		fieldErrors["name"] = "Name erforderlich"
	}
	if emailStr == "" {
		fieldErrors["email"] = "E-Mail erforderlich"
	}
	if vereinUuid == uuid.Nil {
		fieldErrors["verein_uuid"] = "Verein erforderlich"
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattaleitung.ObmannEdit(o, "", fieldErrors), "Bitte alle Pflichtfelder ausfüllen")
		return
	}

	name := pgtype.Text{String: nameStr, Valid: true}
	email := pgtype.Text{String: emailStr, Valid: true}
	phone := pgtype.Text{String: phoneStr, Valid: phoneStr != ""}

	if isNew {
		_, err := crud.CreateObmann(r.Context(), sqlc.CreateObmannParams{
			Uuid:       obmannUuid,
			VereinUuid: vereinUuid,
			Name:       name,
			Email:      email,
			Phone:      phone,
		})
		if err != nil {
			webfw.ErrorWithForm(w, r, regattaleitung.ObmannEdit(o, "", nil), "Fehler beim Erstellen des Obmanns")
			return
		}
	} else {
		_, err := crud.UpdateObmann(r.Context(), obmannUuid, sqlc.UpdateObmannParams{
			Uuid:       obmannUuid,
			Name:       name,
			Email:      email,
			Phone:      phone,
			VereinUuid: vereinUuid,
		})
		if err != nil {
			webfw.ErrorWithForm(w, r, regattaleitung.ObmannEdit(o, "", nil), "Fehler beim Aktualisieren des Obmanns")
			return
		}
	}

	webfw.SetRedirect(w, "/internal/regattaleitung/obleute")
	w.WriteHeader(http.StatusOK)
}

func ObmannDelete(w http.ResponseWriter, r *http.Request) {
	obmannUuid, err := uuid.Parse(webfw.Param(r, "uuid"))
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	if err := crud.DeleteObmann(r.Context(), obmannUuid); err != nil {
		webfw.ErrorToast(w, r, "Error while deleting obmann")
		return
	}

	templ.Handler(regattaleitung.Obleute()).ServeHTTP(w, r)
}
