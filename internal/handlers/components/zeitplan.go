package components

import (
	"net/http"
	"strconv"

	"github.com/bata94/RegattaApi/internal/service"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func ZeitplanPost(w http.ResponseWriter, r *http.Request) {
	startzeit_saStr := r.FormValue("startzeit_sa")
	startzeit_soStr := r.FormValue("startzeit_so")

	fieldErrors := make(map[string]string)

	startzeit_sa, err := strconv.Atoi(startzeit_saStr)
	if err != nil || startzeit_sa < 0 || startzeit_sa > 24 {
		fieldErrors["startzeit_sa"] = "Ungültige Startzeit (0-24)"
	}
	startzeit_so, err := strconv.Atoi(startzeit_soStr)
	if err != nil || startzeit_so < 0 || startzeit_so > 24 {
		fieldErrors["startzeit_so"] = "Ungültige Startzeit (0-24)"
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattaleitung.Zeitplan("", fieldErrors), "Ungültige Startzeit")
		return
	}

	zeitplan := service.SetZeitplanParams{
		SaStartStunde: startzeit_sa,
		SoStartStunde: startzeit_so,
	}

	err = service.SetZeitplan(r.Context(), zeitplan)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while creating zeitplan")
		return
	}

	webfw.SuccessToast(w, r, "Zeitplan erstellt")
}
