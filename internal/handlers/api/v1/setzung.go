package api_v1

import (
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/service"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func shuffle(array []crud.Meldung) []crud.Meldung {
	for i := range array {
		j := rand.IntN(i + 1)
		array[i], array[j] = array[j], array[i]
	}
	return array
}

func SetzungsLosung(w http.ResponseWriter, r *http.Request) {
	check, err := crud.CheckMeldungSetzung(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	if check {
		webfw.APIError(w, webfw.BadRequest("Setzung bereits erledigt! Vorher reseten um zu wiederholen!"))
		return
	}

	allRennen, err := crud.GetAllRennen(r.Context(), crud.GetAllRennenParams{
		GetMeldungen:  true,
		ShowEmpty:     true,
		ShowStarted:   true,
		ShowWettkampf: sqlc.NullWettkampf{},
	})
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	for _, rennen := range allRennen {
		maxBahnen := 1

		switch rennen.Wettkampf {
		case sqlc.WettkampfKurzstrecke:
			maxBahnen = 4
		case sqlc.WettkampfSlalom:
			maxBahnen = 3
		case sqlc.WettkampfLangstrecke:
			maxBahnen = 99999
		case sqlc.WettkampfStaffel:
			maxBahnen = 2
		}

		numMeld := 0
		for _, m := range rennen.Meldungen {
			if !m.Abgemeldet {
				numMeld++
			}
		}
		if numMeld == 0 {
			continue
		}

		remainder := numMeld % maxBahnen
		numAbteilungen := numMeld / maxBahnen
		if remainder > 0 {
			numAbteilungen++
		}

		sizes := make([]int, numAbteilungen)
		for i := range sizes {
			sizes[i] = maxBahnen
		}
		if remainder > 0 {
			sizes[numAbteilungen-1] = remainder
		}
		if remainder == 1 && numAbteilungen >= 2 {
			sizes[numAbteilungen-2]--
			sizes[numAbteilungen-1]++
		}

		rennen.Meldungen = shuffle(rennen.Meldungen)

		abteilungIdx := 0
		bahn := int32(1)
		count := 0

		for _, m := range rennen.Meldungen {
			if m.Abgemeldet {
				continue
			}
			if err := crud.UpdateMeldungSetzung(r.Context(), sqlc.UpdateMeldungSetzungParams{
				Uuid:      m.Uuid,
				Abteilung: int32(abteilungIdx + 1),
				Bahn:      bahn,
			}); err != nil {
				webfw.APIError(w, webfw.InternalError(err.Error()))
				return
			}
			bahn++
			count++
			if count >= sizes[abteilungIdx] {
				abteilungIdx++
				bahn = 1
				count = 0
			}
		}
	}

	if webfw.IsHtmxRequest(r) {
		return
	} else {
		webfw.JSON(w, r, "Setzung erfolgreich!")
	}
}

func ResetSetzung(w http.ResponseWriter, r *http.Request) {
	mLs, err := crud.GetAllMeldungen(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	for _, m := range mLs {
		err = crud.UpdateMeldungSetzung(r.Context(), sqlc.UpdateMeldungSetzungParams{
			Uuid:      m.Uuid,
			Abteilung: 0,
			Bahn:      0,
		})
		if err != nil {
			webfw.APIError(w, webfw.InternalError(err.Error()))
			return
		}
	}

	if webfw.IsHtmxRequest(r) {
		return
	} else {
		webfw.JSON(w, r, "Losung erfolgreich!")
	}
}

func SetStartnummern(w http.ResponseWriter, r *http.Request) {
	if err := service.SetStartnummern(r.Context()); err != nil {
		webfw.APIError(w, webfw.InternalError(fmt.Sprintf("Error while setting startnummern: %s", err.Error())))
		return
	}

	webfw.JSON(w, r, "Startnummern vergeben!")
}

func SetZeitplan(w http.ResponseWriter, r *http.Request) {
	param := new(service.SetZeitplanParams)
	err := webfw.ParseBody(r, param)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	err = service.SetZeitplan(r.Context(), *param)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "Zeitplan gesetzt!")
}
