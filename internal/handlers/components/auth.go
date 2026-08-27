package components

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func LoginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		fieldErrors := make(map[string]string)
		if username == "" {
			fieldErrors["username"] = "Benutzername erforderlich"
		}
		if password == "" {
			fieldErrors["password"] = "Passwort erforderlich"
		}
		webfw.ErrorWithForm(w, r, ui_pages.Login("", fieldErrors), "Bitte alle Felder ausfüllen")
		return
	}

	u, err := crud.AuthLogin(r.Context(), crud.LoginParams{Username: username, Password: password})
	if err != nil {
		webfw.ErrorWithForm(w, r, ui_pages.Login("", nil), "Benutzername oder Passwort ist falsch")
		return
	}

	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    u.Jwt.Token,
		MaxAge:   72 * 60 * 60,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	webfw.SetRedirect(w, "/internal")
	w.WriteHeader(http.StatusOK)
}
