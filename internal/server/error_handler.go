package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/templates/components"
	"github.com/bata94/RegattaApi/internal/templates/pages"
)

func toAppError(err error) *handler.AppError {
	switch e := err.(type) {
	case *handler.AppError:
		return e
	case *handler.Error:
		slog.Warn("deprecated handler.Error returned", "status", e.StatusCode, "msg", e.Message)
		return &handler.AppError{
			Code:       e.StatusCode,
			StatusCode: e.StatusCode,
			Message:    e.Message,
		}
	default:
		slog.Error("unhandled error type", "error", err)
		return handler.InternalError("Ein interner Fehler ist aufgetreten")
	}
}

func handleAppError(c *handler.Context, err error) {
	if ht, ok := c.Writer.(handler.HeaderTracker); ok && ht.HeadersWritten() {
		return
	}

	ae := toAppError(err)

	if ae.Code == 200 {
		writeSuccessToast(c, ae)
		return
	}

	if strings.HasPrefix(c.Path(), "/api/") {
		writeAPIError(c, ae)
		return
	}

	if !c.IsHtmxRequest() {
		writePageError(c, ae)
		return
	}

	writeHTMXError(c, ae)
}

func writeSuccessToast(c *handler.Context, ae *handler.AppError) {
	c.Writer.Header().Set("HX-Retarget", "#toast-container")
	c.Writer.Header().Set("HX-Swap", "beforeend")
	c.Writer.WriteHeader(http.StatusOK)
	templ.Handler(ui_components.Toast(ae.Message, ui_components.Success)).ServeHTTP(c.Writer, c.Request)
}

func writeAPIError(c *handler.Context, ae *handler.AppError) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(ae.StatusCode)
	json.NewEncoder(c.Writer).Encode(map[string]any{
		"code":    ae.Code,
		"status":  ae.StatusCode,
		"message": ae.Message,
	})
}

func writePageError(c *handler.Context, ae *handler.AppError) {
	c.Writer.WriteHeader(ae.StatusCode)
	templ.Handler(ui_pages.Error(ae.StatusCode, ae.Message)).ServeHTTP(c.Writer, c.Request)
}

func writeHTMXError(c *handler.Context, ae *handler.AppError) {
	method := c.Method()

	if ae.FormComp != nil && (method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE") {
		c.Writer.WriteHeader(http.StatusOK)
		ae.FormComp.Render(context.Background(), c.Writer)
		writeOOBToast(c, ae.Message, ui_components.Error)
		return
	}

	if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		c.Writer.Header().Set("HX-Retarget", "#toast-container")
		c.Writer.Header().Set("HX-Swap", "beforeend")
		c.Writer.WriteHeader(http.StatusOK)
		templ.Handler(ui_components.Toast(ae.Message, ui_components.Error)).ServeHTTP(c.Writer, c.Request)
		return
	}

	c.Writer.WriteHeader(ae.StatusCode)
	templ.Handler(ui_pages.Error(ae.StatusCode, ae.Message)).ServeHTTP(c.Writer, c.Request)
}

func writeOOBToast(c *handler.Context, msg string, color ui_components.InputColor) {
	alertColor := "alert-ghost"
	switch color {
	case ui_components.Primary:
		alertColor = "alert-primary"
	case ui_components.Secondary:
		alertColor = "alert-secondary"
	case ui_components.Success:
		alertColor = "alert-success"
	case ui_components.Error:
		alertColor = "alert-error"
	case ui_components.Warning:
		alertColor = "alert-warning"
	case ui_components.Info:
		alertColor = "alert-info"
	case ui_components.Mrgblau:
		alertColor = "alert-mrgblau"
	}
	oobHTML := fmt.Sprintf(
		`<div hx-swap-oob="beforeend:#toast-container"><div class="alert %s flex flex-row justify-between items-center gap-2"><span>%s</span><button class="btn btn-sm btn-circle btn-ghost" onclick="this.parentElement.remove()">✕</button></div></div>`,
		alertColor, msg,
	)
	fmt.Fprint(c.Writer, oobHTML)
}
