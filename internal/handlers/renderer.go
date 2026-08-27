package handlers

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	pdf_templates "github.com/bata94/RegattaApi/internal/templates/pdf"
)

func RenderPdf(w http.ResponseWriter, title string, comp templ.Component) error {
	comp = pdf_templates.PdfLayout(title, comp)
	w.Header().Set("Content-Type", "text/html")
	return comp.Render(context.Background(), w)
}
