package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				w.WriteHeader(http.StatusServiceUnavailable)
				if _, err := w.Write([]byte("Request timeout")); err != nil {
					slog.Error("failed to write timeout response", "error", err)
				}
			}
		})
	}
}
