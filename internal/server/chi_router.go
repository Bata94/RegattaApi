package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/handlers/components"
	"github.com/bata94/RegattaApi/internal/handlers/pages"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
	ui_layouts "github.com/bata94/RegattaApi/internal/templates/layout"
	"github.com/bata94/RegattaApi/pkg/webfw"
	webfw_middleware "github.com/bata94/RegattaApi/pkg/webfw/middleware"
	"github.com/go-chi/chi/v5"
)

func newChiRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(webfw_middleware.Recovery())
	r.Use(webfw_middleware.Compression())
	r.Use(webfw_middleware.Logging())
	r.Use(webfw_middleware.CORS())
	r.Use(webfw_middleware.RateLimit())

	setupPublicRoutes(r)
	setupInternalRoutes(r)
	setupAPIRoutes(r)
	setupComponentRoutes(r)
	setupMetricsRoutes(r)

	go handlers.RunHub()

	return r
}

func setupPublicRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(webfw_middleware.Timeout(60 * time.Second))

		r.Get("/", adaptPageHandler(pages.Index))
		r.Get("/live", adaptPageHandler(pages.Livestream))
		r.Get("/ausschreibung", adaptPageHandler(pages.Ausschreibung))
		r.Get("/zeitplan", adaptPageHandler(pages.Zeitplan))
		r.Get("/meldeergebnis", adaptPageHandler(pages.Meldeergebnis))
		r.Get("/ergebnisse", adaptPageHandler(pages.Ergebnisse))
		r.Get("/login", adaptPageHandler(pages.Login))
		r.Get("/datenschutz", adaptPageHandler(pages.Datenschutz))

		r.Post("/login", components.LoginPost)
		r.Get("/logout", logoutHandler)

		r.Get("/assets/*", staticFileHandler("/assets/", "./assets/", func(string) string {
			return "public, max-age=3600"
		}))
		r.Get("/files/*", staticFileHandler("/files/", config.C.Paths.FilesDir, func(string) string {
			return "public, max-age=3600"
		}))
		r.Get("/public/*", staticFileHandler("/public/", config.C.Paths.PublicDir, func(name string) string {
			if strings.HasSuffix(name, ".webp") {
				return "public, max-age=31536000, immutable"
			}
			return "public, max-age=3600"
		}))
	})
}

func setupInternalRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(webfw_middleware.Auth())
		r.Use(webfw_middleware.Timeout(60 * time.Second))

		r.Get("/internal", adaptPageHandler(pages.InternalIndex))
		r.Get("/internal/profil", adaptPageHandler(pages.ProfilPage))
		r.Get("/internal/profil/password/{uuid}", components.ChangePasswordGet)
		r.Put("/internal/profil/password/{uuid}", components.ChangePasswordPost)

		r.Get("/internal/zeitnahme", adaptPageHandler(pages.InternalZeitnahme))
		r.Get("/internal/zeitnahme/start", adaptPageHandler(pages.InternalZeitnahmeStart))
		r.Get("/internal/zeitnahme/ziel", adaptPageHandler(pages.InternalZeitnahmeZiel))
		r.Get("/internal/zeitnahme/vorsortierung", adaptPageHandler(pages.InternalZeitnahmeVorsortierung))
		r.Get("/internal/zeitnahme/wenderichter", adaptPageHandler(pages.InternalZeitnahmeWenderichter))

		r.Get("/internal/startlisten", adaptPageHandler(pages.InternalStartlisten))

		r.Get("/internal/regattabuero", adaptPageHandler(pages.InternalRegattabuero))
		r.Get("/internal/regattabuero/vereinswahl", adaptPageHandler(pages.InternalVereinswahl))
		r.Get("/internal/regattabuero/{v_uuid}/abmeldung", adaptPageHandler(pages.InternalRegattabueroAbmeldung))
		r.Get("/internal/regattabuero/{v_uuid}/abmeldung/{m_uuid}", adaptPageHandler(pages.InternalRegattabueroAbmeldungMeldung))
		r.Delete("/internal/regattabuero/{v_uuid}/abmeldung/{m_uuid}", components.AbmeldungDelete)
		r.Get("/internal/regattabuero/{v_uuid}/ummeldung", adaptPageHandler(pages.InternalRegattabueroUmmeldung))
		r.Get("/internal/regattabuero/{v_uuid}/ummeldung/{m_uuid}", adaptPageHandler(pages.InternalRegattabueroUmmeldungMeldung))
		r.Post("/internal/regattabuero/{v_uuid}/ummeldung/{m_uuid}", components.UmmeldungPost)
		r.Get("/internal/regattabuero/{v_uuid}/nachmeldung", adaptPageHandler(pages.InternalRegattabueroNachmeldung))
		r.Get("/internal/regattabuero/{v_uuid}/nachmeldung/{r_uuid}", adaptPageHandler(pages.InternalRegattabueroNachmeldungRennen))
		r.Post("/internal/regattabuero/{v_uuid}/nachmeldung/{r_uuid}", components.NachmeldungPost)
		r.Get("/internal/regattabuero/{v_uuid}/nachmeldung/success/{m_uuid}", adaptPageHandler(pages.InternalRegattabueroNachmeldungSuccess))
		r.Get("/internal/regattabuero/{v_uuid}/waage", adaptPageHandler(pages.InternalRegattabueroWaageWahl))
		r.Get("/internal/regattabuero/{v_uuid}/waage/{a_uuid}", adaptPageHandler(pages.InternalRegattabueroWaage))
		r.Post("/internal/regattabuero/{v_uuid}/waage/{a_uuid}", components.WaagePost)
		r.Get("/internal/regattabuero/{v_uuid}/startberechtigung", adaptPageHandler(pages.InternalRegattabueroStartberechtigung))
		r.Get("/internal/regattabuero/{v_uuid}/startberechtigung/{a_uuid}", adaptPageHandler(pages.InternalRegattabueroStartberechtigungAthlet))
		r.Post("/internal/regattabuero/{v_uuid}/startberechtigung/{a_uuid}", components.StartberechtigungPost)
		r.Get("/internal/regattabuero/{v_uuid}/new_athlet", adaptPageHandler(pages.InternalRegattabueroNewAthlet))
		r.Post("/internal/regattabuero/{v_uuid}/new_athlet", components.NewAthletPost)
		r.Get("/internal/regattabuero/kasse", adaptPageHandler(pages.InternalRegattabueroKasse))
		r.Get("/internal/regattabuero/startnummernausgabe", adaptPageHandler(pages.InternalRegattabueroStartnummernAusgabe))
		r.Get("/internal/regattabuero/aenderungen_obleute", adaptPageHandler(pages.InternalRegattabueroAenderungenObleute))
		r.Get("/internal/regattabuero/setzungsverwaltung/aenderung", adaptPageHandler(pages.InternalRegattaleitungSetzungAenderung))
		r.Post("/internal/regattabuero/setzungsverwaltung/aenderung/rennen/{param}", components.SetzungsVerwaltungAenderungRennenPost)
		r.Get("/internal/regattabuero/setzungsverwaltung/aenderung/rennen/{param}", adaptPageHandler(pages.InternalRegattaleitungSetzungAenderungRennen))
		r.Get("/internal/regattabuero/startnummern/aenderung", adaptPageHandler(pages.InternalRegattaleitungStartnummernAendernRennenWahl))

		r.Get("/internal/regattaleitung", adaptPageHandler(pages.InternalRegattaleitung))
		r.Get("/internal/regattaleitung/drvupload", adaptPageHandler(pages.InternalRegattaleitungDrvUpload))
		r.Post("/internal/regattaleitung/drvupload", components.DrvUploadPost)
		r.Get("/internal/regattaleitung/setzungsverwaltung", adaptPageHandler(pages.InternalRegattaleitungSetzung))
		r.Get("/internal/regattaleitung/setzungsverwaltung/losung", adaptPageHandler(pages.InternalRegattaleitungSetzungLosung))
		r.Post("/internal/regattaleitung/setzungsverwaltung/losung", components.SetzungsVerwaltungLosungPost)
		r.Delete("/internal/regattaleitung/setzungsverwaltung/losung", components.SetzungsVerwaltungLosungDelete)
		r.Get("/internal/regattaleitung/setzungsverwaltung/aenderung", adaptPageHandler(pages.InternalRegattaleitungSetzungAenderung))
		r.Post("/internal/regattaleitung/setzungsverwaltung/aenderung/rennen/{param}", components.SetzungsVerwaltungAenderungRennenPost)
		r.Get("/internal/regattaleitung/setzungsverwaltung/aenderung/rennen/{param}", adaptPageHandler(pages.InternalRegattaleitungSetzungAenderungRennen))

		r.Get("/internal/regattaleitung/pausen", adaptPageHandler(pages.InternalRegattaleitungPausen))
		r.Get("/internal/regattaleitung/pausen/new/{nach_rennen_uuid}", components.PausenNew)
		r.Post("/internal/regattaleitung/pausen", components.PausenPost)
		r.Delete("/internal/regattaleitung/pausen/{id}", components.PausenDelete)

		r.Get("/internal/regattaleitung/zeitplan", adaptPageHandler(pages.InternalRegattaleitungZeitplan))
		r.Post("/internal/regattaleitung/zeitplan", components.ZeitplanPost)
		r.Get("/internal/regattaleitung/startnummern", adaptPageHandler(pages.InternalRegattaleitungStartnummern))
		r.Get("/internal/regattaleitung/startnummern/verteilen", adaptPageHandler(pages.InternalRegattaleitungStartnummernVerteilen))
		r.Post("/internal/regattaleitung/startnummern/verteilen", components.StartnummernVerteilenPost)
		r.Delete("/internal/regattaleitung/startnummern/verteilen", components.StartnummernVerteilenDelete)
		r.Get("/internal/regattaleitung/startnummern/bereich", adaptPageHandler(pages.InternalRegattaleitungStartnummernBereich))
		r.Post("/internal/regattaleitung/startnummern/bereich", components.StartnummernBereichPost)
		r.Get("/internal/regattaleitung/startnummern/aenderung", adaptPageHandler(pages.InternalRegattaleitungStartnummernAendernRennenWahl))
		r.Get("/internal/regattaleitung/startnummern/aenderung/{r_uuid}", adaptPageHandler(pages.InternalRegattaleitungStartnummernAendernMeldungsWahl))
		r.Get("/internal/regattaleitung/startnummern/aenderung/{r_uuid}/{m_uuid}", adaptPageHandler(pages.InternalRegattaleitungStartnummernAendern))
		r.Post("/internal/regattaleitung/startnummern/aenderung/{r_uuid}/{m_uuid}", components.StartnummernAendernPost)
		r.Get("/internal/regattaleitung/pdf_meldeergebnis", adaptPageHandler(pages.InternalRegattaleitungPdfMeldeergebnis))
		r.Post("/internal/regattaleitung/pdf_meldeergebnis", components.PdfMeldeergebnisPost)
		r.Get("/internal/regattaleitung/vereine", adaptPageHandler(pages.InternalRegattaleitungVereinsverwaltung))
		r.Get("/internal/regattaleitung/vereine/{uuid}", components.VereinEditNew)
		r.Post("/internal/regattaleitung/vereine/{uuid}", components.VereinEditNewPost)
		r.Delete("/internal/regattaleitung/vereine/{uuid}", components.VereinDelete)
		r.Get("/internal/regattaleitung/obleute", adaptPageHandler(pages.InternalRegattaleitungObleute))
		r.Get("/internal/regattaleitung/obleute/{uuid}", components.ObmannEditNew)
		r.Post("/internal/regattaleitung/obleute/{uuid}", components.ObmannEditNewPost)
		r.Delete("/internal/regattaleitung/obleute/{uuid}", components.ObmannDelete)
		r.Get("/internal/regattaleitung/email", adaptPageHandler(pages.InternalRegattaleitungEmail))
		r.Post("/internal/regattaleitung/email", components.EmailSendPost)

		r.Get("/internal/admin", adaptPageHandler(pages.InternalAdmin))
		r.Get("/internal/admin/users", adaptPageHandler(pages.InternalAdminUsers))
		r.Get("/internal/admin/user/{uuid}", components.UserEditNew)
		r.Post("/internal/admin/user/{uuid}", components.UserEditNewPost)
		r.Get("/internal/admin/usergroups", adaptPageHandler(pages.InternalAdminUserGroups))
		r.Get("/internal/admin/usergroups/{uuid}", components.UserGroupEditNew)
		r.Post("/internal/admin/usergroups/{uuid}", components.UserGroupEditNewPost)
		r.Get("/internal/admin/email_queue", adaptPageHandler(pages.InternalAdminEmailQueue))
		r.Post("/internal/admin/email_queue/{uuid}/retry", components.EmailQueueRetry)
		r.Delete("/internal/admin/email_queue/{uuid}", components.EmailQueueDelete)
	})
}

func setupAPIRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(webfw_middleware.ErrorHandler)
		r.Use(webfw_middleware.Timeout(30 * time.Second))

		r.Post("/api/auth/login", api_v1.Login)
		r.Post("/api/auth/logout", api_v1.Logout)
		r.Get("/api/auth/valid", api_v1.AuthValidate)

		r.Get("/api/auth/me", api_v1.AuthMe)
		r.Get("/api/v1/test", api_v1.TestHandler)

		r.Get("/api/v1/athlet/startberechtigung", api_v1.GetAthletStartberechtigung)
		r.Put("/api/v1/athlet/startberechtigung", api_v1.UpdateAthletStartberechtigung)
		r.Get("/api/v1/athlet/waage", api_v1.GetAthletWaage)
		r.Put("/api/v1/athlet/waage", api_v1.UpdateAthletWaage)
		r.Get("/api/v1/athlet", api_v1.GetAllAthlet)
		r.Get("/api/v1/athlet/{uuid}", api_v1.GetAthlet)
		r.Post("/api/v1/athlet", api_v1.CreateAthlet)

		r.Post("/api/v1/buero/abmeldung", api_v1.PostAbmeldung)
		r.Post("/api/v1/buero/ummeldung", api_v1.PostUmmeldung)
		r.Post("/api/v1/buero/nachmeldung", api_v1.PostNachmeldung)
		r.Post("/api/v1/buero/startnummernausgabe", api_v1.StartnummernAusgabe)
		r.Post("/api/v1/buero/startnummernwechsel", api_v1.StartnummernWechsel)
		r.Post("/api/v1/buero/kasse/einzahlung", api_v1.KasseEinzahlung)
		r.Post("/api/v1/buero/kasse/rechnung/all", api_v1.KasseCreateRechnungAllVereine)
		r.Get("/api/v1/buero/kasse/rechnung/{uuid}", api_v1.KasseCreateRechnungHTML)
		r.Post("/api/v1/buero/kasse/rechnung/{uuid}", api_v1.KasseCreateRechnungPDF)

		r.Get("/api/v1/leitung/pdfFooter", api_v1.GetPdfFooter)
		r.Get("/api/v1/leitung/meldeergebnis", api_v1.GetMeldeergebnisHtml)
		r.Get("/api/v1/leitung/meldeergebnis/list", api_v1.GetMeldeergebnisList)
		r.Get("/api/v1/leitung/meldeergebnis/{filename}", api_v1.GetMeldeergebnisFilename)
		r.Post("/api/v1/leitung/meldeergebnis", api_v1.GenerateMeldeergebnis)
		r.Get("/api/v1/leitung/ergebnis", api_v1.GenerateErgebnisHtml)
		r.Post("/api/v1/leitung/ergebnis", api_v1.GenerateErgebnis)
		r.Post("/api/v1/leitung/drv_meldung_upload", api_v1.DrvMeldungUpload)
		r.Post("/api/v1/leitung/SetzungsLosung", api_v1.SetzungsLosung)
		r.Post("/api/v1/leitung/SetzungsLosung/reset", api_v1.ResetSetzung)
		r.Post("/api/v1/leitung/SetZeitplan", api_v1.SetZeitplan)
		r.Post("/api/v1/leitung/SetStartnummern", api_v1.SetStartnummern)

		r.Get("/api/v1/meldung", api_v1.GetAllMeldung)
		r.Get("/api/v1/meldung/{uuid}", api_v1.GetMeldung)
		r.Post("/api/v1/meldung/updateSetzungBatch", api_v1.UpdateSetzungBatch)
		r.Post("/api/v1/meldung/abmeldung", api_v1.PostAbmeldung)
		r.Post("/api/v1/meldung/ummeldung", api_v1.PostUmmeldung)
		r.Post("/api/v1/meldung/nachmeldung", api_v1.PostNachmeldung)
		r.Get("/api/v1/meldung/verein/{uuid}", api_v1.GetAllMeldungForVerein)

		r.Get("/api/v1/pause", api_v1.GetAllPausen)
		r.Get("/api/v1/pause/{id}", api_v1.GetPause)
		r.Delete("/api/v1/pause/{id}", api_v1.DeletePause)
		r.Post("/api/v1/pause", api_v1.CreatePause)
		r.Put("/api/v1/pause", api_v1.UpdatePause)

		r.Get("/api/v1/rennen", api_v1.GetAllRennen)
		r.Get("/api/v1/rennen/{uuid}", api_v1.GetRennen)

		r.Get("/api/v1/users", api_v1.GetAllUsers)
		r.Get("/api/v1/users/{uuid}", api_v1.GetUser)
		r.Get("/api/v1/users/name/{name}", api_v1.GetUserByName)
		r.Post("/api/v1/users", api_v1.CreateUser)
		r.Get("/api/v1/users/group", api_v1.GetAllUsersGroups)
		r.Get("/api/v1/users/group/{uuid}", api_v1.GetUsersGroup)
		r.Get("/api/v1/users/group/name/{name}", api_v1.GetUsersGroupByName)

		r.Get("/api/v1/verein", api_v1.GetAllVerein)
		r.Get("/api/v1/verein/{uuid}", api_v1.GetVerein)
		r.Get("/api/v1/verein/{uuid}/athlet", api_v1.GetAllAthletenForVerein)
		r.Get("/api/v1/verein/{uuid}/waage", api_v1.GetAllAthletenForVereinWaage)
		r.Get("/api/v1/verein/{uuid}/startberechtigung", api_v1.GetAllAthletenForVereinMissStartber)

		r.Get("/api/v1/zeitnahme/ziel", api_v1.GetOpenZeitnahmeZiel)
		r.Post("/api/v1/zeitnahme/start", api_v1.PostZeitnahmeStart)
		r.Get("/api/v1/zeitnahme/openStarts", api_v1.GetOpenStarts)
		r.Post("/api/v1/zeitnahme/genErgebnis", api_v1.GenerateEndZeit)

		r.Get("/ws/zeitnahme", api_v1.HandleZeitnahmeWS)
	})
}

func setupComponentRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Get("/comp/image", components.ImageComponent)
		r.Get("/comp/zeitplan/{wettkampf}", components.ZeitplanCollapseBody)
		r.Get("/comp/ausschreibung/{wettkampf}", components.AusschreibungRennenCollapseBody)
		r.Get("/comp/meldeergebnis/{wettkampf}", components.MeldeergebnisCollapseBody)
		r.Get("/comp/internal/rennen_tab/{wettkampf}", components.RennenTab)
	})
}

func setupMetricsRoutes(r *chi.Mux) {
	r.Get("/metrics", adaptPageHandler(pages.MetricsPage))
	r.Get("/metricsApi", api_v1.MetricsApi)
}

type newPageHandler func(w http.ResponseWriter, r *http.Request) templ.Component

func adaptPageHandler(h newPageHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := h(w, r)
		if page == nil {
			return
		}

		if webfw.IsHtmxRequest(r) {
			templ.Handler(page).ServeHTTP(w, r)
		} else {
			caps := webfw.GetCapabilities(r)
			loggedIn := webfw.IsLoggedIn(r)
			navbarCfg := ui_components.NavBarConfig{
				Entries:  navBarConfig.Entries,
				UserCaps: caps,
				LoggedIn: loggedIn,
			}
			templ.Handler(ui_layouts.BaseLayout(page, navbarCfg)).ServeHTTP(w, r)
		}
	}
}

func staticFileHandler(prefix, dir string, cacheControl func(name string) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len(prefix):]
		if cc := cacheControl(name); cc != "" {
			w.Header().Set("Cache-Control", cc)
		}
		http.ServeFile(w, r, dir+name)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
