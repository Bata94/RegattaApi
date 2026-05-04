package handlers

import (
	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/handler"
	pdf_templates "github.com/bata94/RegattaApi/internal/templates/pdf"
)

func RenderPdf(c *handler.Context, title string, comp templ.Component) error {
	comp = pdf_templates.PdfLayout(title, comp)
	return c.JSON("PDF rendering not implemented in net/http")
}
