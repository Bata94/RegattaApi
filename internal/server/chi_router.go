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
	"github.com/golang-jwt/jwt/v5"
)

func newChiRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(webfw_middleware.Recovery())
	r.Use(webfw_middleware.Compression())
	r.Use(webfw_middleware.Logging())
	r.Use(webfw_middleware.CORS())
	r.Use(webfw_middleware.RateLimit())

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

	setupPublicRoutes(r)
	setupInternalRoutes(r)
	setupAPIRoutes(r)
	setupComponentRoutes(r)
	setupMetricsRoutes(r)
	setupWSRoutes(r)

	go handlers.RunHub()

	return r
}

func setupPublicRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(webfw_middleware.OptionalAuth())
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

		r.Get("/internal", adaptInternalPageHandler(pages.InternalIndex))
		r.Get("/internal/profil", adaptInternalPageHandler(pages.ProfilPage))
		r.Get("/internal/profil/password/{uuid}", components.ChangePasswordGet)
		r.Put("/internal/profil/password/{uuid}", components.ChangePasswordPost)

		r.Get("/internal/zeitnahme", adaptInternalPageHandler(pages.InternalZeitnahme))
		r.Get("/internal/zeitnahme/start", adaptInternalPageHandler(pages.InternalZeitnahmeStart))
		r.Get("/internal/zeitnahme/ziel", adaptInternalPageHandler(pages.InternalZeitnahmeZiel))
		r.Get("/internal/zeitnahme/vorsortierung", adaptInternalPageHandler(pages.InternalZeitnahmeVorsortierung))
		r.Get("/internal/zeitnahme/wenderichter", adaptInternalPageHandler(pages.InternalZeitnahmeWenderichter))

		r.Get("/internal/startlisten", adaptInternalPageHandler(pages.InternalStartlisten))

		r.Get("/internal/regattabuero", adaptInternalPageHandler(pages.InternalRegattabuero))
		r.Get("/internal/regattabuero/vereinswahl", adaptInternalPageHandler(pages.InternalVereinswahl))
		r.Get("/internal/regattabuero/{v_uuid}/abmeldung", adaptInternalPageHandler(pages.InternalRegattabueroAbmeldung))
		r.Get("/internal/regattabuero/{v_uuid}/abmeldung/{m_uuid}", adaptInternalPageHandler(pages.InternalRegattabueroAbmeldungMeldung))
		r.Delete("/internal/regattabuero/{v_uuid}/abmeldung/{m_uuid}", components.AbmeldungDelete)
		r.Get("/internal/regattabuero/{v_uuid}/ummeldung", adaptInternalPageHandler(pages.InternalRegattabueroUmmeldung))
		r.Get("/internal/regattabuero/{v_uuid}/ummeldung/{m_uuid}", adaptInternalPageHandler(pages.InternalRegattabueroUmmeldungMeldung))
		r.Post("/internal/regattabuero/{v_uuid}/ummeldung/{m_uuid}", components.UmmeldungPost)
		r.Get("/internal/regattabuero/{v_uuid}/nachmeldung", adaptInternalPageHandler(pages.InternalRegattabueroNachmeldung))
		r.Get("/internal/regattabuero/{v_uuid}/nachmeldung/{r_uuid}", adaptInternalPageHandler(pages.InternalRegattabueroNachmeldungRennen))
		r.Post("/internal/regattabuero/{v_uuid}/nachmeldung/{r_uuid}", components.NachmeldungPost)
		r.Get("/internal/regattabuero/{v_uuid}/nachmeldung/success/{m_uuid}", adaptInternalPageHandler(pages.InternalRegattabueroNachmeldungSuccess))
		r.Get("/internal/regattabuero/{v_uuid}/waage", adaptInternalPageHandler(pages.InternalRegattabueroWaageWahl))
		r.Get("/internal/regattabuero/{v_uuid}/waage/{a_uuid}", adaptInternalPageHandler(pages.InternalRegattabueroWaage))
		r.Post("/internal/regattabuero/{v_uuid}/waage/{a_uuid}", components.WaagePost)
		r.Get("/internal/regattabuero/{v_uuid}/startberechtigung", adaptInternalPageHandler(pages.InternalRegattabueroStartberechtigung))
		r.Get("/internal/regattabuero/{v_uuid}/startberechtigung/{a_uuid}", adaptInternalPageHandler(pages.InternalRegattabueroStartberechtigungAthlet))
		r.Post("/internal/regattabuero/{v_uuid}/startberechtigung/{a_uuid}", components.StartberechtigungPost)
		r.Get("/internal/regattabuero/{v_uuid}/new_athlet", adaptInternalPageHandler(pages.InternalRegattabueroNewAthlet))
		r.Post("/internal/regattabuero/{v_uuid}/new_athlet", components.NewAthletPost)
		r.Get("/internal/regattabuero/kasse", adaptInternalPageHandler(pages.InternalRegattabueroKasse))
		r.Get("/internal/regattabuero/startnummernausgabe", adaptInternalPageHandler(pages.InternalRegattabueroStartnummernAusgabe))
		r.Get("/internal/regattabuero/aenderungen_obleute", adaptInternalPageHandler(pages.InternalRegattabueroAenderungenObleute))
		r.Get("/internal/regattabuero/setzungsverwaltung/aenderung", adaptInternalPageHandler(pages.InternalRegattaleitungSetzungAenderung))
		r.Post("/internal/regattabuero/setzungsverwaltung/aenderung/rennen/{param}", components.SetzungsVerwaltungAenderungRennenPost)
		r.Get("/internal/regattabuero/setzungsverwaltung/aenderung/rennen/{param}", adaptInternalPageHandler(pages.InternalRegattaleitungSetzungAenderungRennen))
		r.Get("/internal/regattabuero/startnummern/aenderung", adaptInternalPageHandler(pages.InternalRegattaleitungStartnummernAendernRennenWahl))

		r.Get("/internal/regattaleitung", adaptInternalPageHandler(pages.InternalRegattaleitung))
		r.Get("/internal/regattaleitung/drvupload", adaptInternalPageHandler(pages.InternalRegattaleitungDrvUpload))
		r.Post("/internal/regattaleitung/drvupload", components.DrvUploadPost)
		r.Get("/internal/regattaleitung/setzungsverwaltung", adaptInternalPageHandler(pages.InternalRegattaleitungSetzung))
		r.Get("/internal/regattaleitung/setzungsverwaltung/losung", adaptInternalPageHandler(pages.InternalRegattaleitungSetzungLosung))
		r.Post("/internal/regattaleitung/setzungsverwaltung/losung", components.SetzungsVerwaltungLosungPost)
		r.Delete("/internal/regattaleitung/setzungsverwaltung/losung", components.SetzungsVerwaltungLosungDelete)
		r.Get("/internal/regattaleitung/setzungsverwaltung/aenderung", adaptInternalPageHandler(pages.InternalRegattaleitungSetzungAenderung))
		r.Post("/internal/regattaleitung/setzungsverwaltung/aenderung/rennen/{param}", components.SetzungsVerwaltungAenderungRennenPost)
		r.Get("/internal/regattaleitung/setzungsverwaltung/aenderung/rennen/{param}", adaptInternalPageHandler(pages.InternalRegattaleitungSetzungAenderungRennen))

		r.Get("/internal/regattaleitung/pausen", adaptInternalPageHandler(pages.InternalRegattaleitungPausen))
		r.Get("/internal/regattaleitung/pausen/new/{nach_rennen_uuid}", components.PausenNew)
		r.Post("/internal/regattaleitung/pausen", components.PausenPost)
		r.Delete("/internal/regattaleitung/pausen/{id}", components.PausenDelete)

		r.Get("/internal/regattaleitung/zeitplan", adaptInternalPageHandler(pages.InternalRegattaleitungZeitplan))
		r.Post("/internal/regattaleitung/zeitplan", components.ZeitplanPost)
		r.Get("/internal/regattaleitung/startnummern", adaptInternalPageHandler(pages.InternalRegattaleitungStartnummern))
		r.Get("/internal/regattaleitung/startnummern/verteilen", adaptInternalPageHandler(pages.InternalRegattaleitungStartnummernVerteilen))
		r.Post("/internal/regattaleitung/startnummern/verteilen", components.StartnummernVerteilenPost)
		r.Delete("/internal/regattaleitung/startnummern/verteilen", components.StartnummernVerteilenDelete)
		r.Get("/internal/regattaleitung/startnummern/bereich", adaptInternalPageHandler(pages.InternalRegattaleitungStartnummernBereich))
		r.Post("/internal/regattaleitung/startnummern/bereich", components.StartnummernBereichPost)
		r.Get("/internal/regattaleitung/startnummern/aenderung", adaptInternalPageHandler(pages.InternalRegattaleitungStartnummernAendernRennenWahl))
		r.Get("/internal/regattaleitung/startnummern/aenderung/{r_uuid}", adaptInternalPageHandler(pages.InternalRegattaleitungStartnummernAendernMeldungsWahl))
		r.Get("/internal/regattaleitung/startnummern/aenderung/{r_uuid}/{m_uuid}", adaptInternalPageHandler(pages.InternalRegattaleitungStartnummernAendern))
		r.Post("/internal/regattaleitung/startnummern/aenderung/{r_uuid}/{m_uuid}", components.StartnummernAendernPost)
		r.Get("/internal/regattaleitung/pdf_meldeergebnis", adaptInternalPageHandler(pages.InternalRegattaleitungPdfMeldeergebnis))
		r.Post("/internal/regattaleitung/pdf_meldeergebnis", components.PdfMeldeergebnisPost)
		r.Get("/internal/regattaleitung/vereine", adaptInternalPageHandler(pages.InternalRegattaleitungVereinsverwaltung))
		r.Get("/internal/regattaleitung/vereine/{uuid}", components.VereinEditNew)
		r.Post("/internal/regattaleitung/vereine/{uuid}", components.VereinEditNewPost)
		r.Delete("/internal/regattaleitung/vereine/{uuid}", components.VereinDelete)
		r.Get("/internal/regattaleitung/obleute", adaptInternalPageHandler(pages.InternalRegattaleitungObleute))
		r.Get("/internal/regattaleitung/obleute/{uuid}", components.ObmannEditNew)
		r.Post("/internal/regattaleitung/obleute/{uuid}", components.ObmannEditNewPost)
		r.Delete("/internal/regattaleitung/obleute/{uuid}", components.ObmannDelete)
		r.Get("/internal/regattaleitung/email", adaptInternalPageHandler(pages.InternalRegattaleitungEmail))
		r.Post("/internal/regattaleitung/email", components.EmailSendPost)

		r.Get("/internal/admin", adaptInternalPageHandler(pages.InternalAdmin))
		r.Get("/internal/admin/users", adaptInternalPageHandler(pages.InternalAdminUsers))
		r.Get("/internal/admin/user/{uuid}", components.UserEditNew)
		r.Post("/internal/admin/user/{uuid}", components.UserEditNewPost)
		r.Get("/internal/admin/usergroups", adaptInternalPageHandler(pages.InternalAdminUserGroups))
		r.Get("/internal/admin/usergroups/{uuid}", components.UserGroupEditNew)
		r.Post("/internal/admin/usergroups/{uuid}", components.UserGroupEditNewPost)
		r.Get("/internal/admin/email_queue", adaptInternalPageHandler(pages.InternalAdminEmailQueue))
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
	})

	r.Group(func(r chi.Router) {
		r.Use(webfw_middleware.ErrorHandler)
		r.Use(webfw_middleware.Auth())
		r.Use(webfw_middleware.Timeout(30 * time.Second))

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
	})
}

func setupComponentRoutes(r *chi.Mux) {
	r.Get("/comp/image", components.ImageComponent)

	r.Group(func(r chi.Router) {
		r.Use(webfw_middleware.Timeout(60 * time.Second))
		r.Use(requireHTMX)

		r.Get("/comp/zeitplan/{wettkampf}", components.ZeitplanCollapseBody)
		r.Get("/comp/ausschreibung/{wettkampf}", components.AusschreibungRennenCollapseBody)
		r.Get("/comp/meldeergebnis/{wettkampf}", components.MeldeergebnisCollapseBody)
		r.Get("/comp/internal/rennen_tab/{wettkampf}", components.RennenTab)
	})
}

func requireHTMX(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !webfw.IsHtmxRequest(r) {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setupMetricsRoutes(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(webfw_middleware.Auth())
		r.Get("/metrics", adaptPageHandler(pages.MetricsPage))
		r.Get("/metricsApi", api_v1.MetricsApi)
	})
}

func setupWSRoutes(r *chi.Mux) {
	r.Get("/ws/zeitnahme", api_v1.HandleZeitnahmeWS)
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

func adaptInternalPageHandler(h newPageHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := h(w, r)
		if page == nil {
			return
		}

		if webfw.IsHtmxRequest(r) {
			templ.Handler(page).ServeHTTP(w, r)
			return
		}

		caps := webfw.GetCapabilities(r)
		if caps == nil {
			caps = []string{}
		}

		username := ""
		userToken := webfw.GetUser(r)
		if userToken != nil {
			if claims, ok := userToken.Claims.(jwt.MapClaims); ok {
				if u, ok := claims["username"].(string); ok {
					username = u
				}
			}
		}
		if username == "" {
			username = "?"
		}

		path := r.URL.Path
		cats := buildSidebarCategories(path, caps)
		title := pageTitleFromPath(path)

		templ.Handler(ui_layouts.InternalLayout(page, cats, caps, username, title)).ServeHTTP(w, r)
	}
}

func buildSidebarCategories(currentPath string, caps []string) []ui_components.SidebarCategory {
	return []ui_components.SidebarCategory{
		{
			Name:         "Zeitnahme",
			URL:          "/internal/zeitnahme",
			RequiredCaps: []string{"allowed_zeitnahme"},
			IsActive:     strings.HasPrefix(currentPath, "/internal/zeitnahme"),
			Entries: []ui_components.SidebarEntry{
				{Name: "Startgericht", URL: "/internal/zeitnahme/start", RequiredCaps: []string{"allowed_zeitnahme"}},
				{Name: "Zielgericht", URL: "/internal/zeitnahme/ziel", RequiredCaps: []string{"allowed_zeitnahme"}},
				{Name: "Vorsortierung", URL: "/internal/zeitnahme/vorsortierung", RequiredCaps: []string{"allowed_zeitnahme"}},
				{Name: "Wenderichter", URL: "/internal/zeitnahme/wenderichter", RequiredCaps: []string{"allowed_zeitnahme"}},
			},
		},
		{
			Name:         "Startlisten",
			URL:          "/internal/startlisten",
			RequiredCaps: []string{"allowed_startlisten"},
			IsActive:     strings.HasPrefix(currentPath, "/internal/startlisten"),
			Entries: []ui_components.SidebarEntry{
				{Name: "Langstrecke", URL: "#", RequiredCaps: []string{"allowed_startlisten"}},
				{Name: "Slalom", URL: "#", RequiredCaps: []string{"allowed_startlisten"}},
				{Name: "Kurzstrecke", URL: "#", RequiredCaps: []string{"allowed_startlisten"}},
				{Name: "Staffel", URL: "#", RequiredCaps: []string{"allowed_startlisten"}},
				{Name: "Aktuelles Rennen", URL: "#", RequiredCaps: []string{"allowed_startlisten"}},
			},
		},
		{
			Name:         "Regattabüro",
			URL:          "/internal/regattabuero",
			RequiredCaps: []string{"allowed_regattabuero"},
			IsActive:     strings.HasPrefix(currentPath, "/internal/regattabuero"),
			Entries: []ui_components.SidebarEntry{
				{Name: "Abmeldung", URL: "/internal/regattabuero/vereinswahl?next=abmeldung&title=Vereinswahl%20für%20Abmeldung", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Ummeldung", URL: "/internal/regattabuero/vereinswahl?next=ummeldung&title=Vereinswahl%20für%20Ummeldung", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Nachmeldung", URL: "/internal/regattabuero/vereinswahl?next=nachmeldung&title=Vereinswahl%20für%20Nachmeldung", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Waage", URL: "/internal/regattabuero/vereinswahl?next=waage&title=Vereinswahl%20für%20Waage", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Startberechtigung", URL: "/internal/regattabuero/vereinswahl?next=startberechtigung&title=Vereinswahl%20für%20Startberechtigung", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Sportler anlegen", URL: "/internal/regattabuero/vereinswahl?next=new_athlet&title=Vereinswahl%20für%20neuen%20Sportler", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Kasse", URL: "/internal/regattabuero/kasse", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Startnummern Ausgabe/Rückgabe", URL: "/internal/regattabuero/startnummernausgabe", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Startnummern ändern", URL: "/internal/regattabuero/startnummern/aenderung", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Setzungsänderung", URL: "/internal/regattabuero/setzungsverwaltung/aenderung", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Änderungen von Obleuten", URL: "/internal/regattabuero/aenderungen_obleute", RequiredCaps: []string{"allowed_regattabuero"}},
			},
		},
		{
			Name:         "Regattaleitung",
			URL:          "/internal/regattaleitung",
			RequiredCaps: []string{"allowed_regattaleitung"},
			IsActive:     strings.HasPrefix(currentPath, "/internal/regattaleitung"),
			Entries: []ui_components.SidebarEntry{
				{Name: "DRV Datei Upload", URL: "/internal/regattaleitung/drvupload", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Setzungsauslosung", URL: "/internal/regattaleitung/setzungsverwaltung/losung", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Setzungsänderung", URL: "/internal/regattaleitung/setzungsverwaltung/aenderung", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Pausen", URL: "/internal/regattaleitung/pausen", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Zeitplan", URL: "/internal/regattaleitung/zeitplan", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Startnummern verteilen", URL: "/internal/regattaleitung/startnummern/verteilen", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Startnummernbereiche", URL: "/internal/regattaleitung/startnummern/bereich", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Startnummern ändern", URL: "/internal/regattaleitung/startnummern/aenderung", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "PDF Meldeergebnis", URL: "/internal/regattaleitung/pdf_meldeergebnis", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Vereine verwalten", URL: "/internal/regattaleitung/vereine", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Obleute verwalten", URL: "/internal/regattaleitung/obleute", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "E-Mail senden", URL: "/internal/regattaleitung/email", RequiredCaps: []string{"allowed_regattaleitung"}},
			},
		},
		{
			Name:         "Admin",
			URL:          "/internal/admin",
			RequiredCaps: []string{"allowed_admin"},
			IsActive:     strings.HasPrefix(currentPath, "/internal/admin"),
			Entries: []ui_components.SidebarEntry{
				{Name: "Nutzer Verwaltung", URL: "/internal/admin/users", RequiredCaps: []string{"allowed_admin"}},
				{Name: "Nutzergruppen", URL: "/internal/admin/usergroups", RequiredCaps: []string{"allowed_admin"}},
				{Name: "E-Mail Queue", URL: "/internal/admin/email_queue", RequiredCaps: []string{"allowed_admin"}},
			},
		},
	}
}

func pageTitleFromPath(path string) string {
	switch {
	case path == "/internal" || path == "/internal/":
		return "Dashboard"
	case strings.HasPrefix(path, "/internal/profil"):
		return "Profil"
	case strings.HasPrefix(path, "/internal/zeitnahme"):
		return "Zeitnahme"
	case strings.HasPrefix(path, "/internal/startlisten"):
		return "Startlisten"
	case strings.HasPrefix(path, "/internal/regattabuero"):
		return "Regattabüro"
	case strings.HasPrefix(path, "/internal/regattaleitung"):
		return "Regattaleitung"
	case strings.HasPrefix(path, "/internal/admin"):
		return "Admin"
	default:
		return "Intern"
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
