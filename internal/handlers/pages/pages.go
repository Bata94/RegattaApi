package pages

import (
	"log/slog"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func Index(c *handler.Context) (templ.Component, error) {
	return ui_pages.Index(), nil
}

func Livestream(c *handler.Context) (templ.Component, error) {
	return ui_pages.Livestream(), nil
}

func Ausschreibung(c *handler.Context) (templ.Component, error) {
	return ui_pages.Ausschreibung(), nil
}

func Zeitplan(c *handler.Context) (templ.Component, error) {
	return ui_pages.Zeitplan(), nil
}

func Meldeergebnis(c *handler.Context) (templ.Component, error) {
	return ui_pages.Meldeergebnis(), nil
}

func Ergebnisse(c *handler.Context) (templ.Component, error) {
	return ui_pages.Ergebnisse(), nil
}

func Login(c *handler.Context) (templ.Component, error) {
	return ui_pages.Login("", nil), nil
}

func Datenschutz(c *handler.Context) (templ.Component, error) {
	return ui_pages.Datenschutz(), nil
}

func InternalZeitnahme(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalZeitnahme(), nil
}

func InternalZeitnahmeZiel(c *handler.Context) (templ.Component, error) {
	return ui_pages.ZeitnahmeZiel(), nil
}

func InternalStartlisten(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalStartlisten(), nil
}

func InternalRegattabuero(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattabuero(), nil
}

func InternalRegattaleitung(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitung(), nil
}

func InternalRegattaleitungDrvUpload(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungDrvFileUpload(""), nil
}

func InternalRegattaleitungSetzung(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungSetzung(), nil
}

func InternalRegattaleitungSetzungLosung(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungSetzungLosung(), nil
}

func InternalRegattaleitungSetzungAenderung(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungSetzungAenderung(), nil
}

func InternalRegattaleitungPausen(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungPausen(), nil
}

func InternalRegattaleitungZeitplan(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungZeitplan("", nil), nil
}

func InternalRegattaleitungStartnummern(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungStartnummern(), nil
}

func InternalRegattaleitungStartnummernVerteilen(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungStartnummernVerteilen(), nil
}

func InternalRegattaleitungStartnummernBereich(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungStartnummernBereich(), nil
}

func InternalRegattaleitungStartnummernAendernRennenWahl(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungStartnummernAendernRennenWahl(), nil
}

func InternalRegattaleitungStartnummernAendernMeldungsWahl(c *handler.Context) (templ.Component, error) {
	rUuidStr := c.Param("r_uuid")
	rUuid, err := uuid.Parse(rUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	r, err := crud.GetRennen(c.Request.Context(), rUuid)
	if err != nil {
		return nil, handler.NotFound("Rennen nicht gefunden")
	}

	for i := range r.Meldungen {
		r.Meldungen[i].Rennen = &r
	}

	return ui_pages.InternalRegattaleitungStartnummernAendernMeldungsWahl(r), nil
}

func InternalRegattaleitungStartnummernAendern(c *handler.Context) (templ.Component, error) {
	// rUuidStr := c.Param("r_uuid")
	mUuidStr := c.Param("m_uuid")

	// rUuid, err := uuid.Parse(rUuidStr)
	// if err != nil {
	// 	return nil, handler.NotAcceptable("Invalid UUID")
	// }
	mUuid, err := uuid.Parse(mUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	m, err := crud.GetMeldung(c.Request.Context(), mUuid)
	if err != nil {
		return nil, handler.NotFound("Meldung nicht gefunden")
	}

	return ui_pages.InternalRegattaleitungStartnummernAendern(m, nil), nil
}

func InternalRegattaleitungPdfMeldeergebnis(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungPdfMeldeergebnis(false), nil
}

func InternalRegattaleitungVereinsverwaltung(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungVereinsverwaltung(), nil
}

func InternalRegattaleitungEmail(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalRegattaleitungEmailIndex(), nil
}

func InternalAdmin(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalAdmin(), nil
}

func InternalAdminUsers(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalAdminUsers(), nil
}

func InternalAdminUserGroups(c *handler.Context) (templ.Component, error) {
	return ui_pages.InternalAdminUserGroups(), nil
}

func InternalIndex(c *handler.Context) (templ.Component, error) {
	userToken, ok := c.GetLocals("user").(*jwt.Token)
	if !ok {
		return nil, handler.Unauthorized("Nicht angemeldet")
	}

	claims := userToken.Claims.(jwt.MapClaims)
	var capabilities []string

	capFields := []string{"allowed_admin", "allowed_zeitnahme", "allowed_startlisten", "allowed_regattabuero", "allowed_regattaleitung"}

	for _, field := range capFields {
		if val, exists := claims[field]; exists && val == true {
			capabilities = append(capabilities, field)
		}
	}

	return ui_pages.InternalIndex(capabilities), nil
}

func ProfilPage(c *handler.Context) (templ.Component, error) {
	userToken, ok := c.GetLocals("user").(*jwt.Token)
	if !ok {
		return nil, handler.Unauthorized("Nicht angemeldet")
	}

	claims := userToken.Claims.(jwt.MapClaims)
	userUuidStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, handler.Unauthorized("Invalid token")
	}
	username, ok := claims["username"].(string)
	if !ok {
		return nil, handler.Unauthorized("Invalid token")
	}

	userGroup := ""
	if ug, ok := claims["user_group_name"].(string); ok {
		userGroup = ug
	}

	var capabilities []string
	capFields := []string{
		"allowed_zeitnahme",
		"allowed_startlisten",
		"allowed_regattabuero",
		"allowed_regattaleitung",
		"allowed_admin",
	}

	for _, field := range capFields {
		if val, exists := claims[field]; exists && val == true {
			capabilities = append(capabilities, field)
		}
	}

	userUuid, err := uuid.Parse(userUuidStr)
	if err != nil {
		return nil, handler.Unauthorized("Invalid token")
	}

	data := ui_pages.ProfilData{
		Uuid:         userUuid,
		Username:     username,
		UserGroup:    userGroup,
		Capabilities: capabilities,
	}

	return ui_pages.Profil(data), nil
}

func MetricsPage(c *handler.Context) (templ.Component, error) {
	secret := config.C.Auth.JWTSecret

	tokenString := c.Cookie("auth_token")
	if tokenString == "" {
		return nil, handler.Unauthorized("Nicht angemeldet")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, handler.Unauthorized("Ungültiges oder abgelaufenes Token")
	}

	claims := token.Claims.(jwt.MapClaims)
	admin, ok := claims["allowed_admin"].(bool)
	if !ok || !admin {
		return nil, handler.Forbidden("Keine Admin-Berechtigung")
	}

	return ui_pages.Metrics(), nil
}

func InternalVereinswahl(c *handler.Context) (templ.Component, error) {
	next := c.GetQueryParam("next")
	if next == "" {
		return nil, handler.BadRequest("Next param is required")
	}
	title := c.GetQueryParam("title")
	nextUrl := "/internal/regattabuero/%s/" + next

	var (
		vereine []crud.Verein
		err     error
	)

	switch next {
	case "waage":
		vereine, err = crud.GetForAllVereineMissingAthlet(crud.Waage)
	case "startberechtigung":
		vereine, err = crud.GetForAllVereineMissingAthlet(crud.Startberechtigt)
	default:
		vereine, err = crud.GetAllVerein(c.Request.Context())
	}
	if err != nil {
		return nil, handler.InternalError("Fehler beim Laden der Vereine")
	}

	return ui_pages.InternalVereinswahl(nextUrl, title, vereine), nil
}

func InternalRegattabueroAbmeldung(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	meldungen, err := crud.GetAllMeldungForVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading meldungen")
	}
	return ui_pages.InternalRegattabueroAbmeldung(verein, meldungen), nil
}

func InternalRegattabueroAbmeldungMeldung(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	meldungUuidStr := c.Param("m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	meldung, err := crud.GetMeldung(c.Request.Context(), meldungUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading meldung")
	}

	if meldung.VereinUuid != verein.Uuid {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	return ui_pages.InternalRegattabueroAbmeldungMeldung(verein, meldung), nil
}

func InternalRegattabueroUmmeldung(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	meldungen, err := crud.GetAllMeldungForVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading meldungen")
	}
	return ui_pages.InternalRegattabueroUmmeldung(verein, meldungen), nil
}

func InternalRegattabueroUmmeldungMeldung(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	meldungUuidStr := c.Param("m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	meldung, err := crud.GetMeldung(c.Request.Context(), meldungUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading meldung")
	}

	if meldung.VereinUuid != verein.Uuid {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	athleten, err := crud.GetAllAthletenForVerein(c.Request.Context(), verein.Uuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading athleten")
	}

	// TODO: Filter only viable athleten
	return ui_pages.InternalRegattabueroUmmeldungMeldung(verein, meldung, athleten, "", nil), nil
}

func InternalRegattabueroNachmeldung(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	return ui_pages.InternalRegattabueroNachmeldung(verein), nil
}

func InternalRegattabueroNachmeldungRennen(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		slog.Error("Invalid verein UUID", "err", err)
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	rennenUuidStr := c.Param("r_uuid")
	rennenUuid, err := uuid.Parse(rennenUuidStr)
	if err != nil {
		slog.Error("Invalid rennen UUID", "err", err)
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		slog.Error("Error loading verein", "err", err)
		return nil, handler.InternalError("Error while loading verein")
	}
	rennen, err := crud.GetRennen(c.Request.Context(), rennenUuid)
	if err != nil {
		slog.Error("Error loading rennen", "err", err)
		return nil, handler.InternalError("Error while loading rennen")
	}

	athleten, err := crud.GetAllAthletenForVerein(c.Request.Context(), verein.Uuid)
	if err != nil {
		slog.Error("Error loading athleten", "err", err)
		return nil, handler.InternalError("Error while loading athleten")
	}

	return ui_pages.InternalRegattabueroNachmeldungMeldung(verein, rennen, athleten, "", nil), nil
}

func InternalRegattabueroNachmeldungSuccess(c *handler.Context) (templ.Component, error) {
	meldungUuidStr := c.Param("m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	m, err := crud.GetMeldung(c.Request.Context(), meldungUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading meldung")
	}
	return ui_pages.InternalRegattabueroNachmeldungSuccess(m), nil
}

func InternalRegattabueroWaageWahl(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	athleten, err := crud.GetAllAthletenForVereinWaage(c.Request.Context(), verein.Uuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading athleten")
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	return ui_pages.InternalRegattabueroWaageWahl(verein, athleten), nil
}

func InternalRegattabueroWaage(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	athletUuidStr := c.Param("a_uuid")
	athletUuid, err := uuid.Parse(athletUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	athlet, err := crud.GetAthlet(c.Request.Context(), athletUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading athlet")
	}

	if athlet.VereinUuid != verein.Uuid {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	return ui_pages.InternalRegattabueroWaage(athlet, "", nil), nil
}

func InternalRegattabueroStartberechtigung(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	athleten, err := crud.GetAllAthletenForVereinMissStartber(c.Request.Context(), verein.Uuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading athleten")
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	return ui_pages.InternalRegattabueroStartberechtigungWahl(verein, athleten), nil
}

func InternalRegattabueroStartberechtigungAthlet(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	athletUuidStr := c.Param("a_uuid")
	athletUuid, err := uuid.Parse(athletUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	athlet, err := crud.GetAthlet(c.Request.Context(), athletUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading athlet")
	}

	if athlet.VereinUuid != verein.Uuid {
		return nil, handler.NotAcceptable("Invalid UUID")
	}

	// TODO: Implement Form Errors
	return ui_pages.InternalRegattabueroStartberechtigung(athlet, "", nil), nil
}

func InternalRegattabueroNewAthlet(c *handler.Context) (templ.Component, error) {
	vereinUuidStr := c.Param("v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		return nil, handler.NotAcceptable("Invalid UUID")
	}
	verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
	if err != nil {
		return nil, handler.InternalError("Error while loading verein")
	}
	return ui_pages.InternalRegattabueroNewAthlet(verein, "", nil), nil
}

func InternalRegattaleitungSetzungAenderungRennen(c *handler.Context) (templ.Component, error) {
	paramStr := c.Param("param")
	slog.Debug("Param", "value", paramStr)

	slog.Debug("Param is a Rennen UUID")
	rUuid, err := uuid.Parse(paramStr)
	if err != nil {
		slog.Error("Error", "err", err)
		templ.Handler(ui_pages.Error(404, "Rennen nicht gefunden")).ServeHTTP(c.Writer, c.Request)
		return nil, nil
	}
	return ui_pages.InternalRegattaleitungSetzungAenderungRennen(rUuid), nil
}
