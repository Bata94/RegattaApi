package middleware

import (
	"log"

	"github.com/bata94/RegattaApi/internal/handler"
)

func Recovery() Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PANIC] %v", r)
					c.Status(500)
					c.SendString("Internal Server Error")
				}
			}()
			return next(c)
		}
	}
}