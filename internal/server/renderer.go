package server

import (
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/middleware"

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

		if p := r.Context().Value("pathParams"); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			handleAppError(ctx, err)
		}
	})
}
