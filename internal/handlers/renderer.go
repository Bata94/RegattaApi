package handlers

import (
	pdf_templates "github.com/bata94/RegattaApi/internal/templates/pdf"

  "github.com/a-h/templ"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func RenderPdf(c *fiber.Ctx, title string, comp templ.Component) error {
  comp = pdf_templates.PdfLayout(title, comp)
	return adaptor.HTTPHandler(templ.Handler(comp))(c)
}
