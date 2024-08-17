package api_v1

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	pdf_templates "github.com/bata94/RegattaApi/internal/templates/pdf"
	"github.com/bata94/RegattaApi/internal/utils"
)

type AbmeldungsParams struct {
	Uuid string `json:"uuid"`
}

func StartnummernAusgabe(c *fiber.Ctx) error {
	return &api.NOT_FOUND
}

func StartnummernWechsel(c *fiber.Ctx) error {
	return &api.NOT_FOUND
}

func KasseEinzahlung(c *fiber.Ctx) error {
	return &api.NOT_FOUND
}

func KasseCreateRechnungPDF(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	v, err := crud.GetVereinMinimal(*uuid)
	if err != nil {
		return err
	}

	reNr, err := v.GetNextRechnungsnummer()
	if err != nil {
		return err
	}

	filePath, err := utils.SavePDFfromHTML(
		"buero/kasse/rechnung/"+v.Uuid.String(),
		"rechnung/"+v.Kuerzel,
		reNr,
		true,
	)
	if err != nil {
		return err
	}

	return c.SendFile(filePath, true)
}

func KasseCreateRechnungHTML(c *fiber.Ctx) error {
	uuid, err := api.GetUuidFromCtx(c)
	if err != nil {
		return err
	}

	v, err := crud.GetVereinMinimal(*uuid)
	if err != nil {
		return err
	}

	meld, err := crud.GetAllMeldungForVerein(v.Uuid)
	if err != nil {
		return err
	}

	reNr, err := v.GetNextRechnungsnummer()
	if err != nil {
		return err
	}

	pdfParams := pdf_templates.RechnungParams{
		Entries:         []pdf_templates.RechnungEntry{},
		SumPreis:        0,
		RechnungsNummer: reNr,
	}

	for _, m := range meld {
		if m.RechnungsNummer.String != "" {
			continue
		}

		pdfParams.Entries = append(pdfParams.Entries, pdf_templates.RechnungEntry{
			Tag:         string(m.Rennen.Tag),
			Startnummer: strconv.Itoa(int(m.StartNummer)),
			Rennen:      m.Rennen.Bezeichnung,
			Preis:       strconv.Itoa(int(m.Kosten)) + ",00 €",
		})
		pdfParams.SumPreis += int(m.Kosten)
		// TODO: Use a Transaction
		err := crud.SetMeldungRechnungsNummer(m.Uuid, reNr)
		if err != nil {
			log.Error(err)
		}
	}

	if len(pdfParams.Entries) == 0 {
		retErr := api.NOT_FOUND
		retErr.Msg = "Keine Meldungen gefunden, welche nicht schon abgerechnet sind!"
		return &retErr
	}

	fileName := fmt.Sprintf("Rechnung_%s", reNr)
	return handlers.RenderPdf(
		c,
		fileName,
		pdf_templates.VereinsBericht(
			v.Name,
			"Rechnung",
			false,
			pdf_templates.Rechnung(pdfParams),
		),
	)
}
