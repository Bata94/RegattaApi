package pages

import (
	"net/http"

	"github.com/a-h/templ"
	dashboard "github.com/bata94/RegattaApi/internal/templates/pages/dashboard"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func InternalIndex(w http.ResponseWriter, r *http.Request) templ.Component {
	caps := webfw.GetCapabilities(r)
	if caps == nil {
		caps = []string{}
	}
	return dashboard.Dashboard(caps)
}
