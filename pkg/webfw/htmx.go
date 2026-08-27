package webfw

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	ui_components "github.com/bata94/RegattaApi/internal/templates/components"
)

func IsHtmxRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func IsMutatingMethod(r *http.Request) bool {
	switch r.Method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	}
	return false
}

func SetRedirect(w http.ResponseWriter, location string) {
	w.Header().Set("HX-Redirect", location)
}

func SetPushUrl(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Push-Url", url)
}

func SetCacheControl(w http.ResponseWriter, value string) {
	w.Header().Set("Cache-Control", value)
}

func SuccessToast(w http.ResponseWriter, r *http.Request, msg string) {
	w.Header().Set("HX-Retarget", "#toast-container")
	w.Header().Set("HX-Swap", "beforeend")
	w.WriteHeader(http.StatusOK)
	templ.Handler(ui_components.Toast(msg, ui_components.Success)).ServeHTTP(w, r)
}

func ErrorToast(w http.ResponseWriter, r *http.Request, msg string) {
	w.Header().Set("HX-Retarget", "#toast-container")
	w.Header().Set("HX-Swap", "beforeend")
	w.WriteHeader(http.StatusOK)
	templ.Handler(ui_components.Toast(msg, ui_components.Error)).ServeHTTP(w, r)
}

func SuccessWithForm(w http.ResponseWriter, r *http.Request, form templ.Component, msg string) {
	w.WriteHeader(http.StatusOK)
	if form != nil {
		if err := form.Render(context.Background(), w); err != nil {
			slog.Warn("form render error", "err", err)
		}
	}
	w.Header().Set("HX-Retarget", "#toast-container")
	w.Header().Set("HX-Swap", "beforeend")
	templ.Handler(ui_components.Toast(msg, ui_components.Success)).ServeHTTP(w, r)
}

func ErrorWithForm(w http.ResponseWriter, r *http.Request, form templ.Component, msg string) {
	w.WriteHeader(http.StatusOK)
	if form != nil {
		if err := form.Render(context.Background(), w); err != nil {
			slog.Warn("form render error", "err", err)
		}
	}
	writeOOBErrorToast(w, msg)
}

func writeOOBErrorToast(w http.ResponseWriter, msg string) {
	oobHTML := fmt.Sprintf(
		`<div hx-swap-oob="beforeend:#toast-container"><div class="alert alert-error flex flex-row justify-between items-center gap-2"><span>%s</span><button class="btn btn-sm btn-circle btn-ghost" onclick="this.parentElement.remove()">✕</button></div></div>`,
		msg,
	)
	if _, err := fmt.Fprint(w, oobHTML); err != nil {
		slog.Warn("writeOOBErrorToast error", "err", err)
	}
}
