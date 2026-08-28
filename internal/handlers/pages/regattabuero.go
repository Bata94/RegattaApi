package pages

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	regattabuero "github.com/bata94/RegattaApi/internal/templates/pages/regattabuero"
	"github.com/bata94/RegattaApi/pkg/uuid"
	"github.com/bata94/RegattaApi/pkg/webfw"
)

func InternalRegattabuero(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattabuero.Dashboard()
}

func InternalRegattabueroAbmeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	meldungen, err := crud.GetAllMeldungForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldungen"))
		return nil
	}
	return regattabuero.Abmeldung(verein, meldungen)
}

func InternalRegattabueroAbmeldungMeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	meldung, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldung"))
		return nil
	}

	if meldung.VereinUuid != verein.Uuid {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	return regattabuero.AbmeldungMeldung(verein, meldung)
}

func InternalRegattabueroUmmeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	meldungen, err := crud.GetAllMeldungForVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldungen"))
		return nil
	}
	return regattabuero.Ummeldung(verein, meldungen)
}

func InternalRegattabueroUmmeldungMeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	meldung, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldung"))
		return nil
	}

	if meldung.VereinUuid != verein.Uuid {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	athleten, err := crud.GetAllAthletenForVerein(r.Context(), verein.Uuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athleten"))
		return nil
	}

	return regattabuero.UmmeldungMeldung(verein, meldung, athleten, "", nil)
}

func InternalRegattabueroNachmeldung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	return regattabuero.Nachmeldung(verein)
}

func InternalRegattabueroNachmeldungRennen(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		slog.Error("Invalid verein UUID", "err", err)
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	rennenUuidStr := webfw.Param(r, "r_uuid")
	rennenUuid, err := uuid.Parse(rennenUuidStr)
	if err != nil {
		slog.Error("Invalid rennen UUID", "err", err)
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		slog.Error("Error loading verein", "err", err)
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	rennen, err := crud.GetRennen(r.Context(), rennenUuid)
	if err != nil {
		slog.Error("Error loading rennen", "err", err)
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading rennen"))
		return nil
	}

	athleten, err := crud.GetAllAthletenForVerein(r.Context(), verein.Uuid)
	if err != nil {
		slog.Error("Error loading athleten", "err", err)
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athleten"))
		return nil
	}

	return regattabuero.NachmeldungMeldung(verein, rennen, athleten, "", nil)
}

func InternalRegattabueroNachmeldungSuccess(w http.ResponseWriter, r *http.Request) templ.Component {
	meldungUuidStr := webfw.Param(r, "m_uuid")
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	m, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading meldung"))
		return nil
	}
	return regattabuero.NachmeldungSuccess(m)
}

func InternalRegattabueroWaageWahl(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	athleten, err := crud.GetAllAthletenForVereinWaage(r.Context(), verein.Uuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athleten"))
		return nil
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	return regattabuero.WaageWahl(verein, athleten)
}

func InternalRegattabueroWaage(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	athletUuidStr := webfw.Param(r, "a_uuid")
	athletUuid, err := uuid.Parse(athletUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	athlet, err := crud.GetAthlet(r.Context(), athletUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athlet"))
		return nil
	}

	if athlet.VereinUuid != verein.Uuid {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	return regattabuero.Waage(athlet, "", nil)
}

func InternalRegattabueroStartberechtigung(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	athleten, err := crud.GetAllAthletenForVereinMissStartber(r.Context(), verein.Uuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athleten"))
		return nil
	}

	for i := range athleten {
		athleten[i].Verein = &verein
	}

	return regattabuero.StartberechtigungWahl(verein, athleten)
}

func InternalRegattabueroStartberechtigungAthlet(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	athletUuidStr := webfw.Param(r, "a_uuid")
	athletUuid, err := uuid.Parse(athletUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	athlet, err := crud.GetAthlet(r.Context(), athletUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading athlet"))
		return nil
	}

	if athlet.VereinUuid != verein.Uuid {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}

	return regattabuero.Startberechtigung(athlet, "", nil)
}

func InternalRegattabueroNewAthlet(w http.ResponseWriter, r *http.Request) templ.Component {
	vereinUuidStr := webfw.Param(r, "v_uuid")
	vereinUuid, err := uuid.Parse(vereinUuidStr)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.NotAcceptable("Invalid UUID"))
		return nil
	}
	verein, err := crud.GetVerein(r.Context(), vereinUuid)
	if err != nil {
		webfw.HandlePageError(w, r, webfw.InternalError("Error while loading verein"))
		return nil
	}
	return regattabuero.NewAthlet(verein, "", nil)
}

func InternalRegattabueroKasse(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattabuero.Kasse()
}

func InternalRegattabueroStartnummernAusgabe(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattabuero.StartnummernAusgabe()
}

func InternalRegattabueroAenderungenObleute(w http.ResponseWriter, r *http.Request) templ.Component {
	return regattabuero.AenderungenObleute()
}
