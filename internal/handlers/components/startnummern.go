package components

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/bata94/RegattaApi/internal/crud"
	apierr "github.com/bata94/RegattaApi/internal/errors"
	"github.com/bata94/RegattaApi/internal/service"
	"github.com/bata94/RegattaApi/internal/sqlc"
	regattaleitung "github.com/bata94/RegattaApi/internal/templates/pages/regattaleitung"
	"github.com/bata94/RegattaApi/pkg/webfw"
	"github.com/google/uuid"
)

func StartnummernAendernPost(w http.ResponseWriter, r *http.Request) {
	rennenUuidStr := webfw.Param(r, "r_uuid")
	meldungUuidStr := webfw.Param(r, "m_uuid")

	rennenUuid, err := uuid.Parse(rennenUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}
	meldungUuid, err := uuid.Parse(meldungUuidStr)
	if err != nil {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	m, err := crud.GetMeldung(r.Context(), meldungUuid)
	if err != nil {
		webfw.ErrorToast(w, r, "Meldung nicht gefunden")
		return
	}
	if m.RennenUuid != rennenUuid {
		webfw.ErrorToast(w, r, "Invalid UUID")
		return
	}

	fieldErrors := make(map[string]string)
	startnummer := r.FormValue("startnummer")
	if startnummer == "" {
		fieldErrors["startnummer"] = "Startnummer erforderlich"
	}
	startNummerInt, err := strconv.Atoi(startnummer)
	if err != nil {
		fieldErrors["startnummer"] = "Ungültige Startnummer"
	}

	bereich, err := crud.GetStartnummernBereich(r.Context())
	if err != nil {
		webfw.ErrorToast(w, r, "Error while loading startnummernbereich")
		return
	}

	startNummerInt32 := int32(startNummerInt)
	if !bereich.InBereich(startNummerInt32) {
		fieldErrors["startnummer"] = fmt.Sprintf("Startnummer muss zwischen %d und %d liegen", bereich.MinNummer, bereich.MaxNummer)
	}
	if startNummerInt32 > 0 && bereich.IsFehlend(startNummerInt32) {
		fieldErrors["startnummer"] = "Startnummer ist als fehlend markiert"
	}

	checkStartnummer, err := crud.GetMeldungByStartNrUndTag(r.Context(), startNummerInt, m.Rennen.Tag)
	if err != nil && !errors.As(err, &apierr.ErrNotFound) {
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}
	if checkStartnummer.Uuid != uuid.Nil {
		fieldErrors["startnummer"] = "Startnummer bereits vergeben"
	}

	if len(fieldErrors) > 0 {
		templ.Handler(regattaleitung.StartnummernAendern(m, fieldErrors)).ServeHTTP(w, r)
		return
	}

	err = crud.UpdateStartNummer(r.Context(), sqlc.UpdateStartNummerParams{
		Uuid:        m.Uuid,
		StartNummer: int32(startNummerInt),
	})
	if err != nil {
		slog.Error("UpdateStartNummer error", "err", err)
		webfw.ErrorToast(w, r, "Error while updating startnummer")
		return
	}

	m, err = crud.GetMeldung(r.Context(), m.Uuid)
	if err != nil {
		slog.Error("GetMeldung error", "err", err)
		webfw.ErrorToast(w, r, "Error while loading meldung")
		return
	}

	templ.Handler(regattaleitung.StartnummernAendern(m, fieldErrors)).ServeHTTP(w, r)
}

func StartnummernBereichPost(w http.ResponseWriter, r *http.Request) {
	fieldErrors := make(map[string]string)

	minStr := r.FormValue("min_nummer")
	maxStr := r.FormValue("max_nummer")

	minNummer, err := strconv.Atoi(minStr)
	if err != nil || minNummer < 1 {
		fieldErrors["min_nummer"] = "Kleinste Startnummer muss mindestens 1 sein"
	}
	maxNummer, err := strconv.Atoi(maxStr)
	if err != nil || maxNummer < 1 {
		fieldErrors["max_nummer"] = "Größte Startnummer muss mindestens 1 sein"
	}

	fehlendeStr := r.FormValue("fehlende_nummern")
	fehlende, err := parseFehlendeNummern(fehlendeStr)
	if err != nil {
		fieldErrors["fehlende_nummern"] = "Ungültige fehlende Startnummern"
	}

	if minNummer >= 1 && maxNummer >= 1 && maxNummer < minNummer {
		fieldErrors["max_nummer"] = "Größte Startnummer muss größer oder gleich der kleinsten sein"
	}

	if minNummer >= 1 && maxNummer >= minNummer {
		for _, n := range fehlende {
			if n < int32(minNummer) || n > int32(maxNummer) {
				fieldErrors["fehlende_nummern"] = "Fehlende Startnummern müssen im Bereich liegen"
				break
			}
		}
	}

	b := crud.StartnummernBereichFromSqlc(sqlc.StartnummernBereich{
		MinNummer:       int32(minNummer),
		MaxNummer:       int32(maxNummer),
		FehlendeNummern: fehlende,
	})

	if len(fieldErrors) > 0 {
		webfw.ErrorWithForm(w, r, regattaleitung.StartnummernBereich(b, fieldErrors), "Ungültiger Startnummernbereich")
		return
	}

	if _, err := crud.SetStartnummernBereich(r.Context(), int32(minNummer), int32(maxNummer), fehlende); err != nil {
		slog.Error("SetStartnummernBereich error", "err", err)
		webfw.ErrorToast(w, r, "Error while saving startnummernbereich")
		return
	}

	webfw.SuccessToast(w, r, "Startnummernbereich gespeichert")
}

func parseFehlendeNummern(s string) ([]int32, error) {
	if strings.TrimSpace(s) == "" {
		return []int32{}, nil
	}

	parts := strings.Split(s, ",")
	seen := make(map[int32]struct{})
	ret := make([]int32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 {
			return nil, errors.New("invalid number")
		}
		if _, ok := seen[int32(n)]; ok {
			continue
		}
		seen[int32(n)] = struct{}{}
		ret = append(ret, int32(n))
	}
	return ret, nil
}

func StartnummernVerteilenPost(w http.ResponseWriter, r *http.Request) {
	err := service.SetStartnummern(r.Context())
	if err != nil {
		webfw.ErrorToast(w, r, fmt.Sprintf("Error while setting startnummern: %s", err.Error()))
		return
	}

	webfw.SuccessToast(w, r, "Startnummern erfolgreich verteilt!")
}

func StartnummernVerteilenDelete(w http.ResponseWriter, r *http.Request) {
	err := service.ResetStartnummern(r.Context())
	if err != nil {
		webfw.ErrorToast(w, r, "Error while resetting startnummern")
		return
	}

	webfw.SuccessToast(w, r, "Startnummern erfolgreich zurückgesetzt!")
}
