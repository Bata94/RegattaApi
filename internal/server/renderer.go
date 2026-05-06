package server

import (
	"net/http"

	"github.com/a-h/templ"

	"github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/internal/templates/layout"
	"github.com/bata94/RegattaApi/internal/templates/pages"
)

func baseLayoutHandler(url string, pageBody templ.Component) {
	r.Handle("GET", url, pageHandler(pageBody, ui_layouts.BaseLayout))
}

func pageHandler(pageBody templ.Component, layout func(templ.Component, ui_components.NavBarConfig) templ.Component) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(pageBody).ServeHTTP(w, r)
		} else {
			templ.Handler(layout(pageBody, navBarConfig)).ServeHTTP(w, r)
		}
	}
}

func compHandler(url string, comp templ.Component) {
	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			templ.Handler(comp).ServeHTTP(w, r)
		} else {
			templ.Handler(ui_pages.Error(404, r.URL.Path + " not found")).ServeHTTP(w, r)
		}
	})
}
