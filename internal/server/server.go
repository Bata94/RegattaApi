package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/middleware"
	"github.com/bata94/RegattaApi/internal/sqlc"
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

	for i := 0; i < len(patParts); i++ {
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
		return &handler.Error{StatusCode: 404, Message: "Wettkampf not found"}
	}
	templ.Handler(ui_components.ZeitplanCollapseBody(wettkampf)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func ausschreibungRennenCollapseBodyHandler(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return &handler.Error{StatusCode: 404, Message: "Wettkampf not found"}
	}
	templ.Handler(ui_pages.AusschreibungRennenCollapseBody(wettkampf)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func meldeergebnisCollapseBodyHandler(c *handler.Context) error {
	wettkampfStr := c.Param("wettkampf")
	wettkampf, err := crud.WettkampfFromString(wettkampfStr)
	if err != nil {
		return &handler.Error{StatusCode: 404, Message: "Wettkampf not found"}
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
		http.ServeFile(w, r, "./files/"+r.URL.Path[len("/files/"):])
	})
	r.Handle("GET", "/public/{file}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/"+r.URL.Path[len("/public/"):])
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

	baseLayoutHandler("/internal/profil", func(c *handler.Context) (templ.Component, error) {
		return getProfilePage(c)
	})
	baseLayoutHandler("/internal/zeitnahme", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalZeitnahme(), nil
	})
	baseLayoutHandler("/internal/startlisten", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalStartlisten(), nil
	})
	baseLayoutHandler("/internal/regattabuero", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattabuero(), nil
	})
	baseLayoutHandler("/internal/regattaleitung", func(c *handler.Context) (templ.Component, error) {
		return ui_pages.InternalRegattaleitung(), nil
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

	// Pure HTMX UI Components
	r.Handle("GET", "/comp/image", wrapUIHandler(imageComponentHandler))
	r.Handle("GET", "/comp/zeitplan/{wettkampf}", templHandler(zeitplanCollapseBodyHandler))
	r.Handle("GET", "/comp/ausschreibung/{wettkampf}", templHandler(ausschreibungRennenCollapseBodyHandler))
	r.Handle("GET", "/comp/meldeergebnis/{wettkampf}", templHandler(meldeergebnisCollapseBodyHandler))

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
	r.Handle("GET", "/api/v1/leitung/meldeergebnis", wrapHandler(api_v1.GetMeldeergebnisHtml, true))
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
		middleware.Timeout(30 * time.Second, "Request timeout"),
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

		if ctx.Headers().Get("HX-Request") != "true" {
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

func getProfilePage(c *handler.Context) (templ.Component, error) {
	userToken, ok := c.GetLocals("user").(*jwt.Token)
	if !ok {
		return nil, &handler.Error{StatusCode: 401, Message: "Nicht angemeldet"}
	}

	claims := userToken.Claims.(jwt.MapClaims)
	username, ok := claims["username"].(string)
	if !ok {
		return nil, &handler.Error{StatusCode: 401, Message: "Invalid token"}
	}

	userGroup := ""
	if ug, ok := claims["user_group_name"].(string); ok {
		userGroup = ug
	}

	var capabilities []string
	capFields := []string{"allowed_admin", "allowed_zeitnahme", "allowed_startlisten", "allowed_regattabuero", "allowed_regattaleitung"}
	for _, field := range capFields {
		if val, exists := claims[field]; exists && val == true {
			capabilities = append(capabilities, field)
		}
	}

	data := ui_pages.ProfilData{
		Username:     username,
		UserGroup:    userGroup,
		Capabilities: capabilities,
	}

	return ui_pages.Profil(data), nil
}

func metricsPageHandler(c *handler.Context) (templ.Component, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "DO_NOT_USE_IN_PROD"
	}

	tokenString := c.Cookie("auth_token")
	if tokenString == "" {
		return nil, &handler.Error{StatusCode: 401, Message: "Nicht angemeldet"}
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, &handler.Error{StatusCode: 401, Message: "Ungültiges oder abgelaufenes Token"}
	}

	claims := token.Claims.(jwt.MapClaims)
	admin, ok := claims["allowed_admin"].(bool)
	if !ok || !admin {
		return nil, &handler.Error{StatusCode: 403, Message: "Keine Admin-Berechtigung"}
	}

	return ui_pages.Metrics(), nil
}

func loginPostHandler(c *handler.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	isHTMX := c.Headers().Get("HX-Request") == "true"

	if username == "" || password == "" {
		if isHTMX {
			templ.Handler(ui_pages.Login("Benutzername und Passwort erforderlich")).ServeHTTP(c.Writer, c.Request)
			return nil
		}
		return c.Redirect("/login", http.StatusSeeOther)
	}

	u, err := crud.AuthLogin(crud.LoginParams{Username: username, Password: password})
	if err != nil {
		if isHTMX {
			templ.Handler(ui_pages.Login("Benutzername oder Passwort ist falsch")).ServeHTTP(c.Writer, c.Request)
			return nil
		}
		return c.Redirect("/login", http.StatusSeeOther)
	}

	secure := c.Request.TLS != nil
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    u.Jwt.Token,
		MaxAge:   72 * 60 * 60,
		HttpOnly: true,
		Secure:   secure,
		Path:     "/",
	})

	if isHTMX {
		c.Writer.Header().Set("HX-Redirect", "/")
		c.Writer.WriteHeader(http.StatusOK)
		return nil
	}

	return c.Redirect("/", http.StatusSeeOther)
}

func imageComponentHandler(c *handler.Context) error {
	queryParams := c.Request.URL.Query()
	src := queryParams.Get("src")
	alt := queryParams.Get("alt")

	if src == "" {
		return &handler.Error{StatusCode: 404, Message: "Image src is empty"}
	}

	imgOpt := ui_components.DefaultImageOptions()

	if w := queryParams.Get("width"); w != "" {
		imgOpt.Width = w
	}
	if h := queryParams.Get("height"); h != "" {
		imgOpt.Height = h
	}
	if q := queryParams.Get("quality"); q != "" {
		qFloat64, err := strconv.ParseFloat(q, 32)
		if err != nil || qFloat64 <= 0.0 || qFloat64 > 100.0 {
			log.Printf("Image component: quality is not a float32 value or out of range... Setting default value")
		} else {
			imgOpt.Quality = float32(qFloat64)
		}
	}
	if l := queryParams.Get("lossless"); l != "" {
		imgOpt.Lossless, _ = strconv.ParseBool(l)
		if imgOpt.Lossless {
			log.Printf("Image component: lossless is not a bool value... Setting default value")
			imgOpt.Lossless = false
		}
	}
	if class := queryParams.Get("class"); class != "" {
		imgOpt.ClassImage = class
	}

	templ.Handler(ui_components.RawImageComponent(src, alt, imgOpt)).ServeHTTP(c.Writer, c.Request)
	return nil
}

func userEditNewHandler(c *handler.Context) error {
	var u *crud.User
	if c.Param("uuid") == "" {
		return &handler.Error{StatusCode: 404, Message: "User not found"}
	} else if c.Param("uuid") == "new" {
		u = &crud.User{
			User:      sqlc.User{},
			UserGroup: &sqlc.UsersGroup{},
		}
	} else {
		uuid, err := uuid.Parse(c.Param("uuid"))
		if err != nil {
			return &handler.Error{StatusCode: 406, Message: "Invalid UUID"}
		}
		u, err = crud.GetUser(uuid)
		if err != nil {
			return &handler.Error{StatusCode: 404, Message: "User not found"}
		}
	}

	templ.Handler(ui_components.UserEdit(*u, "")).ServeHTTP(c.Writer, c.Request)
	return nil
}

func userEditNewHandlerPost(c *handler.Context) error {
	var (
		u *crud.User
		userUuid uuid.UUID
		err error
	)

	uuidStr := c.Param("uuid")
	if uuidStr == "new" {
		userUuid, err = uuid.NewV7()
	} else {
		userUuid, err = uuid.Parse(uuidStr)
	}

	groupUuidStr := c.FormValue("user_group_uuid")
	groupUuid, errGroupUuid := uuid.Parse(groupUuidStr)

	if err != nil || errGroupUuid != nil {
		return &handler.Error{StatusCode: 406, Message: "Invalid UUID"}
	}

	u = &crud.User{
		User:      sqlc.User{
			Uuid: userUuid,
			Username: c.FormValue("username"),
		},
		UserGroup: &sqlc.UsersGroup{
			Uuid:  groupUuid,
		},
	}

	if uuidStr == "new" {
		_, err = crud.CreateUser(crud.CreateUserParams{
			GroupUuid: groupUuid,
			Username:  c.FormValue("username"),
			Password:  c.FormValue("password"),
		})
		if err != nil {
			// return &handler.Error{StatusCode: 500, Message: "Error while creating user"}
			templ.Handler(ui_components.UserEdit(*u, "Error while creating user, Err: " + err.Error())).ServeHTTP(c.Writer, c.Request)
			return nil
		}

		c.Writer.Header().Set("HX-Redirect", "/internal/admin/users")
		c.Writer.WriteHeader(http.StatusOK)
		return nil
	}

	templ.Handler(ui_components.UserEdit(*u, "Editing not yet implemented")).ServeHTTP(c.Writer, c.Request)
	return nil
}
