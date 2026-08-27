package pages

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func InternalRegattaleitung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Dashboard()
}

func InternalRegattaleitungDrvUpload(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.DrvFileUpload("")
}

func InternalRegattaleitungSetzung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Setzung()
}

func InternalRegattaleitungSetzungLosung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.SetzungLosung()
}

func InternalRegattaleitungSetzungAenderung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.SetzungAenderung()
}

func InternalRegattaleitungSetzungAenderungRennen(w http.ResponseWriter, r *http.Request) templ.Component {
	paramStr := webfw.Param(r, "param")
	slog.Debug("Param", "value", paramStr)

	slog.Debug("Param is a Rennen UUID")
	rUuid, err := uuid.Parse(paramStr)
	if err != nil {
		slog.Error("Error", "err", err)
		webfw.HandlePageError(w, r, webfw.NotFound("Rennen nicht gefunden"))
		return nil
	}
	return regattaleitung.SetzungAenderungRennen(rUuid)
}

func InternalRegattaleitungPausen(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Pausen()
}

func InternalRegattaleitungZeitplan(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Zeitplan("", nil)
}

func InternalRegattaleitungStartnummern(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Startnummern()
}

func InternalRegattaleitungStartnummernVerteilen(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.StartnummernVerteilen()
}

func InternalRegattaleitungStartnummernBereich(w http.ResponseWriter, r *http.Request) templ.Component {
	b, err := crud.GetStartnummernBereich(r.Context())
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error loading startnummernbereich: "+err.Error()))
		return nil
	}
	return regattaleitung.StartnummernBereich(b, nil)
}

func InternalRegattaleitungStartnummernAendernRennenWahl(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.StartnummernAendernRennenWahl()
}

func InternalRegattaleitungStartnummernAendernMeldungsWahl(w http.ResponseWriter, r *http.Request) templ.Component {
	rUuidStr := webfw.Param(r, "r_uuid")
	rUuid, err := uuid.Parse(rUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	rennen, err := crud.GetRennen(r.Context(), rUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotFound("Rennen nicht gefunden"))
		return nil
	}

	for i := range rennen.Meldungen {
		rennen.Meldungen[i].Rennen = &rennen
	}

	return regattaleitung.StartnummernAendernMeldungsWahl(rennen)
}

func InternalRegattaleitungStartnummernAendern(w http.ResponseWriter, r *http.Request) templ.Component {
	mUuidStr := webfw.Param(r, "m_uuid")

	mUuid, err := uuid.Parse(mUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	m, err := crud.GetMeldung(r.Context(), mUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotFound("Meldung nicht gefunden"))
		return nil
	}

	return regattaleitung.StartnummernAendern(m, nil)
}

func InternalRegattaleitungPdfMeldeergebnis(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.PdfMeldeergebnis(false)
}

func InternalRegattaleitungVereinsverwaltung(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Vereinsverwaltung()
}

func InternalRegattaleitungObleute(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.Obleute()
}

func InternalRegattaleitungEmail(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattaleitung.EmailCompose("", "", "", nil, false, nil, nil)
}
