package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			slog.Info(fmt.Sprintf("%s | %s | %d | %s | %s",
				r.Method,
				r.URL.Path,
				wrapped.statusCode,
				duration,
				IP(r),
			))
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (s *statusWriter) WriteHeader(code int) {
	s.statusCode = code
	s.ResponseWriter.WriteHeader(code)
}

func IP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if idx := strings.Index(forwarded, ","); idx > 0 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}
	return r.RemoteAddr
}
