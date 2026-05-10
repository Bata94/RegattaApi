package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/handlers/api/v1"
	"github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/internal/templates/pages"
)

var (
	corsAllowedOrigins    = os.Getenv("CORS_ALLOWED_ORIGINS")
	corsAllowedMethods     = os.Getenv("CORS_ALLOWED_METHODS")
	corsAllowedHeaders    = os.Getenv("CORS_ALLOWED_HEADERS")

	r = newRouter()
	navBarConfig = ui_components.NewNavBarConfig()
)

func init() {
	if corsAllowedOrigins == "" {
		corsAllowedOrigins = "*"
	}
	if corsAllowedMethods == "" {
		corsAllowedMethods = "GET, POST, PUT, DELETE, OPTIONS"
	}
	if corsAllowedHeaders == "" {
		corsAllowedHeaders = "Content-Type, Authorization"
	}
}

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

type corsRouter struct {
	h http.Handler
}

func (c *corsRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", corsAllowedOrigins)
	w.Header().Set("Access-Control-Allow-Methods", corsAllowedMethods)
	w.Header().Set("Access-Control-Allow-Headers", corsAllowedHeaders)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	c.h.ServeHTTP(w, r)
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

func GetRouter() http.Handler {
	navBarConfig.Entries = []ui_components.NavBarEntry{
		// {Name: "Home", URL: "/"},
		{Name: "Livestream", URL: "/live"},
		{Name: "Ausschreibung", URL: "/ausschreibung"},
		{Name: "Zeitplan", URL: "/zeitplan"},
		{Name: "Meldeergebnis", URL: "/meldeergebnis"},
		{Name: "Ergebnisse", URL: "/ergebnisse"},
	}

	r.Handle("GET", "/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Metrics placeholder"))
	})
	r.Handle("GET", "/metricsApi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API Metrics placeholder"))
	})

	r.Handle("GET", "/assets/{file}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/"+r.URL.Path[len("/assets/"):])
	})
	r.Handle("GET", "/files/{file}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./files/"+r.URL.Path[len("/files/"):])
	})

	// UI Handlers
	baseLayoutHandler("/", ui_pages.Index())
	baseLayoutHandler("/live", ui_pages.Livestream())
	baseLayoutHandler("/ausschreibung", ui_pages.Ausschreibung())
	baseLayoutHandler("/zeitplan", ui_pages.Zeitplan())
	baseLayoutHandler("/meldeergebnis", ui_pages.Meldeergebnis())
	baseLayoutHandler("/ergebnisse", ui_pages.Ergebnisse())
	baseLayoutHandler("/login", ui_pages.Login())

	// Pure HTMX UI Components
	r.Handle("GET", "/comp/image", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Image component endpoint hit!")
		queryParams := r.URL.Query()
		src := queryParams.Get("src")
		alt := queryParams.Get("alt")

		if src == "" {
			log.Printf("Image component: src is empty")
			http.NotFound(w, r)
			return
		}

		imgOpt := ui_components.DefaultImageOptions()

		if w := queryParams.Get("width"); w != "" {
			imgOpt.Width = w
		}
		if h := queryParams.Get("height"); h != "" {
			imgOpt.Height = h
		}
		if c := queryParams.Get("class"); c != "" {
			imgOpt.ClassImage = c
		}

		templ.Handler(ui_components.RawImageComponent(src, alt, imgOpt)).ServeHTTP(w, r)
	})
	r.Handle("GET", "/comp/zeitplan/{wettkampf}", templHandler(zeitplanCollapseBodyHandler))
	r.Handle("GET", "/comp/ausschreibung/{wettkampf}", templHandler(ausschreibungRennenCollapseBodyHandler))

	// API Handlers
	r.Handle("POST", "/api/auth/login", wrapHandler(api_v1.Login))
	r.Handle("POST", "/api/auth/logout", wrapHandler(api_v1.Logout))
	r.Handle("GET", "/api/auth/valid", wrapHandler(api_v1.AuthValidate))
	r.Handle("GET", "/api/auth/me", wrapHandler(api_v1.AuthMe))

	r.Handle("GET", "/api/v1/test", wrapHandler(api_v1.TestHandler))

	r.Handle("GET", "/api/v1/athlet/startberechtigung", wrapHandler(api_v1.GetAthletStartberechtigung))
	r.Handle("PUT", "/api/v1/athlet/startberechtigung", wrapHandler(api_v1.UpdateAthletStartberechtigung))
	r.Handle("GET", "/api/v1/athlet/waage", wrapHandler(api_v1.GetAthletWaage))
	r.Handle("PUT", "/api/v1/athlet/waage", wrapHandler(api_v1.UpdateAthletWaage))
	r.Handle("GET", "/api/v1/athlet", wrapHandler(api_v1.GetAllAthlet))
	r.Handle("GET", "/api/v1/athlet/{uuid}", wrapHandler(api_v1.GetAthlet))
	r.Handle("POST", "/api/v1/athlet", wrapHandler(api_v1.CreateAthlet))

	r.Handle("POST", "/api/v1/buero/abmeldung", wrapHandler(api_v1.PostAbmeldung))
	r.Handle("POST", "/api/v1/buero/ummeldung", wrapHandler(api_v1.PostUmmeldung))
	r.Handle("POST", "/api/v1/buero/nachmeldung", wrapHandler(api_v1.PostNachmeldung))
	r.Handle("POST", "/api/v1/buero/startnummernausgabe", wrapHandler(api_v1.StartnummernAusgabe))
	r.Handle("POST", "/api/v1/buero/startnummernwechsel", wrapHandler(api_v1.StartnummernWechsel))
	r.Handle("POST", "/api/v1/buero/kasse/einzahlung", wrapHandler(api_v1.KasseEinzahlung))
	r.Handle("POST", "/api/v1/buero/kasse/rechnung/all", wrapHandler(api_v1.KasseCreateRechnungAllVereine))
	r.Handle("GET", "/api/v1/buero/kasse/rechnung/{uuid}", wrapHandler(api_v1.KasseCreateRechnungHTML))
	r.Handle("POST", "/api/v1/buero/kasse/rechnung/{uuid}", wrapHandler(api_v1.KasseCreateRechnungPDF))

	r.Handle("GET", "/api/v1/leitung/pdfFooter", wrapHandler(api_v1.GetPdfFooter))
	r.Handle("GET", "/api/v1/leitung/meldeergebnis", wrapHandler(api_v1.GetMeldeergebnisHtml))
	r.Handle("GET", "/api/v1/leitung/meldeergebnis/list", wrapHandler(api_v1.GetMeldeergebnisList))
	r.Handle("GET", "/api/v1/leitung/meldeergebnis/{filename}", wrapHandler(api_v1.GetMeldeergebnisFilename))
	r.Handle("POST", "/api/v1/leitung/meldeergebnis", wrapHandler(api_v1.GenerateMeldeergebnis))
	r.Handle("GET", "/api/v1/leitung/ergebnis", wrapHandler(api_v1.GenerateErgebnisHtml))
	r.Handle("POST", "/api/v1/leitung/ergebnis", wrapHandler(api_v1.GenerateErgebnis))
	r.Handle("POST", "/api/v1/leitung/drv_meldung_upload", wrapHandler(api_v1.DrvMeldungUpload))
	r.Handle("POST", "/api/v1/leitung/SetzungsLosung", wrapHandler(api_v1.SetzungsLosung))
	r.Handle("POST", "/api/v1/leitung/SetzungsLosung/reset", wrapHandler(api_v1.ResetSetzung))
	r.Handle("POST", "/api/v1/leitung/SetZeitplan", wrapHandler(api_v1.SetZeitplan))
	r.Handle("POST", "/api/v1/leitung/SetStartnummern", wrapHandler(api_v1.SetStartnummern))

	r.Handle("GET", "/api/v1/meldung", wrapHandler(api_v1.GetAllMeldung))
	r.Handle("GET", "/api/v1/meldung/{uuid}", wrapHandler(api_v1.GetMeldung))
	r.Handle("POST", "/api/v1/meldung/updateSetzungBatch", wrapHandler(api_v1.UpdateSetzungBatch))
	r.Handle("POST", "/api/v1/meldung/abmeldung", wrapHandler(api_v1.PostAbmeldung))
	r.Handle("POST", "/api/v1/meldung/ummeldung", wrapHandler(api_v1.PostUmmeldung))
	r.Handle("POST", "/api/v1/meldung/nachmeldung", wrapHandler(api_v1.PostNachmeldung))
	r.Handle("GET", "/api/v1/meldung/verein/{uuid}", wrapHandler(api_v1.GetAllMeldungForVerein))

	r.Handle("GET", "/api/v1/pause", wrapHandler(api_v1.GetAllPausen))
	r.Handle("GET", "/api/v1/pause/{id}", wrapHandler(api_v1.GetPause))
	r.Handle("DELETE", "/api/v1/pause/{id}", wrapHandler(api_v1.DeletePause))
	r.Handle("POST", "/api/v1/pause", wrapHandler(api_v1.CreatePause))
	r.Handle("PUT", "/api/v1/pause", wrapHandler(api_v1.UpdatePause))

	r.Handle("GET", "/api/v1/rennen", wrapHandler(api_v1.GetAllRennen))
	r.Handle("GET", "/api/v1/rennen/{uuid}", wrapHandler(api_v1.GetRennen))

	r.Handle("GET", "/api/v1/users", wrapHandler(api_v1.GetAllUsers))
	r.Handle("GET", "/api/v1/users/{uuid}", wrapHandler(api_v1.GetUser))
	r.Handle("GET", "/api/v1/users/name/{name}", wrapHandler(api_v1.GetUserByName))
	r.Handle("POST", "/api/v1/users", wrapHandler(api_v1.CreateUser))
	r.Handle("GET", "/api/v1/users/group", wrapHandler(api_v1.GetAllUsersGroups))
	r.Handle("GET", "/api/v1/users/group/{uuid}", wrapHandler(api_v1.GetUsersGroup))
	r.Handle("GET", "/api/v1/users/group/name/{name}", wrapHandler(api_v1.GetUsersGroupByName))

	r.Handle("GET", "/api/v1/verein", wrapHandler(api_v1.GetAllVerein))
	r.Handle("GET", "/api/v1/verein/{uuid}", wrapHandler(api_v1.GetVerein))
	r.Handle("GET", "/api/v1/verein/{uuid}/athlet", wrapHandler(api_v1.GetAllAthletenForVerein))
	r.Handle("GET", "/api/v1/verein/{uuid}/waage", wrapHandler(api_v1.GetAllAthletenForVereinWaage))
	r.Handle("GET", "/api/v1/verein/{uuid}/startberechtigung", wrapHandler(api_v1.GetAllAthletenForVereinMissStartber))

	r.Handle("GET", "/api/v1/zeitnahme/ziel", wrapHandler(api_v1.WsZeitnahmeZiel))
	r.Handle("POST", "/api/v1/zeitnahme/start", wrapHandler(api_v1.PostZeitnahmeStart))
	r.Handle("GET", "/api/v1/zeitnahme/openStarts", wrapHandler(api_v1.GetOpenStarts))
	r.Handle("POST", "/api/v1/zeitnahme/genErgebnis", wrapHandler(api_v1.GenerateEndZeit))

	go handlers.RunHub()

	return &corsRouter{h: r}
}

func wrapHandler(h handler.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s | %s | %s", r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set("Access-Control-Allow-Origin", corsAllowedOrigins)
		w.Header().Set("Access-Control-Allow-Methods", corsAllowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", corsAllowedHeaders)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			}
		}()
		ctx := handler.NewContext(w, r)
		if p := r.Context().Value("pathParams"); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := h(ctx); err != nil {
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
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := handler.NewContext(w, r)
		if p := r.Context().Value("pathParams"); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := h(ctx); err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}
