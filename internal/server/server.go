package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/middleware"
	"github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/internal/templates/pages"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	r            = newRouter()
	navBarConfig = ui_components.NewNavBarConfig()
)

type router struct {
	handlers map[string]map[string]http.HandlerFunc
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]map[string]http.HandlerFunc),
	}
}

func (r *router) Handle(method, pattern string, h func(http.ResponseWriter, *http.Request)) {
	if r.handlers[method] == nil {
		r.handlers[method] = make(map[string]http.HandlerFunc)
	}
	r.handlers[method][pattern] = h
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	methodHandlers := r.handlers[req.Method]
	if methodHandlers == nil {
		http.NotFound(w, req)
		return
	}

	handler := methodHandlers[req.URL.Path]
	var params map[string]string
	if handler == nil {
		handler, params = r.matchWildcard(req.Method, req.URL.Path)
	}
	if handler == nil {
		http.NotFound(w, req)
		return
	}

	if params != nil {
		ctx := context.WithValue(req.Context(), "pathParams", params)
		req = req.WithContext(ctx)
	}

	handler(w, req)
}

func (r *router) matchWildcard(method, path string) (http.HandlerFunc, map[string]string) {
	methodHandlers := r.handlers[method]
	for pattern := range methodHandlers {
		params, ok := matchPath(pattern, path)
		if ok {
			return methodHandlers[pattern], params
		}
	}
	return nil, nil
}

func matchPath(pattern, path string) (map[string]string, bool) {
	patParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patParts) == 0 || len(pathParts) == 0 {
		return nil, false
	}

	params := make(map[string]string)

	for i := range patParts {
		if i >= len(pathParts) {
			return nil, false
		}

		if strings.HasPrefix(patParts[i], "{") && strings.HasSuffix(patParts[i], "}") {
			if i == len(patParts)-1 {
				params[patParts[i][1:len(patParts[i])-1]] = strings.Join(pathParts[i:], "/")
				return params, true
			}
			params[patParts[i][1:len(patParts[i])-1]] = pathParts[i]
		} else if patParts[i] != pathParts[i] {
			return nil, false
		}
	}

	return params, len(patParts) == len(pathParts)
}

func zeitplanCollapseBodyHandler(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return &handler.Error{StatusCode: http.StatusNotFound, Message: "Wettkampf not found"}
	}
	templ.Handler(ui_components.ZeitplanCollapseBody(wettkampf)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func ausschreibungRennenCollapseBodyHandler(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return &handler.Error{StatusCode: http.StatusNotFound, Message: "Wettkampf not found"}
	}
	templ.Handler(ui_pages.AusschreibungRennenCollapseBody(wettkampf)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func meldeergebnisCollapseBodyHandler(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return &handler.Error{StatusCode: http.StatusNotFound, Message: "Wettkampf not found"}
	}
	templ.Handler(ui_pages.MeldeergebnisCollapseBody(wettkampf)).ServeHTTP(c.Writer, c.Request)
	return nil
}

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

	baseLayoutHandler("/metrics", metricsPageHandler)

	r.Handle("GET", "/metricsApi", wrapHandler(func(c *handler.Context) error {
		return api_v1.MetricsApi(c)
	}, true))

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
	baseLayoutHandler("/", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.Index(), nil
	})
	baseLayoutHandler("/live", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.Livestream(), nil
	})
	baseLayoutHandler("/ausschreibung", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.Ausschreibung(), nil
	})
	baseLayoutHandler("/zeitplan", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.Zeitplan(), nil
	})
	baseLayoutHandler("/meldeergebnis", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.Meldeergebnis(), nil
	})
	baseLayoutHandler("/ergebnisse", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.Ergebnisse(), nil
	})
	baseLayoutHandler("/login", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.Login(""), nil
	})
	r.Handle("POST", "/login", wrapHandler(loginPostHandler, false))
	r.Handle("GET", "/logout", logoutHandler)

	baseLayoutHandler("/datenschutz", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.Datenschutz(), nil
	})

	baseLayoutHandler("/internal", func(c *handler.Context) (templ.Component, error) {
		userToken, ok := c.GetLocals("user").(*jwt.Token)
		if !ok {
			return nil, &handler.Error{StatusCode: http.StatusUnauthorized, Message: "Nicht angemeldet"}
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
	})

	baseLayoutHandler("/internal/profil", func(c *handler.Context) (templ.Component, error) {
		return getProfilePage(c)
	})
	r.Handle("GET", "/internal/profil/password/{uuid}", wrapHandler(changePasswordGetHandler, true))
	r.Handle("PUT", "/internal/profil/password/{uuid}", wrapHandler(changePasswordPostHandler, true))
	baseLayoutHandler("/internal/zeitnahme", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalZeitnahme(), nil
	})

	baseLayoutHandler("/internal/startlisten", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalStartlisten(), nil
	})

	baseLayoutHandler("/internal/regattabuero", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattabuero(), nil
	})
	baseLayoutHandler("/internal/regattabuero/vereinswahl", func(c *handler.Context) (templ.Component, error) {
		next := c.GetQueryParam("next")
		if next == "" {
				return ui_pages.Error(404, "Next param is required"), errors.New("Next param is required")
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
			vereine, err = crud.GetAllVerein(c.Request.Context(), )
		}
		if err != nil {
			return ui_pages.Error(500, "Fehler beim Laden der Vereine"), errors.New("Fehler beim Laden der Vereine")
		}

		return ui_pages.InternalVereinswahl(nextUrl, title, vereine), nil
	})
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/abmeldung", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		meldungen, err := crud.GetAllMeldungForVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading meldungen"), errors.New("Error while loading meldungen")
		}
		return ui_pages.InternalRegattabueroAbmeldung(verein, meldungen), nil
	})
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/abmeldung/{m_uuid}", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		meldungUuidStr := c.Param("m_uuid")
		meldungUuid, err := uuid.Parse(meldungUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		meldung, err := crud.GetMeldung(c.Request.Context(), meldungUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading meldung"), errors.New("Error while loading meldung")
		}

		if meldung.VereinUuid != verein.Uuid {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		return ui_pages.InternalRegattabueroAbmeldungMeldung(verein, meldung), nil
	})
	r.Handle("DELETE", "/internal/regattabuero/{v_uuid}/abmeldung/{m_uuid}", wrapHandler(abmeldungDeleteHandler, true))
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/ummeldung", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		meldungen, err := crud.GetAllMeldungForVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading meldungen"), errors.New("Error while loading meldungen")
		}
		return ui_pages.InternalRegattabueroUmmeldung(verein, meldungen), nil
	})
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/ummeldung/{m_uuid}", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		meldungUuidStr := c.Param("m_uuid")
		meldungUuid, err := uuid.Parse(meldungUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		meldung, err := crud.GetMeldung(c.Request.Context(), meldungUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading meldung"), errors.New("Error while loading meldung")
		}

		if meldung.VereinUuid != verein.Uuid {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		athleten, err := crud.GetAllAthletenForVerein(c.Request.Context(), verein.Uuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading athleten"), errors.New("Error while loading athleten")
		}

		// TODO: Filter only viable athleten
		return ui_pages.InternalRegattabueroUmmeldungMeldung(verein, meldung, athleten), nil
	})
	r.Handle("POST", "/internal/regattabuero/{v_uuid}/ummeldung/{m_uuid}", wrapHandler(ummeldungPostHandler, true))
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/nachmeldung", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		return ui_pages.InternalRegattabueroNachmeldung(verein), nil
	})
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/nachmeldung/{r_uuid}", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		rennenUuidStr := c.Param("r_uuid")
		rennenUuid, err := uuid.Parse(rennenUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		rennen, err := crud.GetRennen(c.Request.Context(), rennenUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading rennen"), errors.New("Error while loading rennen")
		}

		athleten, err := crud.GetAllAthletenForVerein(c.Request.Context(), verein.Uuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading athleten"), errors.New("Error while loading athleten")
		}

		// TODO: Filter only viable athleten
		return ui_pages.InternalRegattabueroNachmeldungMeldung(verein, rennen, athleten), nil
	})
	r.Handle("POST", "/internal/regattabuero/{v_uuid}/nachmeldung/{r_uuid}", wrapHandler(nachmeldungPostHandler, true))
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/nachmeldung/success/{m_uuid}", func(c *handler.Context) (templ.Component, error) {
		meldungUuidStr := c.Param("m_uuid")
		meldungUuid, err := uuid.Parse(meldungUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		m, err := crud.GetMeldung(c.Request.Context(), meldungUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading meldung") , errors.New("Error while loading meldung")
		}
		return ui_pages.InternalRegattabueroNachmeldungSuccess(m), nil
	})
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/waage", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		athleten, err := crud.GetAllAthletenForVereinWaage(c.Request.Context(), verein.Uuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading athleten"), errors.New("Error while loading athleten")
		}

		for i := range athleten {
			athleten[i].Verein = &verein
		}

		return ui_pages.InternalRegattabueroWaageWahl(verein, athleten), nil
	})
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/waage/{a_uuid}", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		athletUuidStr := c.Param("a_uuid")
		athletUuid, err := uuid.Parse(athletUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		athlet, err := crud.GetAthlet(c.Request.Context(), athletUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading athlet"), errors.New("Error while loading athlet")
		}

		if athlet.VereinUuid != verein.Uuid {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		return ui_pages.InternalRegattabueroWaage(athlet), nil
	})
	r.Handle("POST", "/internal/regattabuero/{v_uuid}/waage/{a_uuid}", wrapHandler(waagePostHandler, true))
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/startberechtigung", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		athleten, err := crud.GetAllAthletenForVereinMissStartber(c.Request.Context(), verein.Uuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading athleten"), errors.New("Error while loading athleten")
		}

		for i := range athleten {
			athleten[i].Verein = &verein
		}

		return ui_pages.InternalRegattabueroWaageWahl(verein, athleten), nil
	})
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/startberechtigung/{a_uuid}", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		athletUuidStr := c.Param("a_uuid")
		athletUuid, err := uuid.Parse(athletUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		athlet, err := crud.GetAthletMinimal(c.Request.Context(), athletUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading athlet"), errors.New("Error while loading athlet")
		}

		if athlet.VereinUuid != verein.Uuid {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}

		return ui_pages.InternalRegattabueroStartberechtigung(athlet), nil
	})
	baseLayoutHandler("/internal/regattabuero/{v_uuid}/new_athlet", func(c *handler.Context) (templ.Component, error) {
		vereinUuidStr := c.Param("v_uuid")
		vereinUuid, err := uuid.Parse(vereinUuidStr)
		if err != nil {
			return ui_pages.Error(406, "Invalid UUID"), errors.New("Invalid UUID")
		}
		verein, err := crud.GetVerein(c.Request.Context(), vereinUuid)
		if err != nil {
			return ui_pages.Error(500, "Error while loading verein"), errors.New("Error while loading verein")
		}
		return ui_pages.InternalRegattabueroNewAthlet(verein), nil
	})

	baseLayoutHandler("/internal/regattaleitung", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitung(), nil
	})
	baseLayoutHandler("/internal/regattaleitung/drvupload", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungDrvFileUpload(""), nil
	})
	r.Handle("POST", "/internal/regattaleitung/drvupload", wrapHandler(drvUploadPostHandler, true))
	baseLayoutHandler("/internal/regattaleitung/setzungsverwaltung", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungSetzung(), nil
	})
	baseLayoutHandler("/internal/regattaleitung/setzungsverwaltung/losung", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungSetzungLosung(), nil
	})
	r.Handle("POST", "/internal/regattaleitung/setzungsverwaltung/losung", wrapHandler(setzungsVerwaltungLosungPostHandler, true))
	r.Handle("DELETE", "/internal/regattaleitung/setzungsverwaltung/losung", wrapHandler(setzungsVerwaltungLosungDeleteHandler, true))
	baseLayoutHandler("/internal/regattaleitung/setzungsverwaltung/aenderung", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungSetzungAenderung(), nil
	})
	r.Handle("POST", "/internal/regattaleitung/setzungsverwaltung/aenderung/rennen/{param}", wrapHandler(setzungsVerwaltungAenderungRennenPostHandler, true))
	baseLayoutHandler("/internal/regattaleitung/setzungsverwaltung/aenderung/rennen/{param}", func(c *handler.Context) (templ.Component, error) {
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
	})
	baseLayoutHandler("/internal/regattaleitung/setzungsverwaltung/aenderung", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungSetzungAenderung(), nil
	})

	baseLayoutHandler("/internal/regattaleitung/pausen", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungPausen(), nil
	})
	r.Handle("GET", "/internal/regattaleitung/pausen/new/{nach_rennen_uuid}", wrapHandler(pausenNewHandler, true))
	r.Handle("POST", "/internal/regattaleitung/pausen", wrapHandler(pausenPostHandler, true))
	r.Handle("DELETE", "/internal/regattaleitung/pausen/{id}", wrapHandler(pausenDeleteHandler, true))

	baseLayoutHandler("/internal/regattaleitung/zeitplan", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungZeitplan(), nil
	})
	r.Handle("POST", "/internal/regattaleitung/zeitplan", wrapHandler(zeitplanPostHandler, true))
	baseLayoutHandler("/internal/regattaleitung/startnummern", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungStartnummern(), nil
	})
	baseLayoutHandler("/internal/regattaleitung/startnummern/verteilen", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungStartnummernVerteilen(), nil
	})
	r.Handle("POST", "/internal/regattaleitung/startnummern/verteilen", wrapHandler(startnummernVerteilenPostHandler, true))
	r.Handle("DELETE", "/internal/regattaleitung/startnummern/verteilen", wrapHandler(startnummernVerteilenDeleteHandler, true))
	baseLayoutHandler("/internal/regattaleitung/startnummern/bereich", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungStartnummernBereich(), nil
	})
	baseLayoutHandler("/internal/regattaleitung/startnummern/aenderung", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungStartnummernAendern(), nil
	})
	baseLayoutHandler("/internal/regattaleitung/pdf_meldeergebnis", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungPdfMeldeergebnis(false), nil
	})
	r.Handle("POST", "/internal/regattaleitung/pdf_meldeergebnis", wrapHandler(pdfMeldeergebnisPostHandler, true))
	baseLayoutHandler("/internal/regattaleitung/vereine", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungVereinsverwaltung(), nil
	})
	baseLayoutHandler("/internal/regattaleitung/email", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitungEmailIndex(), nil
	})

	baseLayoutHandler("/internal/admin", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalAdmin(), nil
	})

	baseLayoutHandler("/internal/admin/users", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalAdminUsers(), nil
	})
	r.Handle("GET", "/internal/admin/user/{uuid}", wrapHandler(userEditNewHandler, true))
	r.Handle("POST", "/internal/admin/user/{uuid}", wrapHandler(userEditNewHandlerPost, true))
	baseLayoutHandler("/internal/admin/usergroups", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalAdminUserGroups(), nil
	})
	r.Handle("GET", "/internal/admin/usergroups/{uuid}", wrapHandler(userGroupEditNewHandler, true))
	r.Handle("POST", "/internal/admin/usergroups/{uuid}", wrapHandler(userGroupEditNewHandlerPost, true))

	// Pure HTMX UI Components
	r.Handle("GET", "/comp/image", wrapUIHandler(imageComponentHandler))
	r.Handle("GET", "/comp/zeitplan/{wettkampf}", templHandler(zeitplanCollapseBodyHandler))
	r.Handle("GET", "/comp/ausschreibung/{wettkampf}", templHandler(ausschreibungRennenCollapseBodyHandler))
	r.Handle("GET", "/comp/meldeergebnis/{wettkampf}", templHandler(meldeergebnisCollapseBodyHandler))

	r.Handle("GET", "/comp/internal/rennen_tab/{wettkampf}", templHandler(rennenTabHandler))

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

	r.Handle("GET", "/api/v1/zeitnahme/ziel", wrapHandler(api_v1.WsZeitnahmeZiel, true))
	r.Handle("POST", "/api/v1/zeitnahme/start", wrapHandler(api_v1.PostZeitnahmeStart, true))
	r.Handle("GET", "/api/v1/zeitnahme/openStarts", wrapHandler(api_v1.GetOpenStarts, true))
	r.Handle("POST", "/api/v1/zeitnahme/genErgebnis", wrapHandler(api_v1.GenerateEndZeit, true))

	go handlers.RunHub()

	return r
}

func wrapHandler(h handler.Handler, needAuth bool) func(http.ResponseWriter, *http.Request) {
	defaultStack := []middleware.Middleware{
		middleware.Recovery(),
		middleware.Compression(),
		middleware.Logging(),
		middleware.CORS(),
		middleware.RateLimit(),
		middleware.Timeout(30*time.Second, "Request timeout"),
	}

	stack := defaultStack
	if needAuth {
		stack = append(stack, middleware.Auth())
	}

	wrapped := middleware.Chain(h, stack...)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := handler.NewContext(w, r)
		if p := r.Context().Value("pathParams"); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			headersWritten := false
			if ht, ok := ctx.Writer.(handler.HeaderTracker); ok {
				headersWritten = ht.HeadersWritten()
			}
			if he, ok := err.(*handler.Error); ok {
				if !headersWritten {
					ctx.Writer.WriteHeader(he.StatusCode)
				}
				ctx.Writer.Write([]byte(he.Message))
			} else {
				if !headersWritten {
					ctx.Writer.WriteHeader(http.StatusInternalServerError)
				}
				ctx.Writer.Write([]byte(err.Error()))
			}
		}
	}
}

func templHandler(h handler.Handler) http.HandlerFunc {
	uiStack := []middleware.Middleware{
		middleware.Recovery(),
		middleware.Compression(),
		middleware.Logging(),
		middleware.CORS(),
		middleware.RateLimit(),
		middleware.Timeout(60*time.Second, "Request timeout"),
	}

	wrapped := middleware.Chain(h, uiStack...)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := handler.NewContext(w, r)

		if !ctx.IsHtmxRequest() {
			templ.Handler(ui_pages.Error(404, "")).ServeHTTP(w, r)
			return
		}

		if p := r.Context().Value("pathParams"); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}

func toastReturn(c *handler.Context, msg string, color ui_components.InputColor) error {
	c.Writer.Header().Set("HX-Retarget", "#toast-container")
	c.Writer.Header().Set("HX-Swap", "beforeend")
	templ.Handler(ui_components.Toast(msg, color)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func wrapUIHandler(h handler.Handler) func(http.ResponseWriter, *http.Request) {
	uiStack := []middleware.Middleware{
		middleware.Recovery(),
		middleware.Compression(),
		middleware.Logging(),
		middleware.CORS(),
		middleware.RateLimit(),
		middleware.Timeout(60*time.Second, "Request timeout"),
	}

	wrapped := middleware.Chain(h, uiStack...)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := handler.NewContext(w, r)
		if p := r.Context().Value("pathParams"); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			if he, ok := err.(*handler.Error); ok {
				w.WriteHeader(he.StatusCode)
				w.Write([]byte(he.Message))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}
	}
}
