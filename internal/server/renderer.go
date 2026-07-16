package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/middleware"
	"github.com/golang-jwt/jwt/v5"

	"github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/internal/templates/layout"
)

type PageFunc func(c *handler.Context) (templ.Component, error)

func baseLayoutHandler(url string, getPage PageFunc) {
	r.Handle("GET", url, func(w http.ResponseWriter, r *http.Request) {
		ctx := handler.NewContext(w, r)

		uiStack := []middleware.Middleware{
			middleware.Recovery(),
			middleware.Compression(),
			middleware.Logging(),
			middleware.CORS(),
			middleware.RateLimit(),
			middleware.OptionalAuth(),
			middleware.Timeout(60*time.Second, "Request timeout"),
		}

		h := func(c *handler.Context) error {
			pageBody, err := getPage(c)
			if err != nil {
				return err
			}

			var caps []string
			if c.GetLocals("capabilities") != nil {
				caps = c.GetLocals("capabilities").([]string)
			}
			loggedIn := false
			if c.GetLocals("logged_in") != nil {
				loggedIn = c.GetLocals("logged_in").(bool)
			}

			navbarCfg := ui_components.NavBarConfig{
				Entries:  navBarConfig.Entries,
				UserCaps: caps,
				LoggedIn: loggedIn,
			}

			if c.IsHtmxRequest() {
				templ.Handler(pageBody).ServeHTTP(c.Writer, c.Request)
			} else {
				templ.Handler(ui_layouts.BaseLayout(pageBody, navbarCfg)).ServeHTTP(c.Writer, c.Request)
			}
			return nil
		}

		wrapped := middleware.Chain(h, uiStack...)

		if p := r.Context().Value(handler.CtxKeyPathParams); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			handleAppError(ctx, err)
		}
	})
}

func internalLayoutHandler(url string, getPage PageFunc) {
	r.Handle("GET", url, func(w http.ResponseWriter, r *http.Request) {
		ctx := handler.NewContext(w, r)

		uiStack := []middleware.Middleware{
			middleware.Recovery(),
			middleware.Compression(),
			middleware.Logging(),
			middleware.CORS(),
			middleware.RateLimit(),
			middleware.Auth(),
			middleware.Timeout(60*time.Second, "Request timeout"),
		}

		h := func(c *handler.Context) error {
			pageBody, err := getPage(c)
			if err != nil {
				return err
			}

			var caps []string
			if c.GetLocals("capabilities") != nil {
				caps = c.GetLocals("capabilities").([]string)
			}

			username := ""
			if c.GetLocals("user") != nil {
				token := c.GetLocals("user").(*jwt.Token)
				claims := token.Claims.(jwt.MapClaims)
				if u, ok := claims["username"].(string); ok {
					username = u
				}
			}

			path := r.URL.Path
			cats := buildSidebarCategories(path, caps)
			title := pageTitleFromPath(path)

			if c.IsHtmxRequest() {
				templ.Handler(pageBody).ServeHTTP(c.Writer, c.Request)
			} else {
				templ.Handler(ui_layouts.InternalLayout(pageBody, cats, caps, username, title)).ServeHTTP(c.Writer, c.Request)
			}
			return nil
		}

		wrapped := middleware.Chain(h, uiStack...)

		if p := r.Context().Value(handler.CtxKeyPathParams); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			handleAppError(ctx, err)
		}
	})
}

func buildSidebarCategories(currentPath string, caps []string) []ui_components.SidebarCategory {
	return []ui_components.SidebarCategory{
		{
			Name:         "Zeitnahme",
			URL:          "/internal/zeitnahme",
			RequiredCaps: []string{"allowed_zeitnahme"},
			IsActive:     strings.HasPrefix(currentPath, "/internal/zeitnahme"),
			Entries: []ui_components.SidebarEntry{
				{Name: "Zielgericht", URL: "/internal/zeitnahme/ziel", RequiredCaps: []string{"allowed_zeitnahme"}},
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
				{Name: "Kasse", URL: "#", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Startnummern Ausgabe/Rückgabe", URL: "#", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Startnummern ändern", URL: "/internal/regattabuero/startnummern/aenderung", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Setzungsänderung", URL: "/internal/regattabuero/setzungsverwaltung/aenderung", RequiredCaps: []string{"allowed_regattabuero"}},
				{Name: "Änderungen von Obleuten", URL: "#", RequiredCaps: []string{"allowed_regattabuero"}},
			},
		},
		{
			Name:         "Regattaleitung",
			URL:          "/internal/regattaleitung",
			RequiredCaps: []string{"allowed_regattaleitung"},
			IsActive:     strings.HasPrefix(currentPath, "/internal/regattaleitung"),
			Entries: []ui_components.SidebarEntry{
				{Name: "DRV Datei Upload", URL: "/internal/regattaleitung/drvupload", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Setzungsverwaltung", URL: "/internal/regattaleitung/setzungsverwaltung", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Setzungsauslosung", URL: "/internal/regattaleitung/setzungsverwaltung/losung", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Setzungsänderung", URL: "/internal/regattaleitung/setzungsverwaltung/aenderung", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Pausen", URL: "/internal/regattaleitung/pausen", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Zeitplan", URL: "/internal/regattaleitung/zeitplan", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Startnummern verteilen", URL: "/internal/regattaleitung/startnummern/verteilen", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Startnummernbereiche", URL: "/internal/regattaleitung/startnummern/bereich", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Startnummern ändern", URL: "/internal/regattaleitung/startnummern/aenderung", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "PDF Meldeergebnis", URL: "/internal/regattaleitung/pdf_meldeergebnis", RequiredCaps: []string{"allowed_regattaleitung"}},
				{Name: "Vereine verwalten", URL: "/internal/regattaleitung/vereine", RequiredCaps: []string{"allowed_regattaleitung"}},
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
