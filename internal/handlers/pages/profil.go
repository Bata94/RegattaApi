package pages

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
	profil "github.com/bata94/RegattaApi/internal/templates/pages/profil"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/golang-jwt/jwt/v5"
)

func ProfilPage(w http.ResponseWriter, r *http.Request) templ.Component {
	userToken := webfw.GetUser(r)
	if userToken == nil {
		webfw.HandlePageError(w, r, webfw.Unauthorized("Nicht angemeldet"))
		return nil
	}

	claims := userToken.Claims.(jwt.MapClaims)
	userUuidStr, ok := claims["user_id"].(string)
	if !ok {
		webfw.HandlePageError(w, r, webfw.Unauthorized("Invalid token"))
		return nil
	}
	username, ok := claims["username"].(string)
	if !ok {
		webfw.HandlePageError(w, r, webfw.Unauthorized("Invalid token"))
		return nil
	}

	userGroup := ""
	if ug, ok := claims["user_group_name"].(string); ok {
		userGroup = ug
	}

	var capabilities []string
	if capsRaw, ok := claims["capabilities"].([]any); ok {
		for _, c := range capsRaw {
			if s, ok := c.(string); ok {
				capabilities = append(capabilities, s)
			}
		}
	}

	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.Unauthorized("Invalid token"))
		return nil
	}

	data := profil.ProfilData{
		Uuid:         userUuid,
		Username:     username,
		UserGroup:    userGroup,
		Capabilities: capabilities,
	}

	return profil.Profil(data)
}

func MetricsPage(w http.ResponseWriter, r *http.Request) templ.Component {
	secret := config.C.Auth.JWTSecret

	tokenString := webfw.Cookie(r, "auth_token")
	if tokenString == "" {
		webfw.HandlePageError(w, r, webfw.Unauthorized("Nicht angemeldet"))
		return nil
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		webfw.HandlePageError(w, r, webfw.Unauthorized("Ungültiges oder abgelaufenes Token"))
		return nil
	}

	claims := token.Claims.(jwt.MapClaims)
	caps, ok := claims["capabilities"].([]any)
	if !ok {
		webfw.HandlePageError(w, r, webfw.Forbidden("Keine Admin-Berechtigung"))
		return nil
	}
	hasAdmin := false
	for _, c := range caps {
		if s, ok := c.(string); ok && s == "allowed_admin" {
			hasAdmin = true
			break
		}
	}
	if !hasAdmin {
		webfw.HandlePageError(w, r, webfw.Forbidden("Keine Admin-Berechtigung"))
		return nil
	}

	return ui_pages.Metrics()
}
