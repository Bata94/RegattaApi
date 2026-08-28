package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func UserEditNew(w http.ResponseWriter, r *http.Request) {
	var u *crud.User
	if webfw.Param(r, "uuid") == "" {
		webfw.ErrorToast(w, r, "User not found")
		return
	} else if webfw.Param(r, "uuid") == "new" {
		u = &crud.User{
			User:      sqlc.User{},
			UserGroup: &sqlc.UsersGroup{},
		}
	} else {
		userUUID, err := uuid.Parse(webfw.Param(r, "uuid"))
		if err != nil {
			webfw.ErrorToast(w, r, "Invalid UUID")
			return
		}
		u, err = crud.GetUser(r.Context(), userUUID)
		if err != nil {
			webfw.ErrorToast(w, r, "User not found")
			return
		}
	}

	templ.Handler(ui_components.UserEdit(*u, "", nil)).ServeHTTP(w, r)
}

func UserEditNewPost(w http.ResponseWriter, r *http.Request) {
	var (
		u        *crud.User
		userUuid uuid.UUID
		err      error
	)

	uuidStr := webfw.Param(r, "uuid")
	if uuidStr == "new" {
		userUuid = uuid.NewV7()
	} else {
		userUuid, err = uuid.Parse(uuidStr)
	}

	username := r.FormValue("username")
	groupUuid, errGroupUuid := uuid.Parse(r.FormValue("user_group_uuid"))
	isNotActive := r.FormValue("is_not_active") == "on"

	if err != nil || errGroupUuid != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	fieldErrors := make(map[string]string)
	if username == "" {
		fieldErrors["username"] = "Benutzername erforderlich"
	}
	if groupUuid == uuid.Nil {
		fieldErrors["user_group_uuid"] = "Nutzergruppe erforderlich"
	}
	if uuidStr == "new" {
		password := r.FormValue("password")
		if password == "" {
			fieldErrors["password"] = "Passwort erforderlich"
		}
	}
	if len(fieldErrors) > 0 {
		u = &crud.User{
			Uuid:     userUuid,
			Username: username,
			IsActive: !isNotActive,
			UserGroup: &sqlc.UsersGroup{
				Uuid: groupUuid,
			},
		}
		webfw.ErrorWithForm(w, r, ui_components.UserEdit(*u, "", fieldErrors), "Bitte alle Pflichtfelder ausfüllen")
		return
	}

	if uuidStr == "new" {
		u = &crud.User{
			Uuid:     userUuid,
			Username: username,
			IsActive: !isNotActive,
			UserGroup: &sqlc.UsersGroup{
				Uuid: groupUuid,
			},
		}
		_, err = crud.CreateUser(r.Context(), crud.CreateUserParams{
			GroupUuid: groupUuid,
			Username:  username,
			Password:  r.FormValue("password"),
		})
		if err != nil {
			webfw.ErrorWithForm(w, r, ui_components.UserEdit(*u, "", nil), "Error while creating user, Err: "+err.Error())
			return
		}

		webfw.SetRedirect(w, "/internal/admin/users")
		w.WriteHeader(http.StatusOK)
		return
	}

	u, err = crud.GetUser(r.Context(), userUuid)
	if err != nil {
		u = &crud.User{
			Uuid:     userUuid,
			Username: username,
			IsActive: !isNotActive,
			UserGroup: &sqlc.UsersGroup{
				Uuid: groupUuid,
			},
		}
		webfw.ErrorWithForm(w, r, ui_components.UserEdit(*u, "", nil), "Error while updating user, Err: "+err.Error())
		return
	}

	err = crud.UpdateUser(r.Context(), u.Uuid, crud.UpdateUserParams{
		Username:  username,
		IsActive:  r.FormValue("is_not_active") != "on",
		GroupUuid: groupUuid,
	})
	if err != nil {
		webfw.ErrorWithForm(w, r, ui_components.UserEdit(*u, "", nil), "Error while updating user, Err: "+err.Error())
		return
	}

	webfw.SetRedirect(w, "/internal/admin/users")
	w.WriteHeader(http.StatusOK)
}
