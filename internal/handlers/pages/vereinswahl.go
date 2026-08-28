package pages

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	vereinswahl "github.com/bata94/RegattaApi/internal/templates/pages/vereinswahl"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func InternalVereinswahl(w http.ResponseWriter, r *http.Request) templ.Component {
	next := webfw.Query(r, "next")
	if next == "" {
		webfw.HandlePageError(w, r, webfw.BadRequest("Next param is required"))
		return nil
	}
	title := webfw.Query(r, "title")
	nextUrl := "/internal/regattabuero/%s/" + next

	var (
		vereine []crud.Verein
		err     error
	)

	switch next {
	case "waage":
		vereine, err = crud.GetForAllVereineMissingAthlet(r.Context(), crud.Waage)
	case "startberechtigung":
		vereine, err = crud.GetForAllVereineMissingAthlet(r.Context(), crud.Startberechtigt)
	default:
		vereine, err = crud.GetAllVerein(r.Context())
	}
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Fehler beim Laden der Vereine"))
		return nil
	}

	return vereinswahl.Vereinswahl(nextUrl, title, vereine)
}
