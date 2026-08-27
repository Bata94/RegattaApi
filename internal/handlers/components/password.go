package components

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	profil "github.com/bata94/RegattaApi/internal/templates/pages/profil"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
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

	currentPassword := r.FormValue("current_password")
	newPassword1 := r.FormValue("new_password_1")
	newPassword2 := r.FormValue("new_password_2")

	fieldErrors := make(map[string]string)
	topMsg := ""
	if currentPassword == "" {
		fieldErrors["current_password"] = "Aktuelles Passwort erforderlich"
		topMsg = "Bitte alle Felder ausfüllen"
	}
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
	if currentPassword != "" && !crud.CheckPasswordHash(currentPassword, user.HashedPassword) {
		fieldErrors["current_password"] = "Aktuelles Passwort ist falsch"
		topMsg = "Aktuelles Passwort ist falsch"
	}
	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, profil.ChangePasswordDialogBody(*user, "", fieldErrors), topMsg)
		return
	}

	err = crud.UpdatePassword(r.Context(), userUuid, newPassword1)
	if err != nil {
		webfw.ErrorToast(w, r, "Error while updating password")
		return
	}

	webfw.SetRedirect(w, "/internal/profil")
	w.WriteHeader(http.StatusOK)
}
