package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	profil "github.com/bata94/RegattaApi/internal/templates/pages/profil"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func ChangePasswordGet(w http.ResponseWriter, r *http.Request) {
	userUuidStr := webfw.Param(r, "uuid")
	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	user, err := crud.GetUser(r.Context(), userUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "User not found")
		return
	}

	templ.Handler(profil.ChangePasswordDialogBody(*user, "", nil)).ServeHTTP(w, r)
}

func ChangePasswordPost(w http.ResponseWriter, r *http.Request) {
	targetUuidStr := webfw.Param(r, "uuid")
	targetUuid, err := uuid.Parse(targetUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	callerUuidStr := webfw.GetUserIDString(r)
	callerUuid, err := uuid.Parse(callerUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Unauthorized")
		return
	}

	isAdmin := webfw.HasCapability(r, "allowed_admin")
	if !isAdmin && callerUuid != targetUuid {
		webfw.ErrorToast(w, r, "You can only change your own password")
		return
	}

	user, err := crud.GetUser(r.Context(), targetUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "User not found")
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword1 := r.FormValue("new_password_1")
	newPassword2 := r.FormValue("new_password_2")

	fieldErrors := make(map[string]string)
	topMsg := ""
	isSelf := callerUuid == targetUuid
	if newPassword1 == "" {
		fieldErrors["new_password_1"] = "Neues Passwort erforderlich"
		topMsg = "Bitte alle Felder ausfüllen"
	}
	if newPassword2 == "" {
		fieldErrors["new_password_2"] = "Neues Passwort erneut erforderlich"
		topMsg = "Bitte alle Felder ausfüllen"
	}
	if newPassword1 != "" && newPassword2 != "" && newPassword1 != newPassword2 {
		fieldErrors["new_password_1"] = "Passwörter stimmen nicht überein"
		fieldErrors["new_password_2"] = "Passwörter stimmen nicht überein"
		topMsg = "Passwörter stimmen nicht überein"
	}
	if isSelf && currentPassword == "" {
		fieldErrors["current_password"] = "Aktuelles Passwort erforderlich"
		topMsg = "Bitte alle Felder ausfüllen"
	}
	if isSelf && currentPassword != "" && !crud.CheckPasswordHash(currentPassword, user.HashedPassword) {
		fieldErrors["current_password"] = "Aktuelles Passwort ist falsch"
		topMsg = "Aktuelles Passwort ist falsch"
	}
	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, profil.ChangePasswordDialogBody(*user, "", fieldErrors), topMsg)
		return
	}

	err = crud.UpdatePassword(r.Context(), targetUuid, newPassword1)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while updating password")
		return
	}

	webfw.SetRedirect(w, "/internal/profil")
	w.WriteHeader(http.StatusOK)
}
