package pages

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/templates/pages"
)

func Index(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Index()
}

func Livestream(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Livestream()
}

func Ausschreibung(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Ausschreibung()
}

func Zeitplan(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Zeitplan()
}

func Meldeergebnis(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Meldeergebnis()
}

func Ergebnisse(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Ergebnisse()
}

func Login(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Login("", nil)
}

func Datenschutz(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Datenschutz()
}
