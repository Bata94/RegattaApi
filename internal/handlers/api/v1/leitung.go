package api_v1

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/utils"
)

func GetPdfFooter(c *handler.Context) error {
	return c.JSON("footer placeholder")
}

func GetMeldeergebnisList(c *handler.Context) error {
	files, err := utils.GetFilenames("meldeergebnis")
	if err != nil {
		return err
	}
	return c.JSON(files)
}

func GetMeldeergebnisFilename(c *handler.Context) error {
	filename := c.Param("filename")
	filePath := filepath.Join("./files", "meldeergebnis", filename)
	http.ServeFile(c.Writer, c.Request, filePath)
	return nil
}

func GetMeldeergebnisHtml(c *handler.Context) error {
	return c.JSON("Meldeergebnis not implemented in net/http")
}

func GenerateMeldeergebnis(c *handler.Context) error {
	fp, err := utils.SavePDFfromHTML(
		"leitung/meldeergebnis",
		"meldeergebnis",
		fmt.Sprintf("Meldeergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		return err
	}
	http.ServeFile(c.Writer, c.Request, fp)
	return nil
}

func DrvMeldungUpload(c *handler.Context) error {
	return c.JSON("DRV upload not implemented in net/http")
}

func GenerateErgebnisHtml(c *handler.Context) error {
	return c.JSON("Ergebnis not implemented in net/http")
}

func GenerateErgebnis(c *handler.Context) error {
	fp, err := utils.SavePDFfromHTML(
		"leitung/ergebnis",
		"ergebnis",
		fmt.Sprintf("ergebnis_%s", time.Now().Format("2006-01-02_15-04-05")),
		true,
	)
	if err != nil {
		return err
	}
	http.ServeFile(c.Writer, c.Request, fp)
	return nil
}

func SetzungsLosung(c *handler.Context) error {
	return c.JSON("Setzung not implemented in net/http")
}

func ResetSetzung(c *handler.Context) error {
	return c.JSON("Reset not implemented in net/http")
}

func SetStartnummern(c *handler.Context) error {
	return c.JSON("Startnummern not implemented in net/http")
}

func SetZeitplan(c *handler.Context) error {
	return c.JSON("Zeitplan not implemented in net/http")
}
