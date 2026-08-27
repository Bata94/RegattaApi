package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	uLs, err := crud.GetAllUsers(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, uLs)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	u, err := crud.GetUser(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, *u)
}

func GetUserByName(w http.ResponseWriter, r *http.Request) {
	name := webfw.Param(r, "name")
	if name == "" {
		webfw.APIError(w, webfw.BadRequest("name parameter required"))
		return
	}

	u, err := crud.GetUserByUsername(r.Context(), name)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, *u)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	uParams := new(crud.CreateUserParams)
	err := webfw.ParseBody(r, &uParams)
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	u, err := crud.CreateUser(r.Context(), *uParams)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, u)
}

func GetAllUsersGroups(w http.ResponseWriter, r *http.Request) {
	ugLs, err := crud.GetAllUsersGroups(r.Context())
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, ugLs)
}

func GetUsersGroup(w http.ResponseWriter, r *http.Request) {
	uuid, err := webfw.GetUUID(r, "uuid")
	if err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	ug, err := crud.GetUsersGroup(r.Context(), uuid)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, ug)
}

func GetUsersGroupByName(w http.ResponseWriter, r *http.Request) {
	name := webfw.Param(r, "name")
	if name == "" {
		webfw.APIError(w, webfw.BadRequest("name parameter required"))
		return
	}

	ug, err := crud.GetUsersGroupByName(r.Context(), name)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, ug)
}
