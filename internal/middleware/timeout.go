package middleware

import (
	"net/http"
	"time"

	"github.com/bata94/RegattaApi/internal/handler"
)

func Timeout(timeout time.Duration, message string) Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			done := make(chan error, 1)

			go func() {
				done <- next(c)
			}()

			select {
			case err := <-done:
				return err
			case <-time.After(timeout):
				c.Status(http.StatusServiceUnavailable)
				return handler.InternalError(message)
			}
		}
	}
}