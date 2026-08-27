package webfw

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func Query(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func FormValue(r *http.Request, key string) string {
	return r.FormValue(key)
}

func Param(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func Path(r *http.Request) string {
	return r.URL.Path
}

func Method(r *http.Request) string {
	return r.Method
}

func ParseBody(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func ParseMultipartForm(r *http.Request) error {
	return r.ParseMultipartForm(32 << 20)
}

func FormFile(r *http.Request, key string) (string, []byte, error) {
	file, header, err := r.FormFile(key)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()
	content, err := io.ReadAll(file)
	return header.Filename, content, err
}

func SaveFile(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func GetUUID(r *http.Request, key string) (uuid.UUID, error) {
	param := Param(r, key)
	if param == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(param)
}

func GetUUIDFromQuery(r *http.Request, key string) (uuid.UUID, error) {
	param := Query(r, key)
	if param == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(param)
}

func PathEscape(s string) string {
	return url.PathEscape(s)
}

func JoinPath(elem ...string) string {
	return path.Join(elem...)
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

func Cookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}
