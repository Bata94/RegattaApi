package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/sqlc"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func UserGroupEditNew(w http.ResponseWriter, r *http.Request) {
	var ug sqlc.UsersGroup
	if webfw.Param(r, "uuid") == "" {
		webfw.ErrorToast(w, r, "UserGroup not found")
		return
	} else if webfw.Param(r, "uuid") == "new" {
		ug = sqlc.UsersGroup{}
	} else {
		grpUUID, err := uuid.Parse(webfw.Param(r, "uuid"))
		if err != nil {
			webfw.ErrorToast(w, r, "Invalid UUID")
			return
		}
		ug, err = crud.GetUsersGroupsMinimal(r.Context(), grpUUID)
		if err != nil {
			webfw.ErrorToast(w, r, "UserGroup not found")
			return
		}
	}

	templ.Handler(ui_components.UserGroupEdit(ug, "", nil)).ServeHTTP(w, r)
}

func UserGroupEditNewPost(w http.ResponseWriter, r *http.Request) {
	uuidStr := webfw.Param(r, "uuid")
	var groupUuid uuid.UUID

	if uuidStr == "new" {
		var err error
		groupUuid, err = uuid.NewV7()
		if err != nil {
			webfw.ErrorToast(w, r, "Bad Request")
			return
		}
	} else {
		var err error
		groupUuid, err = uuid.Parse(uuidStr)
		if err != nil {
			webfw.ErrorToast(w, r, "Bad Request")
			return
		}
	}

	name := r.FormValue("name")

	fieldErrors := make(map[string]string)
	if name == "" {
		fieldErrors["name"] = "Gruppenname erforderlich"
	}

	capList := []struct {
		formName string
		value    sqlc.UserCapability
	}{
		{"allowed_admin", sqlc.UserCapabilityAllowedAdmin},
		{"allowed_zeitnahme", sqlc.UserCapabilityAllowedZeitnahme},
		{"allowed_startlisten", sqlc.UserCapabilityAllowedStartlisten},
		{"allowed_regattabuero", sqlc.UserCapabilityAllowedRegattabuero},
		{"allowed_regattaleitung", sqlc.UserCapabilityAllowedRegattaleitung},
	}
	var capabilities []sqlc.UserCapability
	for _, cf := range capList {
		if r.FormValue(cf.formName) == "on" {
			capabilities = append(capabilities, cf.value)
		}
	}

	ug := sqlc.UsersGroup{
		Uuid:         groupUuid,
		Name:         name,
		Capabilities: capabilities,
	}

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, ui_components.UserGroupEdit(ug, "", fieldErrors), "Bitte alle Pflichtfelder ausfüllen")
		return
	}

	if uuidStr == "new" {
		_, err := crud.CreateUserGroup(r.Context(), sqlc.CreateUserGroupParams{
			Name:         name,
			Capabilities: capabilities,
		})
		if err != nil {
			webfw.ErrorWithForm(w, r, ui_components.UserGroupEdit(ug, "", nil), "Fehler beim Erstellen der Nutzergruppe")
			return
		}
	} else {
		err := crud.UpdateUserGroup(r.Context(), groupUuid, sqlc.UpdateUserGroupParams{
			Name:         name,
			Capabilities: capabilities,
		})
		if err != nil {
			webfw.ErrorWithForm(w, r, ui_components.UserGroupEdit(ug, "", nil), "Fehler beim Aktualisieren der Nutzergruppe")
			return
		}
	}

	webfw.SetRedirect(w, "/internal/admin/usergroups")
	w.WriteHeader(http.StatusOK)
}
