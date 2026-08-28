package webfw

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func JSON(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

func Send(w http.ResponseWriter, data []byte) error {
	_, err := w.Write(data)
	return err
}

func SendString(w http.ResponseWriter, msg string) error {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte(msg))
	return err
}

func Redirect(w http.ResponseWriter, r *http.Request, location string, code int) {
	http.Redirect(w, r, location, code)
}

func SetCookie(w http.ResponseWriter, r *http.Request, name, value string, maxAge int) {
	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func DeleteCookie(w http.ResponseWriter, r *http.Request, name string) {
	secure := r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func Status(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}
