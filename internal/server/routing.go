package server

import (
	"net/http"

	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
)

var navBarConfig = ui_components.NewNavBarConfig()

func GetRouter() http.Handler {
	return newChiRouter()
}
