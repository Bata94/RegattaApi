package middleware

import (
	"os"

	"github.com/bata94/RegattaApi/internal/handler"
)

func CORS() Middleware {
	origins := os.Getenv("CORS_ALLOWED_ORIGINS")
	methods := os.Getenv("CORS_ALLOWED_METHODS")
	headers := os.Getenv("CORS_ALLOWED_HEADERS")

	if origins == "" {
		origins = "*"
	}
	if methods == "" {
		methods = "GET, POST, PUT, DELETE, OPTIONS"
	}
	if headers == "" {
		headers = "Content-Type, Authorization"
	}

	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origins)
			c.Writer.Header().Set("Access-Control-Allow-Methods", methods)
			c.Writer.Header().Set("Access-Control-Allow-Headers", headers)

			if c.Method() == "OPTIONS" {
				c.Status(200)
				return nil
			}
			return next(c)
		}
	}
}