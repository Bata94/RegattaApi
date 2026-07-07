package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/bata94/RegattaApi/internal/handler"
)

func Timeout(timeout time.Duration, message string) Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			done := make(chan error, 1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						slog.Error(fmt.Sprintf("[PANIC] %v", r))
						done <- handler.InternalError("Internal Server Error")
					}
				}()
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
