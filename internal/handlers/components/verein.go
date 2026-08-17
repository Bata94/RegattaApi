package components

import (
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/sqlc"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func VereinEditNew(c *handler.Context) error {
	var v crud.Verein
	switch c.Param("uuid") {
	case "":
		return handler.NotFound("Verein not found")
	case "new":
		v = crud.Verein{}
	default:
		vereinUuid, err := uuid.Parse(c.Param("uuid"))
		if err != nil {
			return handler.NotAcceptable("Invalid UUID")
		}
		v, err = crud.GetVereinMinimal(c.Request.Context(), vereinUuid)
		if err != nil {
			return handler.NotFound("Verein not found")
		}
	}

	templ.Handler(regattaleitung.VereinEdit(v, "", nil)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func VereinEditNewPost(c *handler.Context) error {
	uuidStr := c.Param("uuid")
	isNew := uuidStr == "new"

	var vereinUuid uuid.UUID
	if isNew {
		var err error
		vereinUuid, err = uuid.NewV7()
		if err != nil {
			return handler.NotAcceptable("Bad Request")
		}
	} else {
		var err error
		vereinUuid, err = uuid.Parse(uuidStr)
		if err != nil {
			return handler.NotAcceptable("Bad Request")
		}
	}

	name := c.FormValue("name")
	kurzform := c.FormValue("kurzform")
	kuerzel := c.FormValue("kuerzel")

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
		return handler.BadRequest("Bitte alle Pflichtfelder ausfüllen").WithForm(regattaleitung.VereinEdit(v, "", fieldErrors))
	}

	if isNew {
		_, err := crud.CreateVerein(c.Request.Context(), sqlc.CreateVereinParams{
			Uuid:     vereinUuid,
			Name:     name,
			Kurzform: kurzform,
			Kuerzel:  kuerzel,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				fieldErrors["kuerzel"] = "Kürzel bereits vergeben"
				return handler.BadRequest("Kürzel bereits vergeben").WithForm(regattaleitung.VereinEdit(v, "", fieldErrors))
			}
			return handler.BadRequest("Fehler beim Erstellen des Vereins").WithForm(regattaleitung.VereinEdit(v, "", nil))
		}
	} else {
		_, err := crud.UpdateVerein(c.Request.Context(), vereinUuid, sqlc.UpdateVereinParams{
			Uuid:     vereinUuid,
			Name:     name,
			Kurzform: kurzform,
			Kuerzel:  kuerzel,
		})
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				fieldErrors["kuerzel"] = "Kürzel bereits vergeben"
				return handler.BadRequest("Kürzel bereits vergeben").WithForm(regattaleitung.VereinEdit(v, "", fieldErrors))
			}
			return handler.BadRequest("Fehler beim Aktualisieren des Vereins").WithForm(regattaleitung.VereinEdit(v, "", nil))
		}
	}

	c.Writer.Header().Set("HX-Redirect", "/internal/regattaleitung/vereine")
	c.Writer.WriteHeader(http.StatusOK)
	return nil
}

func VereinDelete(c *handler.Context) error {
	vereinUuid, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		return handler.NotAcceptable("Invalid UUID")
	}

	athletenCount, err := crud.CountAthletenForVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return handler.InternalError("Error while checking verein")
	}

	if athletenCount > 0 {
		return handler.BadRequest("Verein kann nicht gelöscht werden, da noch Athleten zugeordnet sind")
	}

	if err := crud.DeleteVerein(c.Request.Context(), vereinUuid); err != nil {
		return handler.InternalError("Error while deleting verein")
	}

	templ.Handler(regattaleitung.Vereinsverwaltung()).ServeHTTP(c.Writer, c.Request)
	return nil
}
