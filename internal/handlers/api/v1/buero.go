package api_v1

import (
	"log"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handler"
	"github.com/bata94/RegattaApi/internal/utils"
)

type AbmeldungsParams struct {
	Uuid string `json:"uuid"`
}

func StartnummernAusgabe(c *handler.Context) error {
	return &handler.Error{StatusCode: 404, Message: "Not found"}
}

func StartnummernWechsel(c *handler.Context) error {
	return &handler.Error{StatusCode: 404, Message: "Not found"}
}

func KasseEinzahlung(c *handler.Context) error {
	return &handler.Error{StatusCode: 404, Message: "Not found"}
}

func KasseCreateRechnungPDF(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	v, err := crud.GetVereinMinimal(uuid)
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
	log.Println("Generated:", filePath)

	toMail := []string{}
	obleute, err := crud.GetAllObmannForVerein(v.Uuid)
	if err != nil {
		return err
	}

	for _, o := range obleute {
		if o.Name.Valid {
			toMail = append(toMail, o.Email.String)
		}
	}

	err = utils.SendMail(utils.SendMailParams{
		To:      toMail,
		CC:      []string{},
		Subject: "MRG Regatta 24 - Rechnung " + reNr,
		Body:    "Anbei finden Sie eine neu erstellte Rechnung für Ihren Verein.",
		Files:   []string{filePath},
	})

	if err != nil {
		return err
	}

	return c.JSON("success")
}

func KasseCreateRechnungAllVereine(c *handler.Context) error {
	vereine, err := crud.GetAllVerein()
	if err != nil {
		return err
	}

	errLs := []error{}
	for _, v := range vereine {
		reNr, err := v.GetNextRechnungsnummer()
		if err != nil {
			errLs = append(errLs, err)
			continue
		}

		filePath, err := utils.SavePDFfromHTML(
			"buero/kasse/rechnung/"+v.Uuid.String(),
			"rechnung/"+v.Kuerzel,
			reNr,
			true,
		)
		if err != nil {
			errLs = append(errLs, err)
			continue
		}
		log.Println("Generated:", filePath)
	}

	if len(errLs) > 0 {
		return c.JSON(errLs)
	}
	return c.JSON("success")
}

func KasseCreateRechnungHTML(c *handler.Context) error {
	uuid, err := c.GetUUID("uuid")
	if err != nil {
		return &handler.Error{StatusCode: 400, Message: err.Error()}
	}

	v, err := crud.GetVereinMinimal(uuid)
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

	type RechnungEntry struct {
		Tag         string
		Startnummer string
		Rennen      string
		Preis       string
	}

	entries := []RechnungEntry{}
	sumPreis := 0

	for _, m := range meld {
		if m.RechnungsNummer.String != "" {
			continue
		}

		entries = append(entries, RechnungEntry{
			Tag:         string(m.Rennen.Tag),
			Startnummer: string(rune(int(m.StartNummer) + '0')),
			Rennen:      m.Rennen.Bezeichnung,
			Preis:       string(rune(int(m.Kosten)+'0')) + ",00 €",
		})
		sumPreis += int(m.Kosten)

		err := crud.SetMeldungRechnungsNummer(m.Uuid, reNr)
		if err != nil {
			log.Println(err)
		}
	}

	if len(entries) == 0 {
		return &handler.Error{StatusCode: 404, Message: "Keine Meldungen gefunden!"}
	}

	err = crud.CreateRechnung(reNr, v.Uuid, sumPreis)
	if err != nil {
		return err
	}

	return c.JSON("Rechnung generated")
}
