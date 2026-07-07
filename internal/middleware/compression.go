package middleware

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bata94/RegattaApi/internal/handler"
)

var excludedExtensions = []string{
	".pdf",
	".zip",
	".tar",
	".gz",
	".jpg",
	".jpeg",
	".png",
	".gif",
	".webp",
	".mp4",
	".webm",
	".mp3",
	".wasm",
	".woff2",
	".woff",
	".ttf",
	".ico",
	".svg",
	".avi",
	".mov",
}

var excludedPaths = []string{
	"/ws",
	"/stream",
	"/events",
}

func Compression() Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(c *handler.Context) error {
			if c.Request.Header.Get("Upgrade") != "" {
				return next(c)
			}

			if c.Request.Header.Get("Content-Encoding") != "" {
				return next(c)
			}

			path := c.Path()
			for _, excluded := range excludedPaths {
				if strings.HasPrefix(path, excluded) {
					return next(c)
				}
			}

			for _, ext := range excludedExtensions {
				if strings.HasSuffix(path, ext) {
					return next(c)
				}
			}

			acceptEncoding := c.Request.Header.Get("Accept-Encoding")
			if acceptEncoding == "" {
				return next(c)
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
				return next(c)
			}

			gw := gzip.NewWriter(c.Writer)
			defer func() {
				if err := gw.Close(); err != nil {
					slog.Error("Error closing gzip writer", "err", err)
				}
			}()

			c.Writer.Header().Set("Content-Encoding", "gzip")
			c.Writer.Header().Set("Vary", "Accept-Encoding")

			originalWriter := c.Writer
			wrapped := &gzipResponseWriter{
				ResponseWriter: c.Writer,
				gw:             gw,
				written:        false,
			}
			c.Writer = wrapped

			err := next(c)

			if err := gw.Flush(); err != nil {
				slog.Error("Error flushing gzip writer", "err", err)
			}
			if err := gw.Close(); err != nil {
				slog.Error("Error closing gzip writer", "err", err)
			}

			c.Writer = originalWriter

			if err != nil {
				return err
			}

			if !wrapped.written {
				return nil
			}

			return nil
		}
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

func (g *gzipResponseWriter) HeadersWritten() bool {
	return g.written
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
