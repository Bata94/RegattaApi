package api_v1

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func Login(w http.ResponseWriter, r *http.Request) {
	loginParams := new(crud.LoginParams)
	if err := webfw.ParseBody(r, loginParams); err != nil {
		webfw.APIError(w, webfw.BadRequest(err.Error()))
		return
	}

	u, err := crud.AuthLogin(r.Context(), *loginParams)
	if err != nil {
		webfw.APIError(w, webfw.Unauthorized("Invalid credentials"))
		return
	}

	webfw.SetCookie(w, r, "auth_token", u.Jwt.Token, 72*60*60)
	w.Header().Set("HX-Redirect", "/internal")
	webfw.JSON(w, r, u)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	webfw.DeleteCookie(w, r, "auth_token")
	webfw.JSON(w, r, "Logout successful!")
}

func AuthValidate(w http.ResponseWriter, r *http.Request) {
	webfw.JSON(w, r, "Auth successful!")
}

func AuthMe(w http.ResponseWriter, r *http.Request) {
	user := webfw.GetUser(r)
	if user == nil {
		webfw.APIError(w, webfw.Unauthorized("Not authenticated"))
		return
	}

	claims := webfw.GetUserClaims(r)
	uuidStr := claims["user_id"].(string)

	userID, err := crud.ParseUUID(uuidStr)
	if err != nil {
		webfw.APIError(w, webfw.Unauthorized("Invalid token"))
		return
	}

	u, err := crud.GetUser(r.Context(), userID)
	if err != nil {
		webfw.APIError(w, webfw.InternalError(err.Error()))
		return
	}

	webfw.JSON(w, r, *u)
}
