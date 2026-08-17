package server

import (
	"net/http"

	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/handlers/components"
	"github.com/bata94/RegattaApi/internal/handlers/pages"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
)

var (
	r            = newRouter()
	navBarConfig = ui_components.NewNavBarConfig()
)

func GetRouter() http.Handler {
	navBarConfig.Entries = []ui_components.NavBarEntry{
		{Name: "Livestream", URL: "/live"},
		{Name: "Ausschreibung", URL: "/ausschreibung"},
		{Name: "Zeitplan", URL: "/zeitplan"},
		{Name: "Meldeergebnis", URL: "/meldeergebnis"},
		{Name: "Ergebnisse", URL: "/ergebnisse"},
		{
			Name:         "Internes",
			URL:          "",
			RequiredCaps: []string{"allowed_logged_in"},
			SubEntries: []ui_components.NavBarEntry{
				{Name: "Profil", URL: "/internal/profil", RequiredCaps: []string{"allowed_logged_in"}},
				{Name: "Zeitnahme", URL: "/internal/zeitnahme", RequiredCaps: []string{"allowed_zeitnahme"}},
				{Name: "Startlisten", URL: "/internal/startlisten", RequiredCaps: []string{"allowed_startlisten"}},
				{Name: "Regattabüro", URL: "/internal/regattabuero", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Regattaleitung", URL: "/internal/regattaleitung", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Admin", URL: "/internal/admin", RequiredCaps: []string{"allowed_admin"}},
			},
		},
	}

	baseLayoutHandler("/metrics", pages.MetricsPage)

	r.Handle("GET", "/metricsApi", wrapHandler(components.MetricsAPIHandler, true))

	r.Handle("GET", "/assets/{file}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/"+r.URL.Path[len("/assets/"):])
	})
	r.Handle("GET", "/files/{file}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, config.C.Paths.FilesDir+r.URL.Path[len("/files/"):])
	})
	r.Handle("GET", "/public/{file}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, config.C.Paths.PublicDir+r.URL.Path[len("/public/"):])
	})

	// UI Handlers
	baseLayoutHandler("/", pages.Index)
	baseLayoutHandler("/live", pages.Livestream)
	baseLayoutHandler("/ausschreibung", pages.Ausschreibung)
	baseLayoutHandler("/zeitplan", pages.Zeitplan)
	baseLayoutHandler("/meldeergebnis", pages.Meldeergebnis)
	baseLayoutHandler("/ergebnisse", pages.Ergebnisse)
	baseLayoutHandler("/login", pages.Login)
	r.Handle("POST", "/login", wrapHandler(components.LoginPost, false))
	r.Handle("GET", "/logout", logoutHandler)

	baseLayoutHandler("/datenschutz", pages.Datenschutz)

	internalLayoutHandler("/internal", pages.InternalIndex)

	internalLayoutHandler("/internal/profil", pages.ProfilPage)
	r.Handle("GET", "/internal/profil/password/{uuid}", wrapHandler(components.ChangePasswordGet, true))
	r.Handle("PUT", "/internal/profil/password/{uuid}", wrapHandler(components.ChangePasswordPost, true))
	internalLayoutHandler("/internal/zeitnahme", pages.InternalZeitnahme)
	internalLayoutHandler("/internal/zeitnahme/start", pages.InternalZeitnahmeStart)
	internalLayoutHandler("/internal/zeitnahme/ziel", pages.InternalZeitnahmeZiel)
	internalLayoutHandler("/internal/zeitnahme/vorsortierung", pages.InternalZeitnahmeVorsortierung)
	internalLayoutHandler("/internal/zeitnahme/wenderichter", pages.InternalZeitnahmeWenderichter)

	internalLayoutHandler("/internal/startlisten", pages.InternalStartlisten)

	internalLayoutHandler("/internal/regattabuero", pages.InternalRegattabuero)
	internalLayoutHandler("/internal/regattabuero/vereinswahl", pages.InternalVereinswahl)
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/abmeldung", pages.InternalRegattabueroAbmeldung)
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/abmeldung/{m_uuid}", pages.InternalRegattabueroAbmeldungMeldung)
	r.Handle("DELETE", "/internal/regattabuero/{v_uuid}/abmeldung/{m_uuid}", wrapHandler(components.AbmeldungDelete, true))
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/ummeldung", pages.InternalRegattabueroUmmeldung)
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/ummeldung/{m_uuid}", pages.InternalRegattabueroUmmeldungMeldung)
	r.Handle("POST", "/internal/regattabuero/{v_uuid}/ummeldung/{m_uuid}", wrapHandler(components.UmmeldungPost, true))
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/nachmeldung", pages.InternalRegattabueroNachmeldung)
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/nachmeldung/{r_uuid}", pages.InternalRegattabueroNachmeldungRennen)
	r.Handle("POST", "/internal/regattabuero/{v_uuid}/nachmeldung/{r_uuid}", wrapHandler(components.NachmeldungPost, true))
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/nachmeldung/success/{m_uuid}", pages.InternalRegattabueroNachmeldungSuccess)
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/waage", pages.InternalRegattabueroWaageWahl)
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/waage/{a_uuid}", pages.InternalRegattabueroWaage)
	r.Handle("POST", "/internal/regattabuero/{v_uuid}/waage/{a_uuid}", wrapHandler(components.WaagePost, true))
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/startberechtigung", pages.InternalRegattabueroStartberechtigung)
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/startberechtigung/{a_uuid}", pages.InternalRegattabueroStartberechtigungAthlet)
	r.Handle("POST", "/internal/regattabuero/{v_uuid}/startberechtigung/{a_uuid}", wrapHandler(components.StartberechtigungPost, true))
	internalLayoutHandler("/internal/regattabuero/{v_uuid}/new_athlet", pages.InternalRegattabueroNewAthlet)
	r.Handle("POST", "/internal/regattabuero/{v_uuid}/new_athlet", wrapHandler(components.NewAthletPost, true))
	internalLayoutHandler("/internal/regattabuero/kasse", pages.InternalRegattabueroKasse)
	internalLayoutHandler("/internal/regattabuero/startnummernausgabe", pages.InternalRegattabueroStartnummernAusgabe)
	internalLayoutHandler("/internal/regattabuero/aenderungen_obleute", pages.InternalRegattabueroAenderungenObleute)
	internalLayoutHandler("/internal/regattabuero/setzungsverwaltung/aenderung", pages.InternalRegattaleitungSetzungAenderung)
	r.Handle("POST", "/internal/regattabuero/setzungsverwaltung/aenderung/rennen/{param}", wrapHandler(components.SetzungsVerwaltungAenderungRennenPost, true))
	internalLayoutHandler("/internal/regattabuero/setzungsverwaltung/aenderung/rennen/{param}", pages.InternalRegattaleitungSetzungAenderungRennen)
	internalLayoutHandler("/internal/regattabuero/startnummern/aenderung", pages.InternalRegattaleitungStartnummernAendernRennenWahl)

	internalLayoutHandler("/internal/regattaleitung", pages.InternalRegattaleitung)
	internalLayoutHandler("/internal/regattaleitung/drvupload", pages.InternalRegattaleitungDrvUpload)
	r.Handle("POST", "/internal/regattaleitung/drvupload", wrapHandler(components.DrvUploadPost, true))
	internalLayoutHandler("/internal/regattaleitung/setzungsverwaltung", pages.InternalRegattaleitungSetzung)
	internalLayoutHandler("/internal/regattaleitung/setzungsverwaltung/losung", pages.InternalRegattaleitungSetzungLosung)
	r.Handle("POST", "/internal/regattaleitung/setzungsverwaltung/losung", wrapHandler(components.SetzungsVerwaltungLosungPost, true))
	r.Handle("DELETE", "/internal/regattaleitung/setzungsverwaltung/losung", wrapHandler(components.SetzungsVerwaltungLosungDelete, true))
	internalLayoutHandler("/internal/regattaleitung/setzungsverwaltung/aenderung", pages.InternalRegattaleitungSetzungAenderung)
	r.Handle("POST", "/internal/regattaleitung/setzungsverwaltung/aenderung/rennen/{param}", wrapHandler(components.SetzungsVerwaltungAenderungRennenPost, true))
	internalLayoutHandler("/internal/regattaleitung/setzungsverwaltung/aenderung/rennen/{param}", pages.InternalRegattaleitungSetzungAenderungRennen)

	internalLayoutHandler("/internal/regattaleitung/pausen", pages.InternalRegattaleitungPausen)
	r.Handle("GET", "/internal/regattaleitung/pausen/new/{nach_rennen_uuid}", wrapHandler(components.PausenNew, true))
	r.Handle("POST", "/internal/regattaleitung/pausen", wrapHandler(components.PausenPost, true))
	r.Handle("DELETE", "/internal/regattaleitung/pausen/{id}", wrapHandler(components.PausenDelete, true))

	internalLayoutHandler("/internal/regattaleitung/zeitplan", pages.InternalRegattaleitungZeitplan)
	r.Handle("POST", "/internal/regattaleitung/zeitplan", wrapHandler(components.ZeitplanPost, true))
	internalLayoutHandler("/internal/regattaleitung/startnummern", pages.InternalRegattaleitungStartnummern)
	internalLayoutHandler("/internal/regattaleitung/startnummern/verteilen", pages.InternalRegattaleitungStartnummernVerteilen)
	r.Handle("POST", "/internal/regattaleitung/startnummern/verteilen", wrapHandler(components.StartnummernVerteilenPost, true))
	r.Handle("DELETE", "/internal/regattaleitung/startnummern/verteilen", wrapHandler(components.StartnummernVerteilenDelete, true))
	internalLayoutHandler("/internal/regattaleitung/startnummern/bereich", pages.InternalRegattaleitungStartnummernBereich)
	internalLayoutHandler("/internal/regattaleitung/startnummern/aenderung", pages.InternalRegattaleitungStartnummernAendernRennenWahl)
	internalLayoutHandler("/internal/regattaleitung/startnummern/aenderung/{r_uuid}", pages.InternalRegattaleitungStartnummernAendernMeldungsWahl)
	internalLayoutHandler("/internal/regattaleitung/startnummern/aenderung/{r_uuid}/{m_uuid}", pages.InternalRegattaleitungStartnummernAendern)
	r.Handle("POST", "/internal/regattaleitung/startnummern/aenderung/{r_uuid}/{m_uuid}", wrapHandler(components.StartnummernAendernPost, true))
	internalLayoutHandler("/internal/regattaleitung/pdf_meldeergebnis", pages.InternalRegattaleitungPdfMeldeergebnis)
	r.Handle("POST", "/internal/regattaleitung/pdf_meldeergebnis", wrapHandler(components.PdfMeldeergebnisPost, true))
	internalLayoutHandler("/internal/regattaleitung/vereine", pages.InternalRegattaleitungVereinsverwaltung)
	internalLayoutHandler("/internal/regattaleitung/email", pages.InternalRegattaleitungEmail)
	r.Handle("POST", "/internal/regattaleitung/email", wrapHandler(components.EmailSendPost, true))

	internalLayoutHandler("/internal/admin", pages.InternalAdmin)

	internalLayoutHandler("/internal/admin/users", pages.InternalAdminUsers)
	r.Handle("GET", "/internal/admin/user/{uuid}", wrapHandler(components.UserEditNew, true))
	r.Handle("POST", "/internal/admin/user/{uuid}", wrapHandler(components.UserEditNewPost, true))
	internalLayoutHandler("/internal/admin/usergroups", pages.InternalAdminUserGroups)
	r.Handle("GET", "/internal/admin/usergroups/{uuid}", wrapHandler(components.UserGroupEditNew, true))
	r.Handle("POST", "/internal/admin/usergroups/{uuid}", wrapHandler(components.UserGroupEditNewPost, true))
	internalLayoutHandler("/internal/admin/email_queue", pages.InternalAdminEmailQueue)
	r.Handle("POST", "/internal/admin/email_queue/{uuid}/retry", wrapHandler(components.EmailQueueRetry, true))
	r.Handle("DELETE", "/internal/admin/email_queue/{uuid}", wrapHandler(components.EmailQueueDelete, true))

	// Pure HTMX UI Components
	r.Handle("GET", "/comp/image", wrapUIHandler(components.ImageComponent))
	r.Handle("GET", "/comp/zeitplan/{wettkampf}", templHandler(components.ZeitplanCollapseBody))
	r.Handle("GET", "/comp/ausschreibung/{wettkampf}", templHandler(components.AusschreibungRennenCollapseBody))
	r.Handle("GET", "/comp/meldeergebnis/{wettkampf}", templHandler(components.MeldeergebnisCollapseBody))

	r.Handle("GET", "/comp/internal/rennen_tab/{wettkampf}", templHandler(components.RennenTab))

	// API Handlers - Public (no auth required)
	r.Handle("POST", "/api/auth/login", wrapHandler(api_v1.Login, false))
	r.Handle("POST", "/api/auth/logout", wrapHandler(api_v1.Logout, false))
	r.Handle("GET", "/api/auth/valid", wrapHandler(api_v1.AuthValidate, false))

	// API Handlers - Protected (auth required)
	r.Handle("GET", "/api/auth/me", wrapHandler(api_v1.AuthMe, true))
	r.Handle("GET", "/api/v1/test", wrapHandler(api_v1.TestHandler, true))

	r.Handle("GET", "/api/v1/athlet/startberechtigung", wrapHandler(api_v1.GetAthletStartberechtigung, true))
	r.Handle("PUT", "/api/v1/athlet/startberechtigung", wrapHandler(api_v1.UpdateAthletStartberechtigung, true))
	r.Handle("GET", "/api/v1/athlet/waage", wrapHandler(api_v1.GetAthletWaage, true))
	r.Handle("PUT", "/api/v1/athlet/waage", wrapHandler(api_v1.UpdateAthletWaage, true))
	r.Handle("GET", "/api/v1/athlet", wrapHandler(api_v1.GetAllAthlet, true))
	r.Handle("GET", "/api/v1/athlet/{uuid}", wrapHandler(api_v1.GetAthlet, true))
	r.Handle("POST", "/api/v1/athlet", wrapHandler(api_v1.CreateAthlet, true))

	r.Handle("POST", "/api/v1/buero/abmeldung", wrapHandler(api_v1.PostAbmeldung, true))
	r.Handle("POST", "/api/v1/buero/ummeldung", wrapHandler(api_v1.PostUmmeldung, true))
	r.Handle("POST", "/api/v1/buero/nachmeldung", wrapHandler(api_v1.PostNachmeldung, true))
	r.Handle("POST", "/api/v1/buero/startnummernausgabe", wrapHandler(api_v1.StartnummernAusgabe, true))
	r.Handle("POST", "/api/v1/buero/startnummernwechsel", wrapHandler(api_v1.StartnummernWechsel, true))
	r.Handle("POST", "/api/v1/buero/kasse/einzahlung", wrapHandler(api_v1.KasseEinzahlung, true))
	r.Handle("POST", "/api/v1/buero/kasse/rechnung/all", wrapHandler(api_v1.KasseCreateRechnungAllVereine, true))
	r.Handle("GET", "/api/v1/buero/kasse/rechnung/{uuid}", wrapHandler(api_v1.KasseCreateRechnungHTML, true))
	r.Handle("POST", "/api/v1/buero/kasse/rechnung/{uuid}", wrapHandler(api_v1.KasseCreateRechnungPDF, true))

	r.Handle("GET", "/api/v1/leitung/pdfFooter", wrapHandler(api_v1.GetPdfFooter, true))
	r.Handle("GET", "/api/v1/leitung/meldeergebnis", wrapHandler(api_v1.GetMeldeergebnisHtml, false))
	r.Handle("GET", "/api/v1/leitung/meldeergebnis/list", wrapHandler(api_v1.GetMeldeergebnisList, true))
	r.Handle("GET", "/api/v1/leitung/meldeergebnis/{filename}", wrapHandler(api_v1.GetMeldeergebnisFilename, true))
	r.Handle("POST", "/api/v1/leitung/meldeergebnis", wrapHandler(api_v1.GenerateMeldeergebnis, true))
	r.Handle("GET", "/api/v1/leitung/ergebnis", wrapHandler(api_v1.GenerateErgebnisHtml, true))
	r.Handle("POST", "/api/v1/leitung/ergebnis", wrapHandler(api_v1.GenerateErgebnis, true))
	r.Handle("POST", "/api/v1/leitung/drv_meldung_upload", wrapHandler(api_v1.DrvMeldungUpload, true))
	r.Handle("POST", "/api/v1/leitung/SetzungsLosung", wrapHandler(api_v1.SetzungsLosung, true))
	r.Handle("POST", "/api/v1/leitung/SetzungsLosung/reset", wrapHandler(api_v1.ResetSetzung, true))
	r.Handle("POST", "/api/v1/leitung/SetZeitplan", wrapHandler(api_v1.SetZeitplan, true))
	r.Handle("POST", "/api/v1/leitung/SetStartnummern", wrapHandler(api_v1.SetStartnummern, true))

	r.Handle("GET", "/api/v1/meldung", wrapHandler(api_v1.GetAllMeldung, true))
	r.Handle("GET", "/api/v1/meldung/{uuid}", wrapHandler(api_v1.GetMeldung, true))
	r.Handle("POST", "/api/v1/meldung/updateSetzungBatch", wrapHandler(api_v1.UpdateSetzungBatch, true))
	r.Handle("POST", "/api/v1/meldung/abmeldung", wrapHandler(api_v1.PostAbmeldung, true))
	r.Handle("POST", "/api/v1/meldung/ummeldung", wrapHandler(api_v1.PostUmmeldung, true))
	r.Handle("POST", "/api/v1/meldung/nachmeldung", wrapHandler(api_v1.PostNachmeldung, true))
	r.Handle("GET", "/api/v1/meldung/verein/{uuid}", wrapHandler(api_v1.GetAllMeldungForVerein, true))

	r.Handle("GET", "/api/v1/pause", wrapHandler(api_v1.GetAllPausen, true))
	r.Handle("GET", "/api/v1/pause/{id}", wrapHandler(api_v1.GetPause, true))
	r.Handle("DELETE", "/api/v1/pause/{id}", wrapHandler(api_v1.DeletePause, true))
	r.Handle("POST", "/api/v1/pause", wrapHandler(api_v1.CreatePause, true))
	r.Handle("PUT", "/api/v1/pause", wrapHandler(api_v1.UpdatePause, true))

	r.Handle("GET", "/api/v1/rennen", wrapHandler(api_v1.GetAllRennen, true))
	r.Handle("GET", "/api/v1/rennen/{uuid}", wrapHandler(api_v1.GetRennen, true))

	r.Handle("GET", "/api/v1/users", wrapHandler(api_v1.GetAllUsers, true))
	r.Handle("GET", "/api/v1/users/{uuid}", wrapHandler(api_v1.GetUser, true))
	r.Handle("GET", "/api/v1/users/name/{name}", wrapHandler(api_v1.GetUserByName, true))
	r.Handle("POST", "/api/v1/users", wrapHandler(api_v1.CreateUser, true))
	r.Handle("GET", "/api/v1/users/group", wrapHandler(api_v1.GetAllUsersGroups, true))
	r.Handle("GET", "/api/v1/users/group/{uuid}", wrapHandler(api_v1.GetUsersGroup, true))
	r.Handle("GET", "/api/v1/users/group/name/{name}", wrapHandler(api_v1.GetUsersGroupByName, true))

	r.Handle("GET", "/api/v1/verein", wrapHandler(api_v1.GetAllVerein, true))
	r.Handle("GET", "/api/v1/verein/{uuid}", wrapHandler(api_v1.GetVerein, true))
	r.Handle("GET", "/api/v1/verein/{uuid}/athlet", wrapHandler(api_v1.GetAllAthletenForVerein, true))
	r.Handle("GET", "/api/v1/verein/{uuid}/waage", wrapHandler(api_v1.GetAllAthletenForVereinWaage, true))
	r.Handle("GET", "/api/v1/verein/{uuid}/startberechtigung", wrapHandler(api_v1.GetAllAthletenForVereinMissStartber, true))

	r.Handle("GET", "/api/v1/zeitnahme/ziel", wrapHandler(api_v1.GetOpenZeitnahmeZiel, true))
	r.Handle("POST", "/api/v1/zeitnahme/start", wrapHandler(api_v1.PostZeitnahmeStart, true))
	r.Handle("GET", "/api/v1/zeitnahme/openStarts", wrapHandler(api_v1.GetOpenStarts, true))
	r.Handle("POST", "/api/v1/zeitnahme/genErgebnis", wrapHandler(api_v1.GenerateEndZeit, true))

	r.Handle("GET", "/ws/zeitnahme", api_v1.HandleZeitnahmeWS)

	go handlers.RunHub()

	return r
}
