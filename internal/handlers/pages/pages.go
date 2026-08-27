package pages

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
	admin "github.com/bata94/RegattaApi/internal/templates/pages/admin"
	dashboard "github.com/bata94/RegattaApi/internal/templates/pages/dashboard"
	profil "github.com/bata94/RegattaApi/internal/templates/pages/profil"
	regattabuero "github.com/bata94/RegattaApi/internal/templates/pages/regattabuero"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	startlisten "github.com/bata94/RegattaApi/internal/templates/pages/startlisten"
	vereinswahl "github.com/bata94/RegattaApi/internal/templates/pages/vereinswahl"
	zeitnahme "github.com/bata94/RegattaApi/internal/templates/pages/zeitnahme"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func Index(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Index()
}

func Livestream(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Livestream()
}

func Ausschreibung(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Ausschreibung()
}

func Zeitplan(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Zeitplan()
}

func Meldeergebnis(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Meldeergebnis()
}

func Ergebnisse(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Ergebnisse()
}

func Login(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Login("", nil)
}

func Datenschutz(w http.ResponseWriter, r *http.Request) templ.Component {
	return ui_pages.Datenschutz()
}

func InternalZeitnahme(w http.ResponseWriter, r *http.Request) templ.Component {
	return zeitnahme.Zeitnahme()
}

func InternalZeitnahmeZiel(w http.ResponseWriter, r *http.Request) templ.Component {
	return zeitnahme.Ziel()
}

func InternalZeitnahmeVorsortierung(w http.ResponseWriter, r *http.Request) templ.Component {
	return zeitnahme.Vorsortierung()
}

func InternalZeitnahmeWenderichter(w http.ResponseWriter, r *http.Request) templ.Component {
	return zeitnahme.Wenderichter()
}

func InternalZeitnahmeStart(w http.ResponseWriter, r *http.Request) templ.Component {
	rennen, err := crud.GetAllRennenWithAthlet(r.Context(), crud.GetAllRennenParams{
		GetMeldungen: true,
		GetAthleten:  true,
		ShowEmpty:    false,
		ShowStarted:  false,
	})
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Fehler beim Laden der Rennen"))
		return nil
	}

	for i := range rennen {
		for j := range rennen[i].Meldungen {
			rennen[i].Meldungen[j].Rennen = nil
		}
	}

	return zeitnahme.Start(rennen)
}

func InternalStartlisten(w http.ResponseWriter, r *http.Request) templ.Component {
	return startlisten.Startlisten()
}

func InternalRegattabuero(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattabuero.Dashboard()
}

func InternalRegattaleitung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Dashboard()
}

func InternalRegattaleitungDrvUpload(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.DrvFileUpload("")
}

func InternalRegattaleitungSetzung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Setzung()
}

func InternalRegattaleitungSetzungLosung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.SetzungLosung()
}

func InternalRegattaleitungSetzungAenderung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.SetzungAenderung()
}

func InternalRegattaleitungPausen(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Pausen()
}

func InternalRegattaleitungZeitplan(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Zeitplan("", nil)
}

func InternalRegattaleitungStartnummern(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Startnummern()
}

func InternalRegattaleitungStartnummernVerteilen(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.StartnummernVerteilen()
}

func InternalRegattaleitungStartnummernBereich(w http.ResponseWriter, r *http.Request) templ.Component {
	b, err := crud.GetStartnummernBereich(r.Context())
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error loading startnummernbereich: "+err.Error()))
		return nil
	}
	return regattaleitung.StartnummernBereich(b, nil)
}

func InternalRegattaleitungStartnummernAendernRennenWahl(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.StartnummernAendernRennenWahl()
}

func InternalRegattaleitungStartnummernAendernMeldungsWahl(w http.ResponseWriter, r *http.Request) templ.Component {
	rUuidStr := webfw.Param(r, "r_uuid")
	rUuid, err := uuid.Parse(rUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	rennen, err := crud.GetRennen(r.Context(), rUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotFound("Rennen nicht gefunden"))
		return nil
	}

	for i := range rennen.Meldungen {
		rennen.Meldungen[i].Rennen = &rennen
	}

	return regattaleitung.StartnummernAendernMeldungsWahl(rennen)
}

func InternalRegattaleitungStartnummernAendern(w http.ResponseWriter, r *http.Request) templ.Component {
	mUuidStr := webfw.Param(r, "m_uuid")

	mUuid, err := uuid.Parse(mUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	m, err := crud.GetMeldung(r.Context(), mUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotFound("Meldung nicht gefunden"))
		return nil
	}

	return regattaleitung.StartnummernAendern(m, nil)
}

func InternalRegattaleitungPdfMeldeergebnis(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.PdfMeldeergebnis(false)
}

func InternalRegattaleitungVereinsverwaltung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Vereinsverwaltung()
}

func InternalRegattaleitungObleute(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Obleute()
}

func InternalRegattaleitungEmail(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.EmailCompose("", "", "", nil, false, nil, nil)
}

func InternalAdmin(w http.ResponseWriter, r *http.Request) templ.Component {
	return admin.Dashboard()
}

func InternalAdminUsers(w http.ResponseWriter, r *http.Request) templ.Component {
	return admin.Users()
}

func InternalAdminUserGroups(w http.ResponseWriter, r *http.Request) templ.Component {
	return admin.UserGroups()
}

func InternalAdminEmailQueue(w http.ResponseWriter, r *http.Request) templ.Component {
	return admin.EmailQueue()
}

func InternalIndex(w http.ResponseWriter, r *http.Request) templ.Component {
	caps := webfw.GetCapabilities(r)
	if caps == nil {
		caps = []string{}
	}
	return dashboard.Dashboard(caps)
}

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

func InternalVereinswahl(w http.ResponseWriter, r *http.Request) templ.Component {
	next := webfw.Query(r, "next")
	if next == "" {
		webfw.HandlePageError(w, r, webfw.BadRequest("Next param is required"))
		return nil
	}
	title := webfw.Query(r, "title")
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
		vereine, err = crud.GetAllVerein(r.Context())
	}
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Fehler beim Laden der Vereine"))
		return nil
	}

	return vereinswahl.Vereinswahl(nextUrl, title, vereine)
}

func InternalRegattabueroAbmeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	meldungen, err := crud.GetAllMeldungForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldungen"))
		return nil
	}
	return regattabuero.Abmeldung(verein, meldungen)
}

func InternalRegattabueroAbmeldungMeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	meldung, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldung"))
		return nil
	}

	if meldung.VereinUuid != verein.Uuid {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	return regattabuero.AbmeldungMeldung(verein, meldung)
}

func InternalRegattabueroUmmeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	meldungen, err := crud.GetAllMeldungForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldungen"))
		return nil
	}
	return regattabuero.Ummeldung(verein, meldungen)
}

func InternalRegattabueroUmmeldungMeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	meldung, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldung"))
		return nil
	}

	if meldung.VereinUuid != verein.Uuid {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	athleten, err := crud.GetAllAthletenForVerein(r.Context(), verein.Uuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athleten"))
		return nil
	}

	return regattabuero.UmmeldungMeldung(verein, meldung, athleten, "", nil)
}

func InternalRegattabueroNachmeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	return regattabuero.Nachmeldung(verein)
}

func InternalRegattabueroNachmeldungRennen(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		slog.Error("Invalid verein UUID", "err", err)
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	rennenUuidStr := webfw.Param(r, "r_uuid")
	rennenUuid, err := uuid.Parse(rennenUuidStr)
	if err != nil {
		slog.Error("Invalid rennen UUID", "err", err)
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		slog.Error("Error loading verein", "err", err)
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	rennen, err := crud.GetRennen(r.Context(), rennenUuid)
	if err != nil {
		slog.Error("Error loading rennen", "err", err)
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading rennen"))
		return nil
	}

	athleten, err := crud.GetAllAthletenForVerein(r.Context(), verein.Uuid)
	if err != nil {
		slog.Error("Error loading athleten", "err", err)
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athleten"))
		return nil
	}

	return regattabuero.NachmeldungMeldung(verein, rennen, athleten, "", nil)
}

func InternalRegattabueroNachmeldungSuccess(w http.ResponseWriter, r *http.Request) templ.Component {
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	m, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldung"))
		return nil
	}
	return regattabuero.NachmeldungSuccess(m)
}

func InternalRegattabueroWaageWahl(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	athleten, err := crud.GetAllAthletenForVereinWaage(r.Context(), verein.Uuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athleten"))
		return nil
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	return regattabuero.WaageWahl(verein, athleten)
}

func InternalRegattabueroWaage(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	athletUuidStr := webfw.Param(r, "a_uuid")
	athletUuid, err := uuid.Parse(athletUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	athlet, err := crud.GetAthlet(r.Context(), athletUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athlet"))
		return nil
	}

	if athlet.VereinUuid != verein.Uuid {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	return regattabuero.Waage(athlet, "", nil)
}

func InternalRegattabueroStartberechtigung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	athleten, err := crud.GetAllAthletenForVereinMissStartber(r.Context(), verein.Uuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athleten"))
		return nil
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	return regattabuero.StartberechtigungWahl(verein, athleten)
}

func InternalRegattabueroStartberechtigungAthlet(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	athletUuidStr := webfw.Param(r, "a_uuid")
	athletUuid, err := uuid.Parse(athletUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	athlet, err := crud.GetAthlet(r.Context(), athletUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athlet"))
		return nil
	}

	if athlet.VereinUuid != verein.Uuid {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	return regattabuero.Startberechtigung(athlet, "", nil)
}

func InternalRegattabueroNewAthlet(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	return regattabuero.NewAthlet(verein, "", nil)
}

func InternalRegattabueroKasse(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattabuero.Kasse()
}

func InternalRegattabueroStartnummernAusgabe(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattabuero.StartnummernAusgabe()
}

func InternalRegattabueroAenderungenObleute(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattabuero.AenderungenObleute()
}

func InternalRegattaleitungSetzungAenderungRennen(w http.ResponseWriter, r *http.Request) templ.Component {
	paramStr := webfw.Param(r, "param")
	slog.Debug("Param", "value", paramStr)

	slog.Debug("Param is a Rennen UUID")
	rUuid, err := uuid.Parse(paramStr)
	if err != nil {
		slog.Error("Error", "err", err)
		webfw.HandlePageError(w, r, webfw.NotFound("Rennen nicht gefunden"))
		return nil
	}
	return regattaleitung.SetzungAenderungRennen(rUuid)
}
