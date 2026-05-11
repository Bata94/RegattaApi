package server

import (
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/middleware"

	"github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/internal/templates/layout"
	"github.com/bata94/RegattaApi/internal/templates/pages"
)

func baseLayoutHandler(url string, pageBody templ.Component) {
	r.Handle("GET", url, pageHandler(pageBody, ui_layouts.BaseLayout))
}

func pageHandler(pageBody templ.Component, layout func(templ.Component, ui_components.NavBarConfig) templ.Component) func(http.ResponseWriter, *http.Request) {
	h := func(c *handler.Context) error {
		if c.Headers().Get("HX-Request") == "true" {
			templ.Handler(pageBody).ServeHTTP(c.Writer, c.Request)
		} else {
			templ.Handler(layout(pageBody, navBarConfig)).ServeHTTP(c.Writer, c.Request)
		}
		return nil
	}

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
			http.Error(w, err.Error(), 500)
		}
	}
}

func compHandler(url string, comp templ.Component) {
	h := func(c *handler.Context) error {
		if c.Headers().Get("HX-Request") == "true" {
			templ.Handler(comp).ServeHTTP(c.Writer, c.Request)
		} else {
			templ.Handler(ui_pages.Error(404, c.Path()+" not found")).ServeHTTP(c.Writer, c.Request)
		}
		return nil
	}

	uiStack := []middleware.Middleware{
		middleware.Recovery(),
		middleware.Compression(),
		middleware.Logging(),
		middleware.CORS(),
		middleware.RateLimit(),
		middleware.Timeout(60*time.Second, "Request timeout"),
	}

	wrapped := middleware.Chain(h, uiStack...)

	r.Handle("GET", url, func(w http.ResponseWriter, r *http.Request) {
		ctx := handler.NewContext(w, r)
		if p := r.Context().Value("pathParams"); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
}
