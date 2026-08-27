package middleware

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

var excludedExtensions = []string{
	".pdf", ".zip", ".tar", ".gz", ".jpg", ".jpeg", ".png", ".gif", ".webp",
	".mp4", ".webm", ".mp3", ".wasm", ".woff2", ".woff", ".ttf", ".ico",
	".svg", ".avi", ".mov",
}

var excludedPaths = []string{
	"/ws", "/stream", "/events", "/comp/image",
}

func Compression() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Upgrade") != "" {
				next.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("Content-Encoding") != "" {
				next.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path
			for _, excluded := range excludedPaths {
				if strings.HasPrefix(path, excluded) {
					next.ServeHTTP(w, r)
					return
				}
			}

			for _, ext := range excludedExtensions {
				if strings.HasSuffix(path, ext) {
					next.ServeHTTP(w, r)
					return
				}
			}

			acceptEncoding := r.Header.Get("Accept-Encoding")
			if acceptEncoding == "" {
				next.ServeHTTP(w, r)
				return
			}

			encoding := ""
			encodings := strings.Split(acceptEncoding, ",")
			for _, enc := range encodings {
				enc = strings.TrimSpace(enc)
				if enc == "gzip" {
					encoding = "gzip"
					break
				}
			}

			if encoding == "" {
				next.ServeHTTP(w, r)
				return
			}

			gw := gzip.NewWriter(w)
			gwClosed := false
			defer func() {
				if !gwClosed {
					if err := gw.Close(); err != nil {
						slog.Error("Error closing gzip writer", "err", err)
					}
				}
			}()

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			wrapped := &gzipResponseWriter{
				ResponseWriter: w,
				gw:             gw,
				written:        false,
			}

			gwClosed = true
			next.ServeHTTP(wrapped, r)

			if err := gw.Close(); err != nil {
				slog.Error("Error closing gzip writer", "err", err)
			}

			if err := gw.Flush(); err != nil {
				slog.Error("Error flushing gzip writer", "err", err)
			}
		})
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gw      *gzip.Writer
	written bool
}

func (g *gzipResponseWriter) Write(p []byte) (int, error) {
	g.written = true
	return g.gw.Write(p)
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	if g.written {
		return
	}
	g.written = true
	g.ResponseWriter.WriteHeader(statusCode)
}

func (g *gzipResponseWriter) Flush() {
	if err := g.gw.Flush(); err != nil {
		slog.Error("Error flushing gzip writer", "err", err)
	}
	if flusher, ok := g.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (g *gzipResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	g.written = true
	return io.Copy(g.gw, r)
}

func (g *gzipResponseWriter) HeadersWritten() bool {
	return g.written
}
