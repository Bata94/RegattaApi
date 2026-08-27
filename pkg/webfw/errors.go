package webfw

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
	ui_pages "github.com/bata94/RegattaApi/internal/templates/pages"
)

type AppError = apierr.AppError

func NotFound(msg string) *AppError {
	slog.Error("NotFound", "msg", msg)
	return apierr.NotFound(msg)
}

func BadRequest(msg string) *AppError {
	slog.Error("BadRequest", "msg", msg)
	return apierr.BadRequest(msg)
}

func Unauthorized(msg string) *AppError {
	slog.Error("Unauthorized", "msg", msg)
	return apierr.Unauthorized(msg)
}

func Forbidden(msg string) *AppError {
	slog.Error("Forbidden", "msg", msg)
	return apierr.Forbidden(msg)
}

func NotAcceptable(msg string) *AppError {
	slog.Error("NotAcceptable", "msg", msg)
	return apierr.NotAcceptable(msg)
}

func InternalError(msg string) *AppError {
	slog.Error("InternalError", "msg", msg)
	return apierr.InternalError(msg)
}

func ValidationError(fieldErrors map[string]string) *AppError {
	slog.Error("ValidationError", "fieldErrors", fieldErrors)
	return apierr.ValidationError(fieldErrors)
}

func OK(msg string) *AppError {
	return apierr.OK(msg)
}

func APIError(w http.ResponseWriter, ae *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ae.StatusCode)
	if _, err := fmt.Fprintf(w, `{"code":%d,"status":%d,"message":"%s"}`, ae.Code, ae.StatusCode, ae.Message); err != nil {
		slog.Error("failed to write API error response", "error", err)
	}
}

func HandlePageError(w http.ResponseWriter, r *http.Request, err error) {
	ae, ok := err.(*AppError)
	if !ok {
		ae = InternalError("Ein interner Fehler ist aufgetreten")
	}

	path := r.URL.Path
	if strings.HasPrefix(path, "/api/") {
		APIError(w, ae)
		return
	}

	if !IsHtmxRequest(r) {
		templ.Handler(ui_pages.Error(ae.StatusCode, ae.Message)).ServeHTTP(w, r)
		return
	}

	handleHTMXError(w, r, ae)
}

func handleHTMXError(w http.ResponseWriter, r *http.Request, ae *AppError) {
	if ae.FormComp != nil && IsMutatingMethod(r) {
		w.WriteHeader(http.StatusOK)
		if err := ae.FormComp.Render(context.Background(), w); err != nil {
			slog.Warn("FormComp render error", "err", err)
		}
		writeOOBErrorToast(w, ae.Message)
		return
	}

	if IsMutatingMethod(r) {
		w.Header().Set("HX-Retarget", "#toast-container")
		w.Header().Set("HX-Swap", "beforeend")
		w.WriteHeader(http.StatusOK)
		templ.Handler(ui_components.Toast(ae.Message, ui_components.Error)).ServeHTTP(w, r)
		return
	}

	w.WriteHeader(ae.StatusCode)
	templ.Handler(ui_pages.Error(ae.StatusCode, ae.Message)).ServeHTTP(w, r)
}
