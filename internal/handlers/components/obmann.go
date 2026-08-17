package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ObmannEditNew(c *handler.Context) error {
	var o crud.Obmann
	switch c.Param("uuid") {
	case "":
		return handler.NotFound("Obmann not found")
	case "new":
		o = crud.Obmann{Obmann: &sqlc.Obmann{}}
	default:
		obmannUuid, err := uuid.Parse(c.Param("uuid"))
		if err != nil {
			return handler.NotAcceptable("Invalid UUID")
		}
		o, err = crud.GetObmannMinimal(c.Request.Context(), obmannUuid)
		if err != nil {
			return handler.NotFound("Obmann not found")
		}
	}

	templ.Handler(regattaleitung.ObmannEdit(o, "", nil)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func ObmannEditNewPost(c *handler.Context) error {
	uuidStr := c.Param("uuid")
	isNew := uuidStr == "new"

	var obmannUuid uuid.UUID
	if isNew {
		var err error
		obmannUuid, err = uuid.NewV7()
		if err != nil {
			return handler.NotAcceptable("Bad Request")
		}
	} else {
		var err error
		obmannUuid, err = uuid.Parse(uuidStr)
		if err != nil {
			return handler.NotAcceptable("Bad Request")
		}
	}

	nameStr := c.FormValue("name")
	emailStr := c.FormValue("email")
	phoneStr := c.FormValue("phone")
	vereinUuidStr := c.FormValue("verein_uuid")

	vereinUuid := uuid.Nil
	if vereinUuidStr != "" {
		parsed, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			fieldErrors := map[string]string{"verein_uuid": "Ungültiger Verein"}
			return handler.BadRequest("Ungültiger Verein").WithForm(regattaleitung.ObmannEdit(crud.Obmann{Obmann: &sqlc.Obmann{
				Uuid:       obmannUuid,
				Name:       pgtype.Text{String: nameStr, Valid: nameStr != ""},
				Email:      pgtype.Text{String: emailStr, Valid: emailStr != ""},
				Phone:      pgtype.Text{String: phoneStr, Valid: phoneStr != ""},
				VereinUuid: vereinUuid,
			}}, "", fieldErrors))
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
		return handler.BadRequest("Bitte alle Pflichtfelder ausfüllen").WithForm(regattaleitung.ObmannEdit(o, "", fieldErrors))
	}

	name := pgtype.Text{String: nameStr, Valid: true}
	email := pgtype.Text{String: emailStr, Valid: true}
	phone := pgtype.Text{String: phoneStr, Valid: phoneStr != ""}

	if isNew {
		_, err := crud.CreateObmann(c.Request.Context(), sqlc.CreateObmannParams{
			Uuid:       obmannUuid,
			VereinUuid: vereinUuid,
			Name:       name,
			Email:      email,
			Phone:      phone,
		})
		if err != nil {
			return handler.BadRequest("Fehler beim Erstellen des Obmanns").WithForm(regattaleitung.ObmannEdit(o, "", nil))
		}
	} else {
		_, err := crud.UpdateObmann(c.Request.Context(), obmannUuid, sqlc.UpdateObmannParams{
			Uuid:       obmannUuid,
			Name:       name,
			Email:      email,
			Phone:      phone,
			VereinUuid: vereinUuid,
		})
		if err != nil {
			return handler.BadRequest("Fehler beim Aktualisieren des Obmanns").WithForm(regattaleitung.ObmannEdit(o, "", nil))
		}
	}

	c.Writer.Header().Set("HX-Redirect", "/internal/regattaleitung/obleute")
	c.Writer.WriteHeader(http.StatusOK)
	return nil
}

func ObmannDelete(c *handler.Context) error {
	obmannUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	if err := crud.DeleteObmann(c.Request.Context(), obmannUuid); err != nil {
		return handler.InternalError("Error while deleting obmann")
	}

	templ.Handler(regattaleitung.Obleute()).ServeHTTP(c.Writer, c.Request)
	return nil
}
