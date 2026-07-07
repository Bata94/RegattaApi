package middleware

import "github.com/bata94/RegattaApi/internal/handler"

type Middleware func(handler.Handler) handler.Handler

func Chain(h handler.Handler, middlewares ...Middleware) handler.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
