package server

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/middleware"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
)

type router struct {
	handlers map[string]map[string]http.HandlerFunc
}

func newRouter() *router {
	return &router{
		handlers: make(map[string]map[string]http.HandlerFunc),
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

	h := methodHandlers[req.URL.Path]
	var params map[string]string
	if h == nil {
		h, params = r.matchWildcard(req.Method, req.URL.Path)
	}
	if h == nil {
		http.NotFound(w, req)
		return
	}

	if params != nil {
		ctx := context.WithValue(req.Context(), handler.CtxKeyPathParams, params)
		req = req.WithContext(ctx)
	}

	h(w, req)
}

func (r *router) matchWildcard(method, path string) (http.HandlerFunc, map[string]string) {
	methodHandlers := r.handlers[method]
	patterns := make([]string, 0, len(methodHandlers))
	for pattern := range methodHandlers {
		patterns = append(patterns, pattern)
	}
	sort.Slice(patterns, func(i, j int) bool {
		return len(patterns[i]) > len(patterns[j])
	})
	for _, pattern := range patterns {
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
		if p := r.Context().Value(handler.CtxKeyPathParams); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			handleAppError(ctx, err)
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

		if p := r.Context().Value(handler.CtxKeyPathParams); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			handleAppError(ctx, err)
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
		if p := r.Context().Value(handler.CtxKeyPathParams); p != nil {
			ctx.SetPathParams(p.(map[string]string))
		}
		if err := wrapped(ctx); err != nil {
			handleAppError(ctx, err)
		}
	}
}
