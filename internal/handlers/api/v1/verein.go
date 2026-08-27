package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func GetAllVerein(w http.ResponseWriter, r *http.Request) {
	vLs, err := crud.GetAllVerein(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, vLs)
}

func GetVerein(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	v, err := crud.GetVerein(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, v)
}

func GetAllAthletenForVerein(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	aLs, err := crud.GetAllAthletenForVerein(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, aLs)
}

func GetAllAthletenForVereinMissStartber(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	aLs, err := crud.GetAllAthletenForVereinMissStartber(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, aLs)
}

func GetAllAthletenForVereinWaage(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	aLs, err := crud.GetAllAthletenForVereinWaage(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, aLs)
}
