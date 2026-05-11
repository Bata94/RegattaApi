package middleware

import (
	"log"
	"time"

	"github.com/bata94/RegattaApi/internal/handler"
)

func Logging() Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			log.Printf("%s | %s | %d | %s | %s",
				c.Method(),
				c.Path(),
				c.StatusCode,
				duration,
				c.IP())

			return err
		}
	}
}