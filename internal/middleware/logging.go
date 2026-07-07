package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bata94/RegattaApi/internal/handler"
)

func Logging() Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			slog.Info(fmt.Sprintf("%s | %s | %d | %s | %s",
				c.Method(),
				c.Path(),
				c.StatusCode(),
				duration,
				c.IP()),
			)

			return err
		}
	}
}
