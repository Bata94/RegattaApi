package api_v1

import (
	"net/http"
	"strings"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

type NewAthletParams struct {
	VereinUUID      string `json:"verein_uuid"`
	Name            string `json:"name"`
	Vorname         string `json:"vorname"`
	Jahrgang        string `json:"jahrgang"`
	Startberechtigt bool   `json:"startberechtigt"`
	Geschlecht      string `json:"geschlecht"`
}

func GetAthlet(w http.ResponseWriter, r *http.Request) {
	id, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	a, err := crud.GetAthletMinimal(r.Context(), id)
	if err != nil {
		webfw.APIError(w, webfw.NotFound("Athlet nicht gefunden"))
		return
	}

	webfw.JSON(w, r, a)
}

func GetAllAthlet(w http.ResponseWriter, r *http.Request) {
	aLs, err := crud.GetAllAthlet(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, aLs)
}

func CreateAthlet(w http.ResponseWriter, r *http.Request) {
	aParams := new(NewAthletParams)
	err := webfw.ParseBody(r, &aParams)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	vereinUuid, err := uuid.Parse(aParams.VereinUUID)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	var geschlecht sqlc.Geschlecht
	aParams.Geschlecht = strings.ToLower(aParams.Geschlecht)
	switch aParams.Geschlecht {
	case "m":
		geschlecht = sqlc.GeschlechtM
	case "f", "w":
		geschlecht = sqlc.GeschlechtW
	case "x":
		geschlecht = sqlc.GeschlechtX
	}

	a, err := crud.CreateAthlet(r.Context(), sqlc.CreateAthletParams{
		Uuid:            uuid.NewV7(),
		VereinUuid:      vereinUuid,
		Name:            aParams.Name,
		Vorname:         aParams.Vorname,
		Jahrgang:        aParams.Jahrgang,
		Startberechtigt: aParams.Startberechtigt,
		Geschlecht:      geschlecht,
	})
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, a)
}

type UpdateAthletStartberechtigungParams struct {
	Uuid            string `json:"uuid"`
	Startberechtigt bool   `json:"startberechtigt"`
}

func UpdateAthletStartberechtigung(w http.ResponseWriter, r *http.Request) {
	p := new(UpdateAthletStartberechtigungParams)
	err := webfw.ParseBody(r, p)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	id, err := uuid.Parse(p.Uuid)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	ath, err := crud.GetAthletMinimal(r.Context(), id)
	if err != nil {
		webfw.APIError(w, webfw.NotFound("Athlet nicht gefunden"))
		return
	}

	err = ath.UpdateStartberechtigung(r.Context(), p.Startberechtigt)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "Athlet erfolgreich angepasst!")
}

func GetAthletWaage(w http.ResponseWriter, r *http.Request) {
	ls, err := crud.GetForAllVereineMissingAthlet(r.Context(), crud.Waage)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	webfw.JSON(w, r, ls)
}

func GetAthletStartberechtigung(w http.ResponseWriter, r *http.Request) {
	ls, err := crud.GetForAllVereineMissingAthlet(r.Context(), crud.Startberechtigt)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}
	webfw.JSON(w, r, ls)
}

type UpdateAthletWaageParams struct {
	Uuid    string `json:"uuid"`
	Gewicht int    `json:"gewicht"`
}

func UpdateAthletWaage(w http.ResponseWriter, r *http.Request) {
	p := new(UpdateAthletWaageParams)
	err := webfw.ParseBody(r, p)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	id, err := uuid.Parse(p.Uuid)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	ath, err := crud.GetAthletMinimal(r.Context(), id)
	if err != nil {
		webfw.APIError(w, webfw.NotFound("Athlet nicht gefunden"))
		return
	}

	err = ath.UpdateGewicht(r.Context(), p.Gewicht)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, "Athlet erfolgreich angepasst!")
}
