package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func RennenTab(w http.ResponseWriter, r *http.Request) {
	wettkampfStr := webfw.Param(r, "wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	showEmpty := webfw.Query(r, "show_empty") == "true"
	showStarted := webfw.Query(r, "show_started") == "true"
	urlFormatStr := webfw.Query(r, "url_format_str")

	templ.Handler(ui_components.RennenTab(wettkampf, urlFormatStr, showEmpty, showStarted)).ServeHTTP(w, r)
}

func ZeitplanCollapseBody(w http.ResponseWriter, r *http.Request) {
	wettkampfStr := webfw.Param(r, "wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Wettkampf not found")
		return
	}
	templ.Handler(ui_components.ZeitplanCollapseBody(wettkampf)).ServeHTTP(w, r)
}

func AusschreibungRennenCollapseBody(w http.ResponseWriter, r *http.Request) {
	wettkampfStr := webfw.Param(r, "wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Wettkampf not found")
		return
	}
	templ.Handler(ui_pages.AusschreibungRennenCollapseBody(wettkampf)).ServeHTTP(w, r)
}

func MeldeergebnisCollapseBody(w http.ResponseWriter, r *http.Request) {
	wettkampfStr := webfw.Param(r, "wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Wettkampf not found")
		return
	}
	templ.Handler(ui_pages.MeldeergebnisCollapseBody(wettkampf)).ServeHTTP(w, r)
}
