package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
)

type headersWrittenChecker interface {
	HeadersWritten() bool
}

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					slog.Error(fmt.Sprintf("[PANIC] %v", rec))
					if hwc, ok := w.(headersWrittenChecker); ok && hwc.HeadersWritten() {
						slog.Warn("panic after headers written, cannot respond safely")
						return
					}
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := w.Write([]byte("Internal Server Error")); err != nil {
						slog.Error("failed to write recovery response", "error", err)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
