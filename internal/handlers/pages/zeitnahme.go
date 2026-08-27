package pages

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	startlisten "github.com/bata94/RegattaApi/internal/templates/pages/startlisten"
	"github.com/bata94/RegattaApi/internal/templates/pages/zeitnahme"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func InternalZeitnahme(w http.ResponseWriter, r *http.Request) templ.Component {
	return zeitnahme.Zeitnahme()
}

func InternalZeitnahmeZiel(w http.ResponseWriter, r *http.Request) templ.Component {
	return zeitnahme.Ziel()
}

func InternalZeitnahmeVorsortierung(w http.ResponseWriter, r *http.Request) templ.Component {
	return zeitnahme.Vorsortierung()
}

func InternalZeitnahmeWenderichter(w http.ResponseWriter, r *http.Request) templ.Component {
	return zeitnahme.Wenderichter()
}

func InternalZeitnahmeStart(w http.ResponseWriter, r *http.Request) templ.Component {
	rennen, err := crud.GetAllRennenWithAthlet(r.Context(), crud.GetAllRennenParams{
		GetMeldungen: true,
		GetAthleten:  true,
		ShowEmpty:    false,
		ShowStarted:  false,
	})
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Fehler beim Laden der Rennen"))
		return nil
	}

	for i := range rennen {
		for j := range rennen[i].Meldungen {
			rennen[i].Meldungen[j].Rennen = nil
		}
	}

	return zeitnahme.Start(rennen)
}

func InternalStartlisten(w http.ResponseWriter, r *http.Request) templ.Component {
	return startlisten.Startlisten()
}
