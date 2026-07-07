package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bata94/RegattaApi/internal/handler"
)

func Recovery() Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			defer func() {
				if r := recover(); r != nil {
					slog.Error(fmt.Sprintf("[PANIC] %v", r))
					c.Status(http.StatusInternalServerError)
					if err := c.SendString("Internal Server Error"); err != nil {
						slog.Error("Error sending recovery response", "err", err)
					}
				}
			}()
			return next(c)
		}
	}
}
