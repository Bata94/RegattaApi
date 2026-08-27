package components

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/config"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/internal/utils"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func PdfMeldeergebnisPost(w http.ResponseWriter, r *http.Request) {
	fileName := fmt.Sprintf("Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05"))
	_, err := utils.SavePDFfromHTML(
		"leitung/meldeergebnis",
		"meldeergebnis",
		fileName,
		true,
	)
	if err != nil {
		if rmErr := os.Remove(fmt.Sprintf("%smeldeergebnis/%s", config.C.Paths.FilesDir, fileName)); rmErr != nil {
			slog.Error("Error removing failed PDF file", "err", rmErr)
		}
		webfw.ErrorToast(w, r, fmt.Sprintf("Fehler während PDF Erstellung: %s", err.Error()))
		return
	}

	templ.Handler(regattaleitung.PdfMeldeergebnis(true)).ServeHTTP(w, r)
}
